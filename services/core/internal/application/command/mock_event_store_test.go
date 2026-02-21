package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/pkg/eventstore"
)

type mockEventStore struct {
	appendErr error
	loadEvents []eventstore.Event
	loadErr    error
}

func (m *mockEventStore) Append(_ context.Context, _ uuid.UUID, _ int, _ []eventstore.Event) error {
	return m.appendErr
}

func (m *mockEventStore) Load(_ context.Context, _ uuid.UUID, _ int) ([]eventstore.Event, error) {
	return m.loadEvents, m.loadErr
}

func (m *mockEventStore) SaveSnapshot(_ context.Context, _ eventstore.Snapshot) error {
	return nil
}

func (m *mockEventStore) LoadSnapshot(_ context.Context, _ uuid.UUID) (*eventstore.Snapshot, error) {
	return nil, nil
}
