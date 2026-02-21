//go:build integration

package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	integrationSchema = "core"
)

func setupPostgresEventStore(t *testing.T) (*PostgresEventStore, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("pgx pool: %v", err)
	}

	if err := applyEventStoreSchema(ctx, pool); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("apply schema: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}

	return NewPostgresEventStore(pool, integrationSchema), cleanup
}

func applyEventStoreSchema(ctx context.Context, pool *pgxpool.Pool) error {
	ddl := []string{
		"CREATE SCHEMA IF NOT EXISTS core;",
		`CREATE TABLE core.events (
			id BIGSERIAL PRIMARY KEY,
			aggregate_id UUID NOT NULL,
			aggregate_type VARCHAR(100) NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			version INT NOT NULL,
			data JSONB NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(aggregate_id, version)
		);`,
		"CREATE INDEX idx_core_events_aggregate ON core.events(aggregate_id, version);",
		"CREATE INDEX idx_core_events_type ON core.events(event_type);",
		`CREATE TABLE core.snapshots (
			aggregate_id UUID PRIMARY KEY,
			aggregate_type VARCHAR(100) NOT NULL,
			version INT NOT NULL,
			data JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, stmt := range ddl {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

func TestPostgresEventStoreAppendAndLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, cleanup := setupPostgresEventStore(t)
	defer cleanup()

	ctx := context.Background()
	aggregateID := uuid.New()

	events := []Event{
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "created",
			Data:          json.RawMessage(`{"v":1}`),
			Metadata:      json.RawMessage(`{"source":"integration"}`),
		},
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "updated",
			Data:          json.RawMessage(`{"v":2}`),
			Metadata:      json.RawMessage(`{"source":"integration"}`),
		},
	}

	if err := store.Append(ctx, aggregateID, 0, events); err != nil {
		t.Fatalf("append events: %v", err)
	}

	loaded, err := store.Load(ctx, aggregateID, 0)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}
	if loaded[0].Version != 1 || loaded[1].Version != 2 {
		t.Fatalf("unexpected versions: %d, %d", loaded[0].Version, loaded[1].Version)
	}
	if loaded[0].EventType != "created" || loaded[1].EventType != "updated" {
		t.Fatalf("unexpected event types: %s, %s", loaded[0].EventType, loaded[1].EventType)
	}
}

func TestPostgresEventStoreLoadFromVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, cleanup := setupPostgresEventStore(t)
	defer cleanup()

	ctx := context.Background()
	aggregateID := uuid.New()

	events := []Event{
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "created",
			Data:          json.RawMessage(`{"v":1}`),
			Metadata:      json.RawMessage(`{}`),
		},
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "updated",
			Data:          json.RawMessage(`{"v":2}`),
			Metadata:      json.RawMessage(`{}`),
		},
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "updated",
			Data:          json.RawMessage(`{"v":3}`),
			Metadata:      json.RawMessage(`{}`),
		},
	}

	if err := store.Append(ctx, aggregateID, 0, events); err != nil {
		t.Fatalf("append events: %v", err)
	}

	loaded, err := store.Load(ctx, aggregateID, 2)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}
	if loaded[0].Version != 2 || loaded[1].Version != 3 {
		t.Fatalf("unexpected versions: %d, %d", loaded[0].Version, loaded[1].Version)
	}
}

func TestPostgresEventStoreOptimisticConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, cleanup := setupPostgresEventStore(t)
	defer cleanup()

	ctx := context.Background()
	aggregateID := uuid.New()

	if err := store.Append(ctx, aggregateID, 0, []Event{
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "created",
			Data:          json.RawMessage(`{"v":1}`),
			Metadata:      json.RawMessage(`{}`),
		},
	}); err != nil {
		t.Fatalf("append initial event: %v", err)
	}

	err := store.Append(ctx, aggregateID, 0, []Event{
		{
			AggregateID:   aggregateID,
			AggregateType: "test",
			EventType:     "stale",
			Data:          json.RawMessage(`{"v":2}`),
			Metadata:      json.RawMessage(`{}`),
		},
	})
	if err == nil {
		t.Fatal("expected optimistic concurrency error")
	}
}

func TestPostgresEventStoreSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, cleanup := setupPostgresEventStore(t)
	defer cleanup()

	ctx := context.Background()
	aggregateID := uuid.New()

	snap, err := store.LoadSnapshot(ctx, aggregateID)
	if err != nil {
		t.Fatalf("load empty snapshot: %v", err)
	}
	if snap != nil {
		t.Fatalf("expected nil snapshot, got version %d", snap.Version)
	}

	if err := store.SaveSnapshot(ctx, Snapshot{
		AggregateID:   aggregateID,
		AggregateType: "test",
		Version:       5,
		Data:          json.RawMessage(`{"state":"v5"}`),
	}); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	snap, err = store.LoadSnapshot(ctx, aggregateID)
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	if snap.Version != 5 {
		t.Fatalf("snapshot version: got %d want 5", snap.Version)
	}
	if !jsonEqual(t, snap.Data, json.RawMessage(`{"state":"v5"}`)) {
		t.Fatalf("snapshot data mismatch: %s", string(snap.Data))
	}

	if err := store.SaveSnapshot(ctx, Snapshot{
		AggregateID:   aggregateID,
		AggregateType: "test",
		Version:       10,
		Data:          json.RawMessage(`{"state":"v10"}`),
	}); err != nil {
		t.Fatalf("save snapshot upsert: %v", err)
	}

	snap, err = store.LoadSnapshot(ctx, aggregateID)
	if err != nil {
		t.Fatalf("load snapshot upsert: %v", err)
	}
	if snap.Version != 10 {
		t.Fatalf("snapshot version: got %d want 10", snap.Version)
	}
	if !jsonEqual(t, snap.Data, json.RawMessage(`{"state":"v10"}`)) {
		t.Fatalf("snapshot data mismatch: %s", string(snap.Data))
	}
}

func jsonEqual(t *testing.T, left json.RawMessage, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		t.Fatalf("unmarshal left json: %v", err)
	}
	if err := json.Unmarshal(right, &rightValue); err != nil {
		t.Fatalf("unmarshal right json: %v", err)
	}
	return string(mustMarshalJSON(t, leftValue)) == string(mustMarshalJSON(t, rightValue))
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}
