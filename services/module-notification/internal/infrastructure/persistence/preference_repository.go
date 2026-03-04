package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
)

// PreferenceRepository persists per-user per-event per-channel notification preferences.
// Missing rows default to enabled (lazy initialization pattern).
type PreferenceRepository struct {
	pool *pgxpool.Pool
}

func NewPreferenceRepository(pool *pgxpool.Pool) *PreferenceRepository {
	return &PreferenceRepository{pool: pool}
}

// GetByUser returns all explicitly stored preference rows for a user.
// Missing rows should be treated as enabled by the caller.
func (r *PreferenceRepository) GetByUser(ctx context.Context, userID string) ([]entity.Preference, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_type, channel, enabled
		FROM notification.preferences
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []entity.Preference
	for rows.Next() {
		var p entity.Preference
		if err := rows.Scan(&p.EventType, &p.Channel, &p.Enabled); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}

// BulkUpsert inserts or updates preference rows for the given user.
func (r *PreferenceRepository) BulkUpsert(ctx context.Context, userID string, prefs []entity.Preference) error {
	for _, p := range prefs {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO notification.preferences (user_id, event_type, channel, enabled)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, event_type, channel)
			DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = NOW()
		`, userID, p.EventType, p.Channel, p.Enabled)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsEnabled checks whether the user has enabled a specific event+channel combination.
// Returns true when no row exists (default: all enabled).
func (r *PreferenceRepository) IsEnabled(ctx context.Context, userID, eventType, channel string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `
		SELECT enabled
		FROM notification.preferences
		WHERE user_id = $1 AND event_type = $2 AND channel = $3
	`, userID, eventType, channel).Scan(&enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil // no row = enabled by default
		}
		return false, err
	}
	return enabled, nil
}
