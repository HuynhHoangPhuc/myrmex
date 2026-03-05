package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
)

// EventPublisher publishes Timetable domain events via the configured messaging backend.
// Implements command.EventPublisher.
type EventPublisher struct {
	pub messaging.Publisher
}

// NewEventPublisher wraps a messaging.Publisher for Timetable domain events.
func NewEventPublisher(pub messaging.Publisher) *EventPublisher {
	return &EventPublisher{pub: pub}
}

// Publish marshals payload to JSON and publishes it to the given subject.
func (p *EventPublisher) Publish(ctx context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.pub.Publish(ctx, subject, data)
}
