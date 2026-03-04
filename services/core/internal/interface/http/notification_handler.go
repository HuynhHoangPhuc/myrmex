package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	notifinfra "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/notification"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

// NotificationHandler exposes notification REST endpoints and the /ws/notifications WebSocket.
type NotificationHandler struct {
	repo   *persistence.NotificationRepository
	broker *notifinfra.WSBroker
	jwtSvc *auth.JWTService
	log    *zap.Logger
}

func NewNotificationHandler(
	repo *persistence.NotificationRepository,
	broker *notifinfra.WSBroker,
	jwtSvc *auth.JWTService,
	log *zap.Logger,
) *NotificationHandler {
	return &NotificationHandler{repo: repo, broker: broker, jwtSvc: jwtSvc, log: log}
}

// List handles GET /api/notifications?page=&page_size=
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	page := int32(parseIntQuery(c.Query("page"), 1, 10000))
	pageSize := int32(parseIntQuery(c.Query("page_size"), 1, 50))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	rows, total, err := h.repo.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "page_size": pageSize})
}

// UnreadCount handles GET /api/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	count, err := h.repo.CountUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count unread"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkRead handles PATCH /api/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	if err := h.repo.MarkRead(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark read"})
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkAllRead handles POST /api/notifications/mark-all-read
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.repo.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all read"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetPreferences handles GET /api/notifications/preferences
func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	prefs, err := h.repo.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get preferences"})
		return
	}
	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences handles PUT /api/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	var body struct {
		EmailEnabled  bool     `json:"email_enabled"`
		InAppEnabled  bool     `json:"inapp_enabled"`
		DisabledTypes []string `json:"disabled_types"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	prefs := persistence.NotificationPreferences{
		UserID:        userID,
		EmailEnabled:  body.EmailEnabled,
		InAppEnabled:  body.InAppEnabled,
		DisabledTypes: body.DisabledTypes,
	}
	if prefs.DisabledTypes == nil {
		prefs.DisabledTypes = []string{}
	}
	if err := h.repo.UpsertPreferences(c.Request.Context(), prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update preferences"})
		return
	}
	c.JSON(http.StatusOK, prefs)
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
