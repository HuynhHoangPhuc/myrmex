package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSPublisher publishes Subject domain events to NATS JetStream.
type NATSPublisher struct {
	js nats.JetStreamContext
}

// NewNATSPublisher creates a publisher connected to the given NATS URL.
func NewNATSPublisher(url string) (*NATSPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS %s: %w", url, err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("create JetStream context: %w", err)
	}
	// Ensure SUBJECT_EVENTS stream exists
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     "SUBJECT_EVENTS",
		Subjects: []string{"subject.>"},
	}); err != nil {
		_ = err
	}
	return &NATSPublisher{js: js}, nil
}

// Publish serialises payload to JSON and publishes to the given NATS subject.
func (p *NATSPublisher) Publish(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if _, err := p.js.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}
