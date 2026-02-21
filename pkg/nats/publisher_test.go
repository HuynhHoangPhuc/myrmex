package nats

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
)

type mockPublisherJetStream struct {
	err error
}

func (m *mockPublisherJetStream) Publish(_ context.Context, _ string, _ []byte, _ ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &jetstream.PubAck{}, nil
}

func TestPublisher_Publish_Success(t *testing.T) {
	js := &mockPublisherJetStream{}
	publisher := NewPublisher(js)
	if err := publisher.Publish(context.Background(), "core.user.created", []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestPublisher_Publish_Error(t *testing.T) {
	js := &mockPublisherJetStream{err: errors.New("publish failed")}
	publisher := NewPublisher(js)
	err := publisher.Publish(context.Background(), "core.user.created", []byte(`{"ok":true}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
