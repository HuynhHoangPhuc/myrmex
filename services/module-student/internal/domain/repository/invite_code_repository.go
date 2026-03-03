package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

// InviteCodeRepository handles persistence of invite codes.
type InviteCodeRepository interface {
	Create(ctx context.Context, code *entity.InviteCode) (*entity.InviteCode, error)
	FindByCodeHash(ctx context.Context, codeHash string) (*entity.InviteCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.InviteCode, error)
	InvalidateStudentCodes(ctx context.Context, studentID uuid.UUID) error
}
