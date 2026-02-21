package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// PublisherJetStream is the minimal JetStream capability required by Publisher.
type PublisherJetStream interface {
	Publish(ctx context.Context, subj string, data []byte, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error)
}

// Publisher wraps JetStream for publishing domain events.
type Publisher struct {
	js PublisherJetStream
}

// NewPublisher creates a new event publisher.
func NewPublisher(js PublisherJetStream) *Publisher {
	return &Publisher{js: js}
}

// Publish sends data to a NATS JetStream subject.
func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}
