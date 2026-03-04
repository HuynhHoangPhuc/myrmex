package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	notifinfra "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/notification"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
)

// NotificationHandler exposes the /ws/notifications WebSocket endpoint.
// REST endpoints (list, mark-read, preferences) are reverse-proxied to module-notification.
type NotificationHandler struct {
	broker *notifinfra.WSBroker
	jwtSvc *auth.JWTService
	log    *zap.Logger
}

func NewNotificationHandler(
	broker *notifinfra.WSBroker,
	jwtSvc *auth.JWTService,
	log *zap.Logger,
) *NotificationHandler {
	return &NotificationHandler{broker: broker, jwtSvc: jwtSvc, log: log}
}

// HandleWebSocket upgrades to WS and registers the connection with WSBroker.
// Auth: JWT in ?token= query param. Route: GET /ws/notifications?token=<jwt>
func (h *NotificationHandler) HandleWebSocket(c *gin.Context) {
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Warn("notification ws upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	userID := claims.UserID
	h.broker.Register(userID, conn)
	defer h.broker.Unregister(userID, conn)

	h.log.Debug("notification ws connected", zap.String("user_id", userID))

	// Keep connection alive — server pushes only, no client messages expected
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
