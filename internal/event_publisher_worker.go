package internal

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
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

func (w *EventPublisherWorker) Work(ctx context.Context) {
	for i := 0; i < 10; i++ {
		go func() {
			for {
				var publishedEventIds []string
				var failedEventIds []string

				err := w.database.WithTransaction(ctx, func(tx pgx.Tx) error {
					events, err := w.database.GetEvents(ctx, tx)
					if err != nil {
						if errors.Is(err, ErrEventNotFound) {
							return nil
						}
						w.logger.Error("failed to get events", zap.Error(err))
						return err
					}

					for _, event := range events {
						err = w.eventStore.Publish(event)
						if err != nil {
							w.logger.Error("failed to publish event", zap.Error(err))
							failedEventIds = append(failedEventIds, event.Id)
						}
						publishedEventIds = append(publishedEventIds, event.Id)
					}

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
					w.logger.Error("failed to publish events", zap.Error(err))
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}
	select {}
}
