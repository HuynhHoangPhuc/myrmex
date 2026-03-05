package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	msgredis "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/redis"
)

// wsPushEvent is the JSON payload published to "notification.push.{userID}"
// and forwarded by the core WSBroker to connected WebSocket clients.
type wsPushEvent struct {
	Type        string    `json:"type"` // always "notification"
	ID          string    `json:"id"`
	NotifType   string    `json:"notif_type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UnreadCount int64     `json:"unread_count"`
}

// RedisPushPublisher publishes in-app push events to Redis Pub/Sub for WS relay.
// Replaces the previous NATS core publish approach.
type RedisPushPublisher struct {
	pub *msgredis.PushPublisher
	log *zap.Logger
}

// NewRedisPushPublisher creates a RedisPushPublisher.
// pub may wrap a nil Redis client (no-op when Redis is unavailable).
func NewRedisPushPublisher(pub *msgredis.PushPublisher, log *zap.Logger) *RedisPushPublisher {
	return &RedisPushPublisher{pub: pub, log: log}
}

// PublishPush sends a push event to "notification.push.{userID}" for WS relay.
func (p *RedisPushPublisher) PublishPush(userID, notifID, notifType, title, body string, createdAt time.Time, unreadCount int64) {
	event := wsPushEvent{
		Type:        "notification",
		ID:          notifID,
		NotifType:   notifType,
		Title:       title,
		Body:        body,
		CreatedAt:   createdAt,
		UnreadCount: unreadCount,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		p.log.Error("marshal push event", zap.Error(err))
		return
	}

	channel := fmt.Sprintf("notification.push.%s", userID)
	if err := p.pub.Publish(context.Background(), channel, payload); err != nil {
		p.log.Warn("redis publish push event failed", zap.String("user_id", userID), zap.Error(err))
	}
}
