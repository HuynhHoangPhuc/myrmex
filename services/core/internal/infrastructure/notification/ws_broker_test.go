package notification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

func TestWSBroker_Register_Single(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	conn := &websocket.Conn{}
	userID := "user-123"

	broker.Register(userID, conn)

	if len(broker.connections[userID]) != 1 {
		t.Errorf("expected 1 connection, got %d", len(broker.connections[userID]))
	}
	if broker.connections[userID][0] != conn {
		t.Error("connection not registered correctly")
	}
}

func TestWSBroker_Register_Multiple(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-456"
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	broker.Register(userID, conn1)
	broker.Register(userID, conn2)

	if len(broker.connections[userID]) != 2 {
		t.Errorf("expected 2 connections, got %d", len(broker.connections[userID]))
	}
}

func TestWSBroker_Unregister_Single(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-789"
	conn := &websocket.Conn{}

	broker.Register(userID, conn)
	broker.Unregister(userID, conn)

	if _, exists := broker.connections[userID]; exists {
		t.Error("expected connection to be cleaned up, user entry should be deleted")
	}
}

func TestWSBroker_Unregister_Multiple(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-multi"
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	broker.Register(userID, conn1)
	broker.Register(userID, conn2)
	broker.Unregister(userID, conn1)

	if len(broker.connections[userID]) != 1 {
		t.Errorf("expected 1 connection after unregister, got %d", len(broker.connections[userID]))
	}
	if broker.connections[userID][0] != conn2 {
		t.Error("wrong connection remaining")
	}
}

func TestWSBroker_Unregister_NonExistent(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-nonexist"
	conn := &websocket.Conn{}

	// Should not panic
	broker.Unregister(userID, conn)

	if _, exists := broker.connections[userID]; exists {
		t.Error("user entry should not exist after unregister of nonexistent connection")
	}
}

func TestWSBroker_Push_NoConnections(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	notif := persistence.NotificationRow{
		ID:        "notif-1",
		Type:      "schedule_published",
		Title:     "Test",
		Body:      "Body",
		CreatedAt: time.Now(),
	}

	// Should not panic with no connections
	broker.Push("user-no-conn", notif, 0)
}

func TestWSBroker_Push_ConcurrentRegisterUnregister(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-concurrent"
	var wg sync.WaitGroup
	conns := make([]*websocket.Conn, 10)

	// Create connections
	for i := 0; i < 10; i++ {
		conns[i] = &websocket.Conn{}
	}

	// Concurrent register
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			broker.Register(userID, conns[idx])
		}(i)
	}
	wg.Wait()

	// Concurrent unregister
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			broker.Unregister(userID, conns[idx])
		}(i)
	}
	wg.Wait()

	if len(broker.connections[userID]) != 5 {
		t.Errorf("expected 5 connections after concurrent ops, got %d", len(broker.connections[userID]))
	}
}

func TestWSBroker_Push_WithMockWSConnection(t *testing.T) {
	// Create a test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Read message from broker
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("read error: %v", err)
			return
		}

		// Parse and validate message
		var event wsPushEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("unmarshal error: %v", err)
			return
		}

		if event.Type != "notification" {
			t.Errorf("expected type notification, got %s", event.Type)
		}
		if event.ID != "notif-123" {
			t.Errorf("expected ID notif-123, got %s", event.ID)
		}
		if event.NotifType != "schedule_published" {
			t.Errorf("expected notif_type schedule_published, got %s", event.NotifType)
		}
		if event.UnreadCount != 5 {
			t.Errorf("expected unread_count 5, got %d", event.UnreadCount)
		}
	}))
	defer server.Close()

	// Connect to test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	broker := NewWSBroker(zap.NewNop())
	userID := "user-push"
	broker.Register(userID, conn)

	// Push notification
	notif := persistence.NotificationRow{
		ID:        "notif-123",
		UserID:    userID,
		Type:      "schedule_published",
		Title:     "Schedule Published",
		Body:      "Your schedule is ready",
		CreatedAt: time.Now(),
	}
	broker.Push(userID, notif, 5)

	// Connection should read the message in server handler above
	time.Sleep(100 * time.Millisecond)
}

func TestWSBroker_Push_DeadConnectionCleanup(t *testing.T) {
	// Create a test server that accepts connection then closes it
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Don't read/write, just let it sit (won't interfere with test)
		conn.Close()
	}))
	defer server.Close()

	broker := NewWSBroker(zap.NewNop())
	userID := "user-dead"

	// Connect to server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("could not create test ws connection: %v", err)
	}
	// Close connection immediately to simulate dead connection
	conn.Close()

	// Register the closed connection
	broker.Register(userID, conn)

	notif := persistence.NotificationRow{
		ID:        "notif-dead",
		Type:      "schedule_published",
		Title:     "Test",
		Body:      "Body",
		CreatedAt: time.Now(),
	}

	// Push should handle dead connection gracefully and clean it up
	broker.Push(userID, notif, 0)

	// Dead connection should be unregistered
	if len(broker.connections[userID]) != 0 {
		t.Errorf("dead connection should be cleaned up, expected 0, got %d", len(broker.connections[userID]))
	}
}

func TestWSBroker_PushPayloadFormat(t *testing.T) {
	// Create test server that validates received payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Read message from broker
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("read error: %v", err)
			return
		}

		// Parse and validate payload structure
		var event wsPushEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Errorf("unmarshal error: %v", err)
			return
		}

		if event.Type != "notification" {
			t.Errorf("event type mismatch: expected notification, got %s", event.Type)
		}
		if event.ID != "notif-999" {
			t.Errorf("event ID mismatch: expected notif-999, got %s", event.ID)
		}
		if event.NotifType != "enrollment_approved" {
			t.Errorf("notif_type mismatch: expected enrollment_approved, got %s", event.NotifType)
		}
		if event.Title != "Enrollment Approved" {
			t.Errorf("title mismatch: expected Enrollment Approved, got %s", event.Title)
		}
		if event.UnreadCount != 3 {
			t.Errorf("unread_count mismatch: expected 3, got %d", event.UnreadCount)
		}
	}))
	defer server.Close()

	broker := NewWSBroker(zap.NewNop())
	userID := "user-payload"

	// Connect to test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("could not create test ws connection: %v", err)
	}
	defer conn.Close()

	broker.Register(userID, conn)

	notif := persistence.NotificationRow{
		ID:        "notif-999",
		UserID:    userID,
		Type:      "enrollment_approved",
		Title:     "Enrollment Approved",
		Body:      "You are enrolled in Math",
		Data:      json.RawMessage(`{"subject":"Math"}`),
		CreatedAt: time.Now(),
	}

	broker.Push(userID, notif, 3)

	// Give server time to read the message
	time.Sleep(100 * time.Millisecond)
}

