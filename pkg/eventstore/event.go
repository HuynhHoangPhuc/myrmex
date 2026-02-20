package eventstore

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents a single domain event persisted in the event store.
type Event struct {
	ID            int64           `json:"id"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	EventType     string          `json:"event_type"`
	Version       int             `json:"version"`
	Data          json.RawMessage `json:"data"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
}

// Snapshot stores a materialized aggregate state at a specific version.
type Snapshot struct {
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	Version       int             `json:"version"`
	Data          json.RawMessage `json:"data"`
	CreatedAt     time.Time       `json:"created_at"`
}
