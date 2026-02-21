package eventstore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvent_FieldAssignment(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC()
	payload := json.RawMessage(`{"email":"a@b.com"}`)
	metadata := json.RawMessage(`{"source":"test"}`)

	event := Event{
		ID:            12,
		AggregateID:   id,
		AggregateType: "user",
		EventType:     "user.created",
		Version:       1,
		Data:          payload,
		Metadata:      metadata,
		CreatedAt:     createdAt,
	}

	if event.ID != 12 {
		t.Fatalf("expected ID 12, got %d", event.ID)
	}
	if event.AggregateID != id {
		t.Fatalf("expected AggregateID %s, got %s", id, event.AggregateID)
	}
	if event.AggregateType != "user" {
		t.Fatalf("expected AggregateType user, got %s", event.AggregateType)
	}
	if event.EventType != "user.created" {
		t.Fatalf("expected EventType user.created, got %s", event.EventType)
	}
	if event.Version != 1 {
		t.Fatalf("expected Version 1, got %d", event.Version)
	}
	if string(event.Data) != string(payload) {
		t.Fatalf("expected Data %s, got %s", payload, event.Data)
	}
	if string(event.Metadata) != string(metadata) {
		t.Fatalf("expected Metadata %s, got %s", metadata, event.Metadata)
	}
	if !event.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %s, got %s", createdAt, event.CreatedAt)
	}
}

func TestSnapshot_FieldAssignment(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC()
	payload := json.RawMessage(`{"state":"active"}`)

	snap := Snapshot{
		AggregateID:   id,
		AggregateType: "user",
		Version:       5,
		Data:          payload,
		CreatedAt:     createdAt,
	}

	if snap.AggregateID != id {
		t.Fatalf("expected AggregateID %s, got %s", id, snap.AggregateID)
	}
	if snap.AggregateType != "user" {
		t.Fatalf("expected AggregateType user, got %s", snap.AggregateType)
	}
	if snap.Version != 5 {
		t.Fatalf("expected Version 5, got %d", snap.Version)
	}
	if string(snap.Data) != string(payload) {
		t.Fatalf("expected Data %s, got %s", payload, snap.Data)
	}
	if !snap.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %s, got %s", createdAt, snap.CreatedAt)
	}
}
