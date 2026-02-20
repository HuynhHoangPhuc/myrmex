package eventstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresEventStore implements EventStore backed by PostgreSQL.
type PostgresEventStore struct {
	pool   *pgxpool.Pool
	schema string // schema-per-module isolation
}

// NewPostgresEventStore creates a new PostgreSQL-backed event store.
func NewPostgresEventStore(pool *pgxpool.Pool, schema string) *PostgresEventStore {
	return &PostgresEventStore{pool: pool, schema: schema}
}

func (s *PostgresEventStore) Append(ctx context.Context, aggregateID uuid.UUID, expectedVersion int, events []Event) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Optimistic concurrency: check current version
	var currentVersion int
	err = tx.QueryRow(ctx,
		fmt.Sprintf(`SELECT COALESCE(MAX(version), 0) FROM %s.events WHERE aggregate_id = $1`, s.schema),
		aggregateID,
	).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("check version: %w", err)
	}
	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion)
	}

	for i, e := range events {
		version := expectedVersion + i + 1
		_, err = tx.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s.events (aggregate_id, aggregate_type, event_type, version, data, metadata)
				VALUES ($1, $2, $3, $4, $5, $6)`, s.schema),
			aggregateID, e.AggregateType, e.EventType, version, e.Data, e.Metadata,
		)
		if err != nil {
			return fmt.Errorf("insert event %d: %w", version, err)
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresEventStore) Load(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error) {
	rows, err := s.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, aggregate_id, aggregate_type, event_type, version, data, metadata, created_at
			FROM %s.events WHERE aggregate_id = $1 AND version >= $2 ORDER BY version`, s.schema),
		aggregateID, fromVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.AggregateType, &e.EventType,
			&e.Version, &e.Data, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *PostgresEventStore) SaveSnapshot(ctx context.Context, snap Snapshot) error {
	_, err := s.pool.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s.snapshots (aggregate_id, aggregate_type, version, data)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (aggregate_id) DO UPDATE SET version = $3, data = $4, created_at = NOW()`, s.schema),
		snap.AggregateID, snap.AggregateType, snap.Version, snap.Data,
	)
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func (s *PostgresEventStore) LoadSnapshot(ctx context.Context, aggregateID uuid.UUID) (*Snapshot, error) {
	var snap Snapshot
	var data []byte
	err := s.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT aggregate_id, aggregate_type, version, data, created_at
			FROM %s.snapshots WHERE aggregate_id = $1`, s.schema),
		aggregateID,
	).Scan(&snap.AggregateID, &snap.AggregateType, &snap.Version, &data, &snap.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("load snapshot: %w", err)
	}
	snap.Data = json.RawMessage(data)
	return &snap, nil
}
