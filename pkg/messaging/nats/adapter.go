// Package nats implements the messaging.Publisher and messaging.Consumer interfaces
// using NATS JetStream v2 API. Used for local development via Docker Compose.
package nats

import (
	"context"
	"fmt"
	"strings"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	pkgnats "github.com/HuynhHoangPhuc/myrmex/pkg/nats"
	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
)

// --- Publisher ---

// Publisher wraps a NATS JetStream connection to implement messaging.Publisher.
type Publisher struct {
	js   jetstream.JetStream
	conn *natsgo.Conn
}

// NewPublisher connects to NATS and returns a JetStream publisher.
func NewPublisher(url string) (*Publisher, error) {
	nc, js, err := pkgnats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats publisher connect: %w", err)
	}
	return &Publisher{js: js, conn: nc}, nil
}

// NewPublisherFromJS wraps an already-established NATS connection and JetStream context.
// conn ownership is NOT transferred; caller must close it separately.
func NewPublisherFromJS(nc *natsgo.Conn, js jetstream.JetStream) *Publisher {
	return &Publisher{js: js, conn: nc}
}

// EnsureStream creates or updates a JetStream stream. Call before publishing
// to a subject that belongs to a stream.
func (p *Publisher) EnsureStream(ctx context.Context, cfg messaging.StreamConfig) error {
	scfg := jetstream.StreamConfig{
		Name:     cfg.Name,
		Subjects: cfg.Subjects,
	}
	if cfg.MaxAge > 0 {
		scfg.MaxAge = time.Duration(cfg.MaxAge) * time.Second
	}
	_, err := p.js.CreateOrUpdateStream(ctx, scfg)
	return err
}

// Publish sends data to a JetStream subject.
func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("nats publish %s: %w", subject, err)
	}
	return nil
}

// Close drains and closes the NATS connection.
func (p *Publisher) Close() error {
	p.conn.Close()
	return nil
}

// --- Consumer ---

// Consumer wraps a NATS JetStream connection to implement messaging.Consumer.
type Consumer struct {
	js      jetstream.JetStream
	conn    *natsgo.Conn
	cancels []func()
}

// NewConsumer connects to NATS and returns a JetStream consumer.
func NewConsumer(url string) (*Consumer, error) {
	nc, js, err := pkgnats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats consumer connect: %w", err)
	}
	return &Consumer{js: js, conn: nc}, nil
}

// NewConsumerFromJS wraps an already-established NATS connection and JetStream context.
// conn ownership is NOT transferred; caller must close it separately.
func NewConsumerFromJS(nc *natsgo.Conn, js jetstream.JetStream) *Consumer {
	return &Consumer{js: js, conn: nc}
}

// EnsureStream creates or updates a JetStream stream before subscribing.
func (c *Consumer) EnsureStream(ctx context.Context, cfg messaging.StreamConfig) error {
	scfg := jetstream.StreamConfig{
		Name:     cfg.Name,
		Subjects: cfg.Subjects,
	}
	if cfg.MaxAge > 0 {
		scfg.MaxAge = time.Duration(cfg.MaxAge) * time.Second
	}
	_, err := c.js.CreateOrUpdateStream(ctx, scfg)
	return err
}

// Subscribe creates a durable JetStream consumer for the given subject and
// starts a background goroutine that calls handler for each message.
// Infers the stream name from the subject prefix via inferStream.
// Handler return value: nil → ack, non-nil → nack (redelivery).
func (c *Consumer) Subscribe(ctx context.Context, durableName, subject string, handler func(*messaging.Message) error) error {
	stream, ok := inferStream(subject)
	if !ok {
		return fmt.Errorf("nats consumer: no stream mapping for subject %q", subject)
	}

	// Ensure the stream exists with a minimal config (caller should call EnsureStream
	// with full config if custom retention is needed).
	if _, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     stream,
		Subjects: []string{streamWildcard(stream)},
	}); err != nil {
		return fmt.Errorf("ensure stream %s: %w", stream, err)
	}

	cfg := jetstream.ConsumerConfig{
		Durable:   durableName,
		AckPolicy: jetstream.AckExplicitPolicy,
	}
	if subject != streamWildcard(stream) {
		cfg.FilterSubject = subject
	}

	cons, err := c.js.CreateOrUpdateConsumer(ctx, stream, cfg)
	if err != nil {
		return fmt.Errorf("create consumer %s: %w", durableName, err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		m := messaging.NewMessage(msg.Subject(), msg.Data(),
			func() error { return msg.Ack() },
			func() error { return msg.Nak() },
		)
		if err := handler(m); err != nil {
			_ = msg.Nak()
		} else {
			_ = msg.Ack()
		}
	})
	if err != nil {
		return fmt.Errorf("start consuming %s: %w", durableName, err)
	}

	c.cancels = append(c.cancels, cc.Stop)
	return nil
}

// Close stops all consumers and closes the NATS connection.
func (c *Consumer) Close() error {
	for _, stop := range c.cancels {
		stop()
	}
	c.conn.Close()
	return nil
}

// inferStream maps a subject (or wildcard) to its JetStream stream name.
func inferStream(subject string) (string, bool) {
	switch {
	case strings.HasPrefix(subject, "hr."):
		return "HR_EVENTS", true
	case strings.HasPrefix(subject, "subject."):
		return "SUBJECT_EVENTS", true
	case strings.HasPrefix(subject, "student."):
		return "STUDENT_EVENTS", true
	case strings.HasPrefix(subject, "timetable."):
		return "TIMETABLE", true
	case strings.HasPrefix(subject, "AUDIT."):
		return "AUDIT_EVENTS", true
	case strings.HasPrefix(subject, "core."):
		return "CORE_EVENTS", true
	case strings.HasPrefix(subject, "notification."):
		return "NOTIFICATION_EVENTS", true
	}
	return "", false
}

// streamWildcard returns the catch-all subject for a stream.
func streamWildcard(stream string) string {
	switch stream {
	case "HR_EVENTS":
		return "hr.>"
	case "SUBJECT_EVENTS":
		return "subject.>"
	case "STUDENT_EVENTS":
		return "student.>"
	case "TIMETABLE":
		return "timetable.>"
	case "AUDIT_EVENTS":
		return "AUDIT.>"
	case "CORE_EVENTS":
		return "core.>"
	case "NOTIFICATION_EVENTS":
		return "notification.>"
	}
	return stream + ".>"
}
