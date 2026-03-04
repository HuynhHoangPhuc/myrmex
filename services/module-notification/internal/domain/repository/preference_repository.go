package repository

import (
	"context"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
)

// PreferenceRepository defines persistence operations for notification preferences.
type PreferenceRepository interface {
	GetByUser(ctx context.Context, userID string) ([]entity.Preference, error)
	BulkUpsert(ctx context.Context, userID string, prefs []entity.Preference) error
	IsEnabled(ctx context.Context, userID, eventType, channel string) (bool, error)
}
