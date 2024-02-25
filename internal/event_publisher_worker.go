package internal

import (
	"context"
	"errors"
	"github.com/ArkamFahry/outbox/internal/database"
	"github.com/ArkamFahry/outbox/internal/eventstore"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type EventPublisherWorker struct {
	config     *Config
	database   database.IDatabase
	eventStore eventstore.IEventStore
	logger     *zap.Logger
}

func NewEventPublisherWorker(config *Config, database database.IDatabase, eventStore eventstore.IEventStore, logger *zap.Logger) *EventPublisherWorker {
	return &EventPublisherWorker{
		config:     config,
		database:   database,
		eventStore: eventStore,
		logger:     logger,
	}
}

func (w *EventPublisherWorker) Work(ctx context.Context) {
	for i := 0; i <= w.config.WorkerCount; i++ {
		go func() {
			for {
				var publishedEventIds []string
				var failedEventIds []string

				err := w.database.WithTransaction(ctx, func(tx pgx.Tx) error {
					events, err := w.database.GetEvents(ctx, tx)
					if err != nil {
						if errors.Is(err, database.ErrEventNotFound) {
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
