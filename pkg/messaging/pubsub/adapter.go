// Package pubsub implements the messaging.Publisher and messaging.Consumer interfaces
// using Google Cloud Pub/Sub. Used for GCP production (Cloud Run deployment).
// Subjects map to topics via inferTopic(); filtering uses the "subject" message attribute.
package pubsub

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
)

// --- Publisher ---

// Publisher wraps a Cloud Pub/Sub client to implement messaging.Publisher.
type Publisher struct {
	client *pubsub.Client
	topics map[string]*pubsub.Topic // cached topic handles keyed by topic ID
	log    *zap.Logger
}

// NewPublisher creates a Pub/Sub publisher for the given GCP project.
// Uses Application Default Credentials (ADC) when running on GCP.
func NewPublisher(ctx context.Context, projectID string, log *zap.Logger) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub publisher connect project=%s: %w", projectID, err)
	}
	return &Publisher{client: client, topics: make(map[string]*pubsub.Topic), log: log}, nil
}

// EnsureStream creates Pub/Sub topics derived from cfg.Name (lowercased, dashes).
// For Pub/Sub, one topic per stream is pre-created in GCP; this call is a no-op
// when the topic already exists.
func (p *Publisher) EnsureStream(ctx context.Context, cfg messaging.StreamConfig) error {
	topicID := streamToTopic(cfg.Name)
	if _, err := p.client.CreateTopic(ctx, topicID); err != nil {
		// AlreadyExists is fine
		if !isAlreadyExists(err) {
			return fmt.Errorf("ensure pubsub topic %s: %w", topicID, err)
		}
	}
	return nil
}

// Publish serializes data to the Pub/Sub topic mapped from subject.
// Sets the "subject" message attribute for subscriber-side filtering.
func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	topicID := inferTopic(subject)
	topic, err := p.getOrCreateTopic(ctx, topicID)
	if err != nil {
		return err
	}
	result := topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: map[string]string{"subject": subject},
	})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("pubsub publish %s: %w", subject, err)
	}
	return nil
}

// Close stops all cached topics and closes the client.
func (p *Publisher) Close() error {
	for _, t := range p.topics {
		t.Stop()
	}
	return p.client.Close()
}

func (p *Publisher) getOrCreateTopic(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	if t, ok := p.topics[topicID]; ok {
		return t, nil
	}
	t := p.client.Topic(topicID)
	exists, err := t.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("check topic %s: %w", topicID, err)
	}
	if !exists {
		t, err = p.client.CreateTopic(ctx, topicID)
		if err != nil && !isAlreadyExists(err) {
			return nil, fmt.Errorf("create topic %s: %w", topicID, err)
		}
		t = p.client.Topic(topicID)
	}
	p.topics[topicID] = t
	return t, nil
}

// --- Consumer ---

// Consumer wraps a Cloud Pub/Sub client to implement messaging.Consumer.
type Consumer struct {
	client        *pubsub.Client
	projectID     string
	cancelFuncs   []context.CancelFunc
	log           *zap.Logger
}

// NewConsumer creates a Pub/Sub consumer for the given GCP project.
func NewConsumer(ctx context.Context, projectID string, log *zap.Logger) (*Consumer, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub consumer connect project=%s: %w", projectID, err)
	}
	return &Consumer{client: client, projectID: projectID, log: log}, nil
}

// EnsureStream creates the Pub/Sub topic for the given stream config.
func (c *Consumer) EnsureStream(ctx context.Context, cfg messaging.StreamConfig) error {
	topicID := streamToTopic(cfg.Name)
	if _, err := c.client.CreateTopic(ctx, topicID); err != nil {
		if !isAlreadyExists(err) {
			return fmt.Errorf("ensure pubsub topic %s: %w", topicID, err)
		}
	}
	return nil
}

// Subscribe creates (or reuses) a Pub/Sub pull subscription for the given subject.
// The subscription ID is derived from durableName. Subject filtering uses the
// "subject" message attribute set by Publisher.Publish.
func (c *Consumer) Subscribe(ctx context.Context, durableName, subject string, handler func(*messaging.Message) error) error {
	topicID := inferTopic(subject)
	topic := c.client.Topic(topicID)

	// Subscription ID: sanitize durableName for Pub/Sub naming rules.
	subID := sanitizeSubID(durableName)
	sub := c.client.Subscription(subID)

	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check subscription %s: %w", subID, err)
	}
	if !exists {
		cfg := pubsub.SubscriptionConfig{
			Topic: topic,
		}
		// Add subject attribute filter if this is not a wildcard subscription.
		if !strings.HasSuffix(subject, ".>") && !strings.HasSuffix(subject, "*") {
			cfg.Filter = fmt.Sprintf(`attributes.subject = "%s"`, subject)
		}
		if _, err := c.client.CreateSubscription(ctx, subID, cfg); err != nil && !isAlreadyExists(err) {
			return fmt.Errorf("create subscription %s: %w", subID, err)
		}
		sub = c.client.Subscription(subID)
	}

	subCtx, cancel := context.WithCancel(ctx)
	c.cancelFuncs = append(c.cancelFuncs, cancel)

	go func() {
		if err := sub.Receive(subCtx, func(ctx context.Context, msg *pubsub.Message) {
			m := messaging.NewMessage(
				msg.Attributes["subject"],
				msg.Data,
				func() error { msg.Ack(); return nil },
				func() error { msg.Nack(); return nil },
			)
			if err := handler(m); err != nil {
				c.log.Warn("pubsub handler error, nacking", zap.String("sub", subID), zap.Error(err))
				msg.Nack()
			} else {
				msg.Ack()
			}
		}); err != nil && subCtx.Err() == nil {
			c.log.Error("pubsub receive error", zap.String("sub", subID), zap.Error(err))
		}
	}()

	return nil
}

// Close cancels all active subscriptions and closes the client.
func (c *Consumer) Close() error {
	for _, cancel := range c.cancelFuncs {
		cancel()
	}
	return c.client.Close()
}

// inferTopic maps a NATS-style subject to a Pub/Sub topic ID.
func inferTopic(subject string) string {
	switch {
	case strings.HasPrefix(subject, "hr."):
		return "hr-events"
	case strings.HasPrefix(subject, "subject."):
		return "subject-events"
	case strings.HasPrefix(subject, "student."):
		return "student-events"
	case strings.HasPrefix(subject, "timetable."):
		return "timetable-events"
	case strings.HasPrefix(subject, "AUDIT."):
		return "audit-events"
	case strings.HasPrefix(subject, "core."):
		return "core-events"
	case strings.HasPrefix(subject, "notification."):
		return "notification-events"
	}
	return strings.ReplaceAll(strings.ToLower(subject), ".", "-")
}

// streamToTopic converts a NATS stream name (e.g., "HR_EVENTS") to a Pub/Sub topic ID.
func streamToTopic(stream string) string {
	return strings.ReplaceAll(strings.ToLower(stream), "_", "-")
}

// sanitizeSubID replaces characters not allowed in Pub/Sub subscription IDs.
func sanitizeSubID(name string) string {
	return strings.NewReplacer(".", "-", "_", "-").Replace(name)
}

// isAlreadyExists returns true for Pub/Sub "already exists" API errors.
func isAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "AlreadyExists")
}
