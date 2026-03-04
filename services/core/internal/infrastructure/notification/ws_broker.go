package notification

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

const wsBrokerWriteTimeout = 5 * time.Second

// WSBroker manages active WebSocket connections per user and pushes notifications.
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

// wsPushEvent is the JSON payload sent over the notification WebSocket.
type wsPushEvent struct {
	Type         string          `json:"type"` // "notification"
	ID           string          `json:"id"`
	NotifType    string          `json:"notif_type"`
	Title        string          `json:"title"`
	Body         string          `json:"body"`
	Data         json.RawMessage `json:"data,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UnreadCount  int64           `json:"unread_count"`
}

// Push sends a notification to all open connections for a user.
func (b *WSBroker) Push(userID string, n persistence.NotificationRow, unreadCount int64) {
	b.mu.RLock()
	conns := b.connections[userID]
	b.mu.RUnlock()

	if len(conns) == 0 {
		return
	}

	event := wsPushEvent{
		Type:        "notification",
		ID:          n.ID,
		NotifType:   n.Type,
		Title:       n.Title,
		Body:        n.Body,
		Data:        n.Data,
		CreatedAt:   n.CreatedAt,
		UnreadCount: unreadCount,
	}
	payload, _ := json.Marshal(event)

	// Dead connections discovered during write are cleaned up
	var dead []*websocket.Conn
	for _, conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(wsBrokerWriteTimeout))
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			b.log.Debug("ws push failed, marking dead", zap.String("user_id", userID), zap.Error(err))
			dead = append(dead, conn)
		}
	}
	for _, c := range dead {
		b.Unregister(userID, c)
	}
}
