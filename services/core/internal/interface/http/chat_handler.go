package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

const (
	wsWriteTimeout  = 10 * time.Second
	maxConvMessages = 40 // keep last 40 messages per session (in-memory)
)

// wsClientMessage is received from the browser over WebSocket.
type wsClientMessage struct {
	Type    string `json:"type"`    // "message"
	Content string `json:"content"` // user text
}

// wsServerEvent is sent from the server to the browser over WebSocket.
type wsServerEvent struct {
	Type    string `json:"type"`             // "text","tool_call","tool_result","done","error"
	Content string `json:"content,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Args    any    `json:"args,omitempty"`
	Result  string `json:"result,omitempty"`
}

// ChatHandler manages WebSocket AI chat sessions.
type ChatHandler struct {
	chatHandler *command.ChatMessageHandler
	jwtSvc      *auth.JWTService
	log         *zap.Logger
}

// NewChatHandler constructs a ChatHandler.
func NewChatHandler(
	chatHandler *command.ChatMessageHandler,
	jwtSvc *auth.JWTService,
	log *zap.Logger,
) *ChatHandler {
	return &ChatHandler{
		chatHandler: chatHandler,
		jwtSvc:      jwtSvc,
		log:         log,
	}
}

// HandleWebSocket upgrades the connection and drives the chat loop.
// Auth: JWT in ?token= query param (browsers cannot set WS headers).
// Route: GET /ws/chat?token=<jwt>
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.jwtSvc.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		h.log.Warn("websocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.CloseNow()

	h.log.Info("chat session started",
		zap.String("user_id", claims.UserID),
		zap.String("role", claims.Role),
	)

	// Per-session in-memory conversation history
	var history []llm.Message
	baseCtx := c.Request.Context()

	for {
		var clientMsg wsClientMessage
		if err := wsjson.Read(baseCtx, conn, &clientMsg); err != nil {
			if websocket.CloseStatus(err) != -1 {
				h.log.Debug("client disconnected", zap.String("user_id", claims.UserID))
			} else {
				h.log.Warn("ws read error", zap.String("user_id", claims.UserID), zap.Error(err))
			}
			return
		}

		if clientMsg.Type != "message" || clientMsg.Content == "" {
			continue
		}

		h.log.Debug("chat message received",
			zap.String("user_id", claims.UserID),
			zap.String("preview", truncate(clientMsg.Content, 80)),
		)

		events, err := h.chatHandler.Handle(baseCtx, clientMsg.Content, history)
		if err != nil {
			_ = h.writeEvent(baseCtx, conn, wsServerEvent{Type: "error", Content: err.Error()})
			continue
		}

		var assistantReply string
		for event := range events {
			ev := mapToWireEvent(event)
			if writeErr := h.writeEvent(baseCtx, conn, ev); writeErr != nil {
				h.log.Warn("ws write error", zap.Error(writeErr))
				return
			}
			if event.Type == "text" {
				assistantReply += event.Content
			}
		}

		// Persist turn in session history
		history = appendTurn(history, clientMsg.Content, assistantReply, maxConvMessages)
	}
}

// writeEvent sends a single JSON event to the WebSocket client with a write deadline.
func (h *ChatHandler) writeEvent(parent context.Context, conn *websocket.Conn, ev wsServerEvent) error {
	ctx, cancel := context.WithTimeout(parent, wsWriteTimeout)
	defer cancel()
	return wsjson.Write(ctx, conn, ev)
}

// mapToWireEvent converts an internal LLM stream event to the WebSocket wire format.
func mapToWireEvent(e llm.StreamEvent) wsServerEvent {
	switch e.Type {
	case "tool_call":
		if e.ToolCall == nil {
			return wsServerEvent{Type: "tool_call"}
		}
		var args any
		_ = json.Unmarshal(e.ToolCall.Arguments, &args)
		return wsServerEvent{Type: "tool_call", Tool: e.ToolCall.ToolName, Args: args}
	case "tool_result":
		toolName := ""
		if e.ToolCall != nil {
			toolName = e.ToolCall.ToolName
		}
		return wsServerEvent{Type: "tool_result", Tool: toolName, Result: e.Content}
	default:
		return wsServerEvent{Type: e.Type, Content: e.Content}
	}
}

// appendTurn adds user + assistant messages to history, trimming to maxMsgs.
func appendTurn(history []llm.Message, userMsg, assistantMsg string, maxMsgs int) []llm.Message {
	history = append(history,
		llm.Message{Role: "user", Content: userMsg},
		llm.Message{Role: "assistant", Content: assistantMsg},
	)
	if len(history) > maxMsgs {
		history = history[len(history)-maxMsgs:]
	}
	return history
}

// truncate shortens a string to at most n runes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
