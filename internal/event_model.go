package internal

import (
	"fmt"
	"time"
)

type Event struct {
	Id            string     `json:"id" db:"id"`
	Version       int        `json:"version" db:"version"`
	AggregateType string     `json:"aggregate_type" db:"aggregate_type"`
	EventType     string     `json:"event_type" db:"event_type"`
	Content       []byte     `json:"content" db:"content"`
	Status        string     `json:"status" db:"status"`
	PublishedAt   *time.Time `json:"published_at" db:"published_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

func (e *Event) Subject(serviceName string) string {
	return fmt.Sprintf("%s.%s.%s", serviceName, e.AggregateType, e.EventType)
}
