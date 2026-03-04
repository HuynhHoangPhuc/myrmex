package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// wsPushEvent is the JSON payload published to NOTIFICATION.push.{userID}
// and forwarded by the core WSBroker to connected WebSocket clients.
type wsPushEvent struct {
	Type        string    `json:"type"`    // always "notification"
	ID          string    `json:"id"`
	NotifType   string    `json:"notif_type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UnreadCount int64     `json:"unread_count"`
}

// NATSPublisher publishes in-app push events to core NATS for WS relay.
type NATSPublisher struct {
	nc  *natsgo.Conn
	log *zap.Logger
}

// NewNATSPublisher creates a NATSPublisher. nc may be nil (no-op when NATS unavailable).
func NewNATSPublisher(nc *natsgo.Conn, log *zap.Logger) *NATSPublisher {
	return &NATSPublisher{nc: nc, log: log}
}

// PublishPush sends a push event to NOTIFICATION.push.{userID} for WS relay.
func (p *NATSPublisher) PublishPush(userID, notifID, notifType, title, body string, createdAt time.Time, unreadCount int64) {
	if p.nc == nil {
		return
	}

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

	subject := fmt.Sprintf("NOTIFICATION.push.%s", userID)
	if err := p.nc.Publish(subject, payload); err != nil {
		p.log.Warn("publish push event failed", zap.String("user_id", userID), zap.Error(err))
	}
}
