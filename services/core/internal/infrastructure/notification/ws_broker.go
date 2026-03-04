package notification

import (
	"strings"
	"sync"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const wsBrokerWriteTimeout = 5 * time.Second

// WSBroker manages active WebSocket connections per user and relays NATS push events.
// module-notification publishes to NOTIFICATION.push.{userID}; the broker forwards
// the raw JSON payload to all connected WebSocket clients for that user.
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

// StartNATSRelay subscribes to NOTIFICATION.push.* and forwards each message
// to connected WebSocket clients for the target user. nc may be nil (no-op).
func (b *WSBroker) StartNATSRelay(nc *natsgo.Conn) {
	if nc == nil {
		b.log.Warn("NATS not available, WS push relay disabled")
		return
	}

	_, err := nc.Subscribe("NOTIFICATION.push.*", func(msg *natsgo.Msg) {
		// Subject format: NOTIFICATION.push.{userID}
		parts := strings.Split(msg.Subject, ".")
		if len(parts) < 3 {
			return
		}
		userID := parts[2]
		b.send(userID, msg.Data)
	})
	if err != nil {
		b.log.Error("failed to subscribe to NOTIFICATION.push.*", zap.Error(err))
		return
	}
	b.log.Info("NATS → WS relay started on NOTIFICATION.push.*")
}
