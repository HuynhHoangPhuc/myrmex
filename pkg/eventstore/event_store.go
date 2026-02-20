package eventstore

import (
	"context"

	"github.com/google/uuid"
)

// EventStore defines the interface for appending and loading domain events.
// Implementations must guarantee optimistic concurrency via expectedVersion.
type EventStore interface {
	// Append persists events for an aggregate. Fails if current version != expectedVersion.
	Append(ctx context.Context, aggregateID uuid.UUID, expectedVersion int, events []Event) error
	// Load returns events for an aggregate starting from fromVersion (inclusive).
	Load(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error)
	// SaveSnapshot persists an aggregate snapshot (upsert).
	SaveSnapshot(ctx context.Context, snap Snapshot) error
	// LoadSnapshot returns the latest snapshot for an aggregate.
	LoadSnapshot(ctx context.Context, aggregateID uuid.UUID) (*Snapshot, error)
}
