package messaging

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

// auditPayload mirrors the JSON published by the audit middleware.
type auditPayload struct {
	UserID       string    `json:"user_id"`
	UserRole     string    `json:"user_role"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	OldValue     any       `json:"old_value,omitempty"`
	NewValue     any       `json:"new_value,omitempty"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	StatusCode   int       `json:"status_code"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditConsumer subscribes to AUDIT.logs events and writes them to core.audit_logs.
type AuditConsumer struct {
	consumer messaging.Consumer
	repo     *persistence.AuditLogRepository
	logger   *zap.Logger
}

func NewAuditConsumer(consumer messaging.Consumer, repo *persistence.AuditLogRepository, logger *zap.Logger) *AuditConsumer {
	return &AuditConsumer{consumer: consumer, repo: repo, logger: logger}
}

// Start ensures the AUDIT_EVENTS stream exists and begins consuming audit log events.
func (c *AuditConsumer) Start(ctx context.Context) error {
	// ~12-month retention (in seconds)
	const auditRetentionSecs = 366 * 24 * 60 * 60

	if err := c.consumer.EnsureStream(ctx, messaging.StreamConfig{
		Name:     "AUDIT_EVENTS",
		Subjects: []string{"AUDIT.>"},
		MaxAge:   auditRetentionSecs,
	}); err != nil {
		return err
	}

	if err := c.consumer.Subscribe(ctx, "audit-db-writer", "AUDIT.logs", c.handleMessage); err != nil {
		return err
	}

	c.logger.Info("audit consumer started")
	return nil
}

func (c *AuditConsumer) handleMessage(msg *messaging.Message) error {
	var event auditPayload
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		c.logger.Warn("audit: malformed event, discarding", zap.Error(err))
		return nil // non-retryable — ack and discard
	}

	entry := persistence.AuditLogEntry{
		UserID:       event.UserID,
		UserRole:     event.UserRole,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		OldValue:     event.OldValue,
		NewValue:     event.NewValue,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		StatusCode:   event.StatusCode,
		CreatedAt:    event.CreatedAt,
	}

	if err := c.repo.Insert(context.Background(), entry); err != nil {
		c.logger.Error("audit: DB insert failed", zap.Error(err), zap.String("action", event.Action))
		return err // retryable — nack
	}
	return nil
}
