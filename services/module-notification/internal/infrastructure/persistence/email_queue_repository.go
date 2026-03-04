package persistence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EmailQueueRow is a pending email delivery record.
type EmailQueueRow struct {
	ID             string
	NotificationID string
	RecipientEmail string
	Subject        string
	HTMLBody       string
	Status         string // pending | sent | failed
	RetryCount     int32
	MaxRetries     int32
	NextRetryAt    time.Time
	ErrorMessage   string
	CreatedAt      time.Time
	SentAt         *time.Time
}

// EmailQueueRepository manages the notification.email_queue table.
type EmailQueueRepository struct {
	pool *pgxpool.Pool
}

func NewEmailQueueRepository(pool *pgxpool.Pool) *EmailQueueRepository {
	return &EmailQueueRepository{pool: pool}
}

// Enqueue inserts a new pending email into the queue.
func (r *EmailQueueRepository) Enqueue(ctx context.Context, notificationID, recipientEmail, subject, htmlBody string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notification.email_queue
			(notification_id, recipient_email, subject, html_body, status, retry_count, max_retries, next_retry_at)
		VALUES ($1::uuid, $2, $3, $4, 'pending', 0, 3, NOW())`,
		notificationID, recipientEmail, subject, htmlBody,
	)
	return err
}

// FetchPending returns up to `limit` pending rows whose next_retry_at is due.
func (r *EmailQueueRepository) FetchPending(ctx context.Context, limit int32) ([]EmailQueueRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, notification_id, recipient_email, subject, html_body,
		       status, retry_count, max_retries, next_retry_at, COALESCE(error_message,''), created_at, sent_at
		FROM notification.email_queue
		WHERE status = 'pending' AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []EmailQueueRow
	for rows.Next() {
		var q EmailQueueRow
		if err := rows.Scan(
			&q.ID, &q.NotificationID, &q.RecipientEmail, &q.Subject, &q.HTMLBody,
			&q.Status, &q.RetryCount, &q.MaxRetries, &q.NextRetryAt, &q.ErrorMessage, &q.CreatedAt, &q.SentAt,
		); err != nil {
			return nil, err
		}
		results = append(results, q)
	}
	return results, rows.Err()
}

// MarkSent marks a queue item as successfully delivered.
func (r *EmailQueueRepository) MarkSent(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification.email_queue SET status='sent', sent_at=NOW() WHERE id=$1::uuid`,
		id,
	)
	return err
}

// MarkFailed marks a queue item as permanently failed (retry limit exceeded).
func (r *EmailQueueRepository) MarkFailed(ctx context.Context, id, errMsg string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification.email_queue SET status='failed', error_message=$2 WHERE id=$1::uuid`,
		id, errMsg,
	)
	return err
}

// IncrementRetry increments retry count and sets next_retry_at using exponential backoff (2^n * 30s).
func (r *EmailQueueRepository) IncrementRetry(ctx context.Context, id, errMsg string, retryCount int32) error {
	// backoff: 30s, 60s, 120s, ...
	backoff := time.Duration(1<<retryCount) * 30 * time.Second
	nextRetry := time.Now().Add(backoff)
	_, err := r.pool.Exec(ctx, `
		UPDATE notification.email_queue
		SET retry_count = retry_count + 1, next_retry_at = $2, error_message = $3
		WHERE id = $1::uuid`,
		id, nextRetry, errMsg,
	)
	return err
}
