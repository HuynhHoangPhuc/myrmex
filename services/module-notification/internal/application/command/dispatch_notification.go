package command

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// PushPublisher is the minimal interface for publishing WS push events.
type PushPublisher interface {
	PublishPush(userID, notifID, notifType, title, body string, createdAt time.Time, unreadCount int64)
}

// DispatchNotificationCommand stores an in-app notification and publishes a NATS push event.
// Preference checks are performed before storage/delivery on each channel.
type DispatchNotificationCommand struct {
	repo      *persistence.NotificationRepository
	prefRepo  repository.PreferenceRepository
	publisher PushPublisher
	log       *zap.Logger
}

func NewDispatchNotificationCommand(
	repo *persistence.NotificationRepository,
	prefRepo repository.PreferenceRepository,
	publisher PushPublisher,
	log *zap.Logger,
) *DispatchNotificationCommand {
	return &DispatchNotificationCommand{repo: repo, prefRepo: prefRepo, publisher: publisher, log: log}
}

// DispatchInput carries the data needed to create and deliver one notification.
type DispatchInput struct {
	UserID    string
	Type      string         // e.g. "schedule.published"
	Channel   string         // "in_app" | "email"
	Title     string
	Body      string
	Metadata  map[string]any // resource_type, resource_id, link
}

// Execute stores the notification in the DB, then publishes a NATS push for WS delivery.
// Returns ("", nil) when the user has disabled this channel for this event type.
// For email channel, the caller is responsible for enqueueing to email_queue separately.
func (c *DispatchNotificationCommand) Execute(ctx context.Context, in DispatchInput) (string, error) {
	// Check user preference before storing or delivering
	enabled, err := c.prefRepo.IsEnabled(ctx, in.UserID, in.Type, in.Channel)
	if err != nil {
		c.log.Warn("dispatch: preference check failed, defaulting to enabled",
			zap.String("user_id", in.UserID),
			zap.String("type", in.Type),
			zap.String("channel", in.Channel),
			zap.Error(err),
		)
	} else if !enabled {
		c.log.Debug("dispatch: skipped — user disabled channel for event",
			zap.String("user_id", in.UserID),
			zap.String("type", in.Type),
			zap.String("channel", in.Channel),
		)
		return "", nil
	}

	id, err := c.repo.Insert(ctx, in.UserID, in.Type, in.Channel, in.Title, in.Body, in.Metadata)
	if err != nil {
		c.log.Error("dispatch: DB insert failed",
			zap.String("user_id", in.UserID),
			zap.String("type", in.Type),
			zap.Error(err),
		)
		return "", err
	}

	// Publish WS push event only for in-app channel
	if in.Channel == "in_app" {
		unread, _ := c.repo.CountUnread(ctx, in.UserID)
		c.publisher.PublishPush(in.UserID, id, in.Type, in.Title, in.Body, time.Now(), unread)
	}

	return id, nil
}
