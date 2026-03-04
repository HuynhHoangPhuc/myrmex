package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLogEntry is the domain record inserted into core.audit_logs.
type AuditLogEntry struct {
	UserID       string
	UserRole     string
	Action       string
	ResourceType string
	ResourceID   string // empty string → NULL
	OldValue     any    // nil → NULL
	NewValue     any    // nil → NULL
	IPAddress    string
	UserAgent    string
	StatusCode   int
	CreatedAt    time.Time
}

// AuditLogRow is returned by list queries.
type AuditLogRow struct {
	ID           string
	UserID       string
	UserRole     string
	Action       string
	ResourceType string
	ResourceID   *string
	OldValue     json.RawMessage
	NewValue     json.RawMessage
	IPAddress    *string
	UserAgent    *string
	StatusCode   *int
	CreatedAt    time.Time
}

// AuditLogFilter holds optional filter params for list queries.
type AuditLogFilter struct {
	UserID       string
	ResourceType string
	Action       string
	From         time.Time
	To           time.Time
	Page         int32
	PageSize     int32
}

// AuditLogRepository handles writes and reads to core.audit_logs.
type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) Insert(ctx context.Context, e AuditLogEntry) error {
	oldJSON, _ := json.Marshal(e.OldValue)
	newJSON, _ := json.Marshal(e.NewValue)

	if string(oldJSON) == "null" {
		oldJSON = nil
	}
	if string(newJSON) == "null" {
		newJSON = nil
	}

	var resourceID *string
	if e.ResourceID != "" {
		resourceID = &e.ResourceID
	}

	var ip, ua *string
	if e.IPAddress != "" {
		ip = &e.IPAddress
	}
	if e.UserAgent != "" {
		ua = &e.UserAgent
	}

	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO core.audit_logs
			(user_id, user_role, action, resource_type, resource_id,
			 old_value, new_value, ip_address, user_agent, status_code, created_at)
		VALUES ($1::uuid, $2, $3, $4, $5::uuid, $6, $7, $8::inet, $9, $10, $11)`,
		e.UserID, e.UserRole, e.Action, e.ResourceType, resourceID,
		oldJSON, newJSON, ip, ua, e.StatusCode, createdAt,
	)
	return err
}

// List returns audit logs with optional filters. Zero-value filter fields are ignored.
func (r *AuditLogRepository) List(ctx context.Context, f AuditLogFilter) ([]AuditLogRow, int64, error) {
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	from := f.From
	to := f.To
	if from.IsZero() {
		from = time.Now().AddDate(-1, 0, 0)
	}
	if to.IsZero() {
		to = time.Now().Add(24 * time.Hour)
	}

	// Build nullable filter args
	var userID, resourceType, action *string
	if f.UserID != "" {
		userID = &f.UserID
	}
	if f.ResourceType != "" {
		resourceType = &f.ResourceType
	}
	if f.Action != "" {
		action = &f.Action
	}

	// Count query
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM core.audit_logs
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3::uuid IS NULL OR user_id = $3::uuid)
		  AND ($4::text IS NULL OR resource_type = $4)
		  AND ($5::text IS NULL OR action ILIKE '%' || $5 || '%')`,
		from, to, userID, resourceType, action,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, user_role, action, resource_type, resource_id,
		       old_value, new_value, ip_address, user_agent, status_code, created_at
		FROM core.audit_logs
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3::uuid IS NULL OR user_id = $3::uuid)
		  AND ($4::text IS NULL OR resource_type = $4)
		  AND ($5::text IS NULL OR action ILIKE '%' || $5 || '%')
		ORDER BY created_at DESC
		LIMIT $6 OFFSET $7`,
		from, to, userID, resourceType, action, f.PageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []AuditLogRow
	for rows.Next() {
		var row AuditLogRow
		var idStr, userIDStr string
		if err := rows.Scan(
			&idStr, &userIDStr, &row.UserRole, &row.Action, &row.ResourceType,
			&row.ResourceID, &row.OldValue, &row.NewValue,
			&row.IPAddress, &row.UserAgent, &row.StatusCode, &row.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		row.ID = idStr
		row.UserID = userIDStr
		results = append(results, row)
	}
	return results, total, rows.Err()
}
