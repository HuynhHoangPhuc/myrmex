// Package redis provides a Redis Pub/Sub relay for WebSocket push notifications.
// Replaces the NATS core (non-JetStream) pub/sub used by the WSBroker.
// Works in both local (Redis container) and cloud (GCP Memorystore) environments.
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// PushMessage is a received push event from Redis Pub/Sub.
type PushMessage struct {
	Channel string
	Payload []byte
}

// PushPublisher publishes WebSocket push events to Redis Pub/Sub.
type PushPublisher struct {
	rdb *redis.Client
}

// NewPushPublisher wraps a Redis client for publishing push events.
// rdb may be nil; all calls become no-ops.
func NewPushPublisher(rdb *redis.Client) *PushPublisher {
	return &PushPublisher{rdb: rdb}
}

// Publish sends payload to the given Redis channel.
func (p *PushPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	if p.rdb == nil {
		return nil
	}
	return p.rdb.Publish(ctx, channel, payload).Err()
}

// PushSubscriber subscribes to Redis Pub/Sub channels for push relay.
type PushSubscriber struct {
	rdb *redis.Client
}

// NewPushSubscriber wraps a Redis client for pattern-based push subscriptions.
func NewPushSubscriber(rdb *redis.Client) *PushSubscriber {
	return &PushSubscriber{rdb: rdb}
}

// PSubscribe subscribes to all channels matching the given glob pattern
// (e.g., "notification.push.*") and returns a channel of PushMessage.
// The returned channel is closed when ctx is cancelled.
func (s *PushSubscriber) PSubscribe(ctx context.Context, pattern string) (<-chan PushMessage, error) {
	if s.rdb == nil {
		return nil, fmt.Errorf("redis not available")
	}
	ps := s.rdb.PSubscribe(ctx, pattern)
	// Validate the subscription was established.
	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		return nil, fmt.Errorf("redis psubscribe %s: %w", pattern, err)
	}

	out := make(chan PushMessage, 64)
	go func() {
		defer close(out)
		defer ps.Close() //nolint:errcheck
		ch := ps.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				out <- PushMessage{
					Channel: msg.Channel,
					Payload: []byte(msg.Payload),
				}
			}
		}
	}()

	return out, nil
}
