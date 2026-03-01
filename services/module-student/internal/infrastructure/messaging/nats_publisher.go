package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSPublisher publishes student domain events to NATS JetStream.
type NATSPublisher struct {
	nc *nats.Conn
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
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     "STUDENT_EVENTS",
		Subjects: []string{"student.>"},
	}); err != nil {
		_ = err
	}
	return &NATSPublisher{nc: nc, js: js}, nil
}

// Publish serializes payload to JSON and publishes it.
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

func (p *NATSPublisher) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if p.nc == nil {
		return nil, fmt.Errorf("nats connection is not available")
	}
	sub, err := p.nc.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}
	return sub, nil
}
