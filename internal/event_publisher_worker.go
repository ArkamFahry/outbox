package internal

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"sync"
)

type EventPublisherWorker struct {
	database   IDatabase
	eventStore IEventStore
	logger     *zap.Logger
}

func NewEventPublisherWorker(database IDatabase, eventStore IEventStore, logger *zap.Logger) *EventPublisherWorker {
	return &EventPublisherWorker{
		database:   database,
		eventStore: eventStore,
		logger:     logger,
	}
}

func (w *EventPublisherWorker) Work(ctx context.Context) error {
	var publishedEventIds []string
	var failedEventIds []string
	var mutex sync.Mutex

	err := w.database.WithTransaction(ctx, func(tx pgx.Tx) error {
		events, err := w.database.GetEvents(ctx, tx)
		if err != nil {
			if errors.Is(err, ErrEventNotFound) {
				return nil
			}
			w.logger.Error("failed to get events", zap.Error(err))
			return err
		}

		var waitGroup sync.WaitGroup

		for _, event := range events {
			waitGroup.Add(1)
			go func(event *Event) {
				defer waitGroup.Done()

				err = w.eventStore.Publish(event)
				if err != nil {
					w.logger.Error("failed to publish event", zap.Error(err))
					mutex.Lock()
					failedEventIds = append(failedEventIds, event.Id)
					mutex.Unlock()
					return
				}
				mutex.Lock()
				publishedEventIds = append(publishedEventIds, event.Id)
				mutex.Unlock()
			}(event)
		}

		waitGroup.Wait()

		err = w.database.UpdateEventsStatusPublished(ctx, tx, publishedEventIds)
		if err != nil {
			w.logger.Error(
				"failed to update events status to published",
				zap.Error(err),
				zap.Strings("published_event_ids", publishedEventIds),
			)
			return err
		}

		err = w.database.UpdateEventsStatusFailed(ctx, tx, failedEventIds)
		if err != nil {
			w.logger.Error(
				"failed to update events status to failed",
				zap.Error(err),
				zap.Strings("failed_event_ids", failedEventIds),
			)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
