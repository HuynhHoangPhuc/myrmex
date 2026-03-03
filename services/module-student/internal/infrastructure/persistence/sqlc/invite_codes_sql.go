package sqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// InviteCodeRow holds an invite code record from the student.invite_codes table.
// Manually maintained until sqlc generate can be run against a live DB.
type InviteCodeRow struct {
	ID           pgtype.UUID        `json:"id"`
	CodeHash     string             `json:"code_hash"`
	StudentID    pgtype.UUID        `json:"student_id"`
	ExpiresAt    pgtype.Timestamptz `json:"expires_at"`
	UsedAt       pgtype.Timestamptz `json:"used_at"`
	UsedByUserID pgtype.UUID        `json:"used_by_user_id"`
	CreatedBy    pgtype.UUID        `json:"created_by"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

type CreateInviteCodeParams struct {
	CodeHash  string             `json:"code_hash"`
	StudentID pgtype.UUID        `json:"student_id"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	CreatedBy pgtype.UUID        `json:"created_by"`
}

type MarkInviteCodeUsedParams struct {
	ID           pgtype.UUID `json:"id"`
	UsedByUserID pgtype.UUID `json:"used_by_user_id"`
}

func (q *Queries) CreateInviteCode(ctx context.Context, arg CreateInviteCodeParams) (InviteCodeRow, error) {
	const sql = `
		INSERT INTO student.invite_codes (code_hash, student_id, expires_at, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, code_hash, student_id, expires_at, used_at, used_by_user_id, created_by, created_at`

	var row InviteCodeRow
	err := q.db.QueryRow(ctx, sql,
		arg.CodeHash,
		arg.StudentID,
		arg.ExpiresAt,
		arg.CreatedBy,
	).Scan(
		&row.ID,
		&row.CodeHash,
		&row.StudentID,
		&row.ExpiresAt,
		&row.UsedAt,
		&row.UsedByUserID,
		&row.CreatedBy,
		&row.CreatedAt,
	)
	if err != nil {
		return InviteCodeRow{}, fmt.Errorf("create invite code: %w", err)
	}
	return row, nil
}

func (q *Queries) FindInviteCodeByHash(ctx context.Context, codeHash string) (InviteCodeRow, error) {
	const sql = `
		SELECT id, code_hash, student_id, expires_at, used_at, used_by_user_id, created_by, created_at
		FROM student.invite_codes
		WHERE code_hash = $1`

	var row InviteCodeRow
	err := q.db.QueryRow(ctx, sql, codeHash).Scan(
		&row.ID,
		&row.CodeHash,
		&row.StudentID,
		&row.ExpiresAt,
		&row.UsedAt,
		&row.UsedByUserID,
		&row.CreatedBy,
		&row.CreatedAt,
	)
	if err != nil {
		return InviteCodeRow{}, fmt.Errorf("find invite code by hash: %w", err)
	}
	return row, nil
}

// MarkInviteCodeUsed atomically claims a code using WHERE used_at IS NULL.
// Returns an error if the code was already claimed (TOCTOU protection).
func (q *Queries) MarkInviteCodeUsed(ctx context.Context, arg MarkInviteCodeUsedParams) (InviteCodeRow, error) {
	const sql = `
		UPDATE student.invite_codes
		SET used_at = NOW(), used_by_user_id = $2
		WHERE id = $1 AND used_at IS NULL
		RETURNING id, code_hash, student_id, expires_at, used_at, used_by_user_id, created_by, created_at`

	var row InviteCodeRow
	err := q.db.QueryRow(ctx, sql, arg.ID, arg.UsedByUserID).Scan(
		&row.ID,
		&row.CodeHash,
		&row.StudentID,
		&row.ExpiresAt,
		&row.UsedAt,
		&row.UsedByUserID,
		&row.CreatedBy,
		&row.CreatedAt,
	)
	if err != nil {
		return InviteCodeRow{}, fmt.Errorf("mark invite code used: %w", err)
	}
	return row, nil
}

func (q *Queries) InvalidateStudentInviteCodes(ctx context.Context, studentID pgtype.UUID) error {
	const sql = `
		UPDATE student.invite_codes
		SET expires_at = NOW()
		WHERE student_id = $1 AND used_at IS NULL AND expires_at > NOW()`

	_, err := q.db.Exec(ctx, sql, studentID)
	if err != nil {
		return fmt.Errorf("invalidate student invite codes: %w", err)
	}
	return nil
}
