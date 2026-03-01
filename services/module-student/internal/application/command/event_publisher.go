package command

import "context"

// EventPublisher publishes domain events. Implementations: NATSPublisher, noopPublisher.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

type noopPublisher struct{}

func NewNoopPublisher() EventPublisher { return &noopPublisher{} }

func (n *noopPublisher) Publish(context.Context, string, any) error { return nil }
