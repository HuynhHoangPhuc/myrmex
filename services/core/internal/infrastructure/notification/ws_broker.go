package notification

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	msgredis "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/redis"
)

const wsBrokerWriteTimeout = 5 * time.Second

// WSBroker manages active WebSocket connections per user and relays Redis push events.
// module-notification publishes to "notification.push.{userID}" via Redis Pub/Sub;
// the broker forwards the raw JSON payload to all connected WebSocket clients.
type WSBroker struct {
	mu          sync.RWMutex
	connections map[string][]*websocket.Conn // userID → open connections
	log         *zap.Logger
}

func NewWSBroker(log *zap.Logger) *WSBroker {
	return &WSBroker{
		connections: make(map[string][]*websocket.Conn),
		log:         log,
	}
}

// Register adds a WebSocket connection for a user.
func (b *WSBroker) Register(userID string, conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.connections[userID] = append(b.connections[userID], conn)
}

// Unregister removes a WebSocket connection for a user.
func (b *WSBroker) Unregister(userID string, conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	conns := b.connections[userID]
	filtered := conns[:0]
	for _, c := range conns {
		if c != conn {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		delete(b.connections, userID)
	} else {
		b.connections[userID] = filtered
	}
}

// send delivers a raw JSON payload to all open connections for a user, cleaning up dead ones.
func (b *WSBroker) send(userID string, payload []byte) {
	b.mu.RLock()
	conns := b.connections[userID]
	b.mu.RUnlock()

	if len(conns) == 0 {
		return
	}

	var dead []*websocket.Conn
	for _, conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(wsBrokerWriteTimeout)) //nolint:errcheck
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			b.log.Debug("ws push failed, marking dead", zap.String("user_id", userID), zap.Error(err))
			dead = append(dead, conn)
		}
	}
	for _, c := range dead {
		b.Unregister(userID, c)
	}
}

// StartRedisRelay subscribes to "notification.push.*" via Redis Pub/Sub and
// forwards each message to connected WebSocket clients for the target user.
// sub may be nil (no-op when Redis is unavailable).
func (b *WSBroker) StartRedisRelay(ctx context.Context, sub *msgredis.PushSubscriber) {
	if sub == nil {
		b.log.Warn("Redis not available, WS push relay disabled")
		return
	}

	ch, err := sub.PSubscribe(ctx, "notification.push.*")
	if err != nil {
		b.log.Error("failed to subscribe to notification.push.*", zap.Error(err))
		return
	}

	go func() {
		for msg := range ch {
			// Channel format: notification.push.{userID}
			parts := strings.Split(msg.Channel, ".")
			if len(parts) < 3 {
				continue
			}
			userID := parts[2]
			b.send(userID, msg.Payload)
		}
	}()

	b.log.Info("Redis → WS relay started on notification.push.*")
}
