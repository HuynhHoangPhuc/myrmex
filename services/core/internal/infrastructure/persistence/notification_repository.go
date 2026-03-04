package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationRow is a notification record returned from the DB.
type NotificationRow struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Body      string
	Data      json.RawMessage
	ReadAt    *time.Time
	CreatedAt time.Time
}

// NotificationPreferences holds per-user channel opt-in settings.
type NotificationPreferences struct {
	UserID        string
	EmailEnabled  bool
	InAppEnabled  bool
	DisabledTypes []string
}

// NotificationRepository provides read/write access to core.notifications.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// Insert persists a new notification row.
func (r *NotificationRepository) Insert(ctx context.Context, userID, notifType, title, body string, data any) (string, error) {
	var dataJSON []byte
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return "", err
		}
	}

	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO core.notifications (user_id, type, title, body, data)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id`,
		userID, notifType, title, body, dataJSON,
	).Scan(&id)
	return id, err
}

// ListByUser returns paginated notifications for a user (newest first).
func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, page, pageSize int32) ([]NotificationRow, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM core.notifications WHERE user_id = $1::uuid`,
		userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, type, title, body, data, read_at, created_at
		FROM core.notifications
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []NotificationRow
	for rows.Next() {
		var n NotificationRow
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, n)
	}
	return results, total, rows.Err()
}

// CountUnread returns the number of unread notifications for a user.
func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM core.notifications WHERE user_id = $1::uuid AND read_at IS NULL`,
		userID,
	).Scan(&count)
	return count, err
}

// MarkRead marks a single notification as read (only if owned by userID).
func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE core.notifications SET read_at = NOW()
		WHERE id = $1::uuid AND user_id = $2::uuid AND read_at IS NULL`,
		id, userID,
	)
	return err
}

// MarkAllRead marks all unread notifications for a user as read.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE core.notifications SET read_at = NOW()
		WHERE user_id = $1::uuid AND read_at IS NULL`,
		userID,
	)
	return err
}

// GetPreferences returns user notification preferences (defaults if not set).
func (r *NotificationRepository) GetPreferences(ctx context.Context, userID string) (NotificationPreferences, error) {
	prefs := NotificationPreferences{UserID: userID, EmailEnabled: true, InAppEnabled: true}
	err := r.pool.QueryRow(ctx, `
		SELECT email_enabled, inapp_enabled, disabled_types
		FROM core.notification_preferences WHERE user_id = $1::uuid`,
		userID,
	).Scan(&prefs.EmailEnabled, &prefs.InAppEnabled, &prefs.DisabledTypes)
	if errors.Is(err, pgx.ErrNoRows) {
		return prefs, nil // no row → use defaults
	}
	return prefs, err
}

// UpsertPreferences creates or updates user notification preferences.
func (r *NotificationRepository) UpsertPreferences(ctx context.Context, prefs NotificationPreferences) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO core.notification_preferences (user_id, email_enabled, inapp_enabled, disabled_types)
		VALUES ($1::uuid, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			email_enabled  = EXCLUDED.email_enabled,
			inapp_enabled  = EXCLUDED.inapp_enabled,
			disabled_types = EXCLUDED.disabled_types,
			updated_at     = NOW()`,
		prefs.UserID, prefs.EmailEnabled, prefs.InAppEnabled, prefs.DisabledTypes,
	)
	return err
}

// FindDeptHeadUserID returns the user ID of the dept_head for a given department.
func (r *NotificationRepository) FindDeptHeadUserID(ctx context.Context, departmentID string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM core.users WHERE role = 'dept_head' AND department_id = $1::uuid LIMIT 1`,
		departmentID,
	).Scan(&userID)
	return userID, err
}
