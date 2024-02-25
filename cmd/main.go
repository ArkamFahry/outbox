package main

import (
	"context"
	"github.com/ArkamFahry/outbox/internal/database"
	"github.com/ArkamFahry/outbox/internal/eventstore"

	"github.com/ArkamFahry/outbox/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

func main() {
	config := internal.NewConfig()

	logger := internal.NewLogger(config)

	natsClient, err := nats.Connect(config.NatsUrl)
	if err != nil {
		logger.Fatal("error connecting to nats", zap.Error(err))
	}

	jetStreamsClient, err := jetstream.New(natsClient)
	if err != nil {
		logger.Fatal("error connecting to jetstream", zap.Error(err))
	}

	pgxPoolConfig, err := pgxpool.ParseConfig(config.PostgresUrl)
	if err != nil {
		logger.Fatal("error parsing postgres url",
			zap.Error(err),
		)
	}

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		logger.Fatal("error connecting to postgres",
			zap.Error(err),
		)
	}

	eventDatabase := database.NewDatabase(pgxPool, config)

	eventStore := eventstore.NewEventStore(jetStreamsClient, config)

	publisher := internal.NewEventPublisherWorker(config, eventDatabase, eventStore, logger)

	publisher.Work(context.Background())
}
