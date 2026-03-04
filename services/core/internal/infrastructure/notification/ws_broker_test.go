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
)

func TestWSBroker_Register_Single(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	conn := &websocket.Conn{}
	userID := "user-123"

	broker.Register(userID, conn)

	if len(broker.connections[userID]) != 1 {
		t.Errorf("expected 1 connection, got %d", len(broker.connections[userID]))
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

func TestWSBroker_Send_NoConnections(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	// Should not panic with no connections
	broker.send("user-no-conn", []byte(`{"type":"notification"}`))
}

func TestWSBroker_Send_ConcurrentRegisterUnregister(t *testing.T) {
	broker := NewWSBroker(zap.NewNop())
	userID := "user-concurrent"
	var wg sync.WaitGroup
	conns := make([]*websocket.Conn, 10)
	for i := range conns {
		conns[i] = &websocket.Conn{}
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			broker.Register(userID, conns[idx])
		}(i)
	}
	wg.Wait()
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

func TestWSBroker_Send_WithMockWSConnection(t *testing.T) {
	type pushPayload struct {
		Type        string `json:"type"`
		ID          string `json:"id"`
		NotifType   string `json:"notif_type"`
		UnreadCount int64  `json:"unread_count"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{}
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		defer conn.Close()
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var event pushPayload
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
		if event.UnreadCount != 5 {
			t.Errorf("expected unread_count 5, got %d", event.UnreadCount)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	broker := NewWSBroker(zap.NewNop())
	broker.Register("user-push", conn)

	payload := `{"type":"notification","id":"notif-123","notif_type":"schedule_published","title":"Test","body":"Body","unread_count":5}`
	broker.send("user-push", []byte(payload))
	time.Sleep(100 * time.Millisecond)
}

func TestWSBroker_Send_DeadConnectionCleanup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{}
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer server.Close()

	broker := NewWSBroker(zap.NewNop())
	userID := "user-dead"

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("could not create test ws connection: %v", err)
	}
	conn.Close()

	broker.Register(userID, conn)
	broker.send(userID, []byte(`{"type":"notification"}`))

	if len(broker.connections[userID]) != 0 {
		t.Errorf("dead connection should be cleaned up, expected 0, got %d", len(broker.connections[userID]))
	}
}
