package internal

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go/jetstream"
)

type IEventStore interface {
	Publish(event *Event) error
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

func (es *EventStore) Publish(event *Event) error {
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
