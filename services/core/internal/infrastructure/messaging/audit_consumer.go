package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

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
	js     jetstream.JetStream
	repo   *persistence.AuditLogRepository
	logger *zap.Logger
	cancel context.CancelFunc
}

func NewAuditConsumer(js jetstream.JetStream, repo *persistence.AuditLogRepository, logger *zap.Logger) *AuditConsumer {
	return &AuditConsumer{js: js, repo: repo, logger: logger}
}

// Start creates (or reuses) the AUDIT_EVENTS stream and begins the durable consumer goroutine.
func (c *AuditConsumer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "AUDIT_EVENTS",
		Subjects: []string{"AUDIT.>"},
		MaxAge:   366 * 24 * time.Hour, // ~12-month retention
	})
	if err != nil {
		return err
	}

	cons, err := c.js.CreateOrUpdateConsumer(ctx, "AUDIT_EVENTS", jetstream.ConsumerConfig{
		Durable:       "audit-db-writer",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "AUDIT.logs",
	})
	if err != nil {
		return err
	}

	go func() {
		iter, err := cons.Messages()
		if err != nil {
			c.logger.Error("audit consumer subscribe failed", zap.Error(err))
			return
		}
		for {
			select {
			case <-ctx.Done():
				iter.Stop()
				return
			default:
				msg, err := iter.Next()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					c.logger.Warn("audit consumer next error", zap.Error(err))
					continue
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	c.logger.Info("audit consumer started")
	return nil
}

func (c *AuditConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *AuditConsumer) handleMessage(ctx context.Context, msg jetstream.Msg) {
	var event auditPayload
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		c.logger.Warn("audit: malformed event, discarding", zap.Error(err))
		_ = msg.Ack()
		return
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

	if err := c.repo.Insert(ctx, entry); err != nil {
		c.logger.Error("audit: DB insert failed", zap.Error(err), zap.String("action", event.Action))
		_ = msg.Nak() // nack → JetStream retries
		return
	}

	_ = msg.Ack()
}
