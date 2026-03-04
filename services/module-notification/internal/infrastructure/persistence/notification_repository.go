package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationRow is a single notification record from the DB.
type NotificationRow struct {
	ID        string
	UserID    string
	Type      string
	Channel   string
	Title     string
	Body      string
	Metadata  json.RawMessage
	ReadAt    *time.Time
	CreatedAt time.Time
}

// NotificationRepository provides read/write access to notification.notifications.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// Insert persists a new in-app notification and returns its UUID.
func (r *NotificationRepository) Insert(ctx context.Context, userID, notifType, channel, title, body string, metadata any) (string, error) {
	var metaJSON []byte
	if metadata != nil {
		var err error
		metaJSON, err = json.Marshal(metadata)
		if err != nil {
			return "", err
		}
	}

	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notification.notifications (user_id, type, channel, title, body, metadata)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id`,
		userID, notifType, channel, title, body, metaJSON,
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
		`SELECT COUNT(*) FROM notification.notifications WHERE user_id = $1::uuid`,
		userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, type, channel, title, body, metadata, read_at, created_at
		FROM notification.notifications
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
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Title, &n.Body, &n.Metadata, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, n)
	}
	return results, total, rows.Err()
}

// CountUnread returns the number of unread in-app notifications for a user.
func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notification.notifications WHERE user_id = $1::uuid AND channel = 'in_app' AND read_at IS NULL`,
		userID,
	).Scan(&count)
	return count, err
}

// MarkRead marks a single notification as read (only if owned by userID).
func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notification.notifications SET read_at = NOW()
		WHERE id = $1::uuid AND user_id = $2::uuid AND read_at IS NULL`,
		id, userID,
	)
	return err
}

// MarkAllRead marks all unread in-app notifications for a user as read.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notification.notifications SET read_at = NOW()
		WHERE user_id = $1::uuid AND channel = 'in_app' AND read_at IS NULL`,
		userID,
	)
	return err
}

// GetByID fetches a single notification (for publish-push after insert).
func (r *NotificationRepository) GetByID(ctx context.Context, id string) (NotificationRow, error) {
	var n NotificationRow
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, type, channel, title, body, metadata, read_at, created_at
		FROM notification.notifications WHERE id = $1::uuid`,
		id,
	).Scan(&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Title, &n.Body, &n.Metadata, &n.ReadAt, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return n, pgx.ErrNoRows
	}
	return n, err
}
