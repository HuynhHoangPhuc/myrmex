package repository

import (
	"context"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
)

// NotificationRepository defines persistence operations for notifications.
type NotificationRepository interface {
	Insert(ctx context.Context, userID, notifType, channel, title, body string, metadata any) (string, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int32) ([]entity.Notification, int64, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
}
