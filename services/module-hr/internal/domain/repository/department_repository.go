package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
)

// DepartmentRepository defines persistence operations for Department.
type DepartmentRepository interface {
	Create(ctx context.Context, dept *entity.Department) (*entity.Department, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Department, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Department, error)
	Count(ctx context.Context) (int64, error)
}
