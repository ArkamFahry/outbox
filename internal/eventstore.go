package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ArkamFahry/outbox/internal/models"
	"github.com/nats-io/nats.go/jetstream"
)

type IEventStore interface {
	Publish(event *models.Event) error
}

type EventStore struct {
	eventStore jetstream.JetStream
	config     *Config
}

func NewPublisher(event jetstream.JetStream, config *Config) *EventStore {
	return &EventStore{
		eventStore: event,
		config:     config,
	}
}

func (es *EventStore) CreateStream() error {
	_, err := es.eventStore.CreateStream(context.Background(), jetstream.StreamConfig{
		Name: es.config.ServiceName,
		Subjects: []string{
			fmt.Sprintf("%s.>", es.config.ServiceName),
		},
	})
	if err != nil {
		if errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
			return nil
		}
		return err
	}

	return nil
}

func (es *EventStore) Publish(event *models.Event) error {
	subject := event.Subject(es.config.ServiceName)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marhal event to json")
	}

	_, err = es.eventStore.PublishAsync(subject, eventBytes)
	if err != nil {
		return err
	}

	return nil
}
