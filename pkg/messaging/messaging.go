// Package messaging provides a backend-agnostic pub/sub abstraction.
// Supports NATS JetStream (local dev) and Cloud Pub/Sub (GCP production).
// Select backend via MESSAGING_BACKEND env var: "nats" (default) or "pubsub".
package messaging

import "context"

// Message represents a received event from the messaging backend.
type Message struct {
	Subject string
	Data    []byte
	ack     func() error
	nack    func() error
}

// NewMessage constructs a Message. For internal use by adapters.
func NewMessage(subject string, data []byte, ack, nack func() error) *Message {
	return &Message{Subject: subject, Data: data, ack: ack, nack: nack}
}

// Ack acknowledges successful processing. No-op if backend is fire-and-forget.
func (m *Message) Ack() error {
	if m.ack != nil {
		return m.ack()
	}
	return nil
}

// Nack signals failed processing; triggers redelivery when supported.
func (m *Message) Nack() error {
	if m.nack != nil {
		return m.nack()
	}
	return nil
}

// Publisher publishes domain events to the configured messaging backend.
type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Close() error
}

// StreamConfig describes a stream (NATS JetStream) or topic group (Pub/Sub) to pre-create.
type StreamConfig struct {
	Name     string
	Subjects []string        // NATS: subject filters; Pub/Sub: used for topic naming
	MaxAge   int64           // seconds; 0 = unlimited
}

// Consumer subscribes to domain events from the configured messaging backend.
type Consumer interface {
	// EnsureStream pre-creates a stream (NATS) or topics (Pub/Sub) before subscribing.
	// Safe to call multiple times; idempotent.
	EnsureStream(ctx context.Context, cfg StreamConfig) error

	// Subscribe registers a durable handler for the given subject.
	// Handler should return nil to ack, non-nil error to nack/retry.
	// Runs in a background goroutine until Close() is called.
	Subscribe(ctx context.Context, durableName, subject string, handler func(*Message) error) error

	Close() error
}

// NoopPublisher discards all published events. Used when messaging is unavailable.
type NoopPublisher struct{}

func (n *NoopPublisher) Publish(_ context.Context, _ string, _ []byte) error { return nil }
func (n *NoopPublisher) Close() error                                          { return nil }
