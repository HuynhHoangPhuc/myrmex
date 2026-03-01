package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

// StudentRepository defines persistence operations for Student aggregate.
type StudentRepository interface {
	Create(ctx context.Context, student *entity.Student) (*entity.Student, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Student, error)
	List(ctx context.Context, departmentID *uuid.UUID, status *string, limit, offset int32) ([]*entity.Student, error)
	Count(ctx context.Context, departmentID *uuid.UUID, status *string) (int64, error)
	Update(ctx context.Context, student *entity.Student) (*entity.Student, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
