package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
)

// InviteCodeRepositoryImpl implements InviteCodeRepository using sqlc queries.
type InviteCodeRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewInviteCodeRepository(queries *sqlc.Queries) repository.InviteCodeRepository {
	return &InviteCodeRepositoryImpl{queries: queries}
}

func (r *InviteCodeRepositoryImpl) Create(ctx context.Context, code *entity.InviteCode) (*entity.InviteCode, error) {
	row, err := r.queries.CreateInviteCode(ctx, sqlc.CreateInviteCodeParams{
		CodeHash:  code.CodeHash,
		StudentID: uuidToPgtype(code.StudentID),
		ExpiresAt: pgtype.Timestamptz{Time: code.ExpiresAt, Valid: true},
		CreatedBy: uuidToPgtype(code.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create invite code: %w", err)
	}
	return inviteCodeRowToEntity(row), nil
}

func (r *InviteCodeRepositoryImpl) FindByCodeHash(ctx context.Context, codeHash string) (*entity.InviteCode, error) {
	row, err := r.queries.FindInviteCodeByHash(ctx, codeHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("find invite code: %w", err)
	}
	return inviteCodeRowToEntity(row), nil
}

func (r *InviteCodeRepositoryImpl) MarkUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.InviteCode, error) {
	row, err := r.queries.MarkInviteCodeUsed(ctx, sqlc.MarkInviteCodeUsedParams{
		ID:           uuidToPgtype(id),
		UsedByUserID: uuidToPgtype(userID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invite code already claimed or not found")
		}
		return nil, fmt.Errorf("mark invite code used: %w", err)
	}
	return inviteCodeRowToEntity(row), nil
}

func (r *InviteCodeRepositoryImpl) InvalidateStudentCodes(ctx context.Context, studentID uuid.UUID) error {
	return r.queries.InvalidateStudentInviteCodes(ctx, uuidToPgtype(studentID))
}

func inviteCodeRowToEntity(row sqlc.InviteCodeRow) *entity.InviteCode {
	code := &entity.InviteCode{
		ID:        pgtypeToUUID(row.ID),
		CodeHash:  row.CodeHash,
		StudentID: pgtypeToUUID(row.StudentID),
		ExpiresAt: row.ExpiresAt.Time,
		CreatedBy: pgtypeToUUID(row.CreatedBy),
		CreatedAt: row.CreatedAt.Time,
	}
	if row.UsedAt.Valid {
		t := row.UsedAt.Time
		code.UsedAt = &t
	}
	if row.UsedByUserID.Valid {
		uid := pgtypeToUUID(row.UsedByUserID)
		code.UsedByUserID = &uid
	}
	return code
}

// inviteCodeExpiresAt computes expiry from hours (default 48).
func InviteCodeExpiresAt(hours int32) time.Time {
	if hours <= 0 {
		hours = 48
	}
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
