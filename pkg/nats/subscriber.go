package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// MessageHandler processes a JetStream message.
type MessageHandler func(msg jetstream.Msg)

// Subscribe creates a durable consumer on a stream and starts consuming messages.
// Returns the consumer for lifecycle management.
func Subscribe(ctx context.Context, js jetstream.JetStream, stream, consumer, subject string, handler MessageHandler) (jetstream.ConsumeContext, error) {
	// Ensure stream exists
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     stream,
		Subjects: []string{subject},
	})
	if err != nil {
		return nil, fmt.Errorf("create stream %s: %w", stream, err)
	}

	// Create durable consumer
	cons, err := js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:       consumer,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer %s: %w", consumer, err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		handler(msg)
	})
	if err != nil {
		return nil, fmt.Errorf("consume %s: %w", consumer, err)
	}

	return cc, nil
}
