package models

import (
	"fmt"
)

type Event struct {
	Id            string         `json:"id" db:"id"`
	AggregateType string         `json:"aggregate_type" db:"aggregate_type"`
	EventType     string         `json:"event_type" db:"event_type"`
	Content       map[string]any `json:"content" db:"content"`
}

func (e *Event) Subject(serviceName string) string {
	return fmt.Sprintf("%s.%s.%s", serviceName, e.AggregateType, e.EventType)
}
