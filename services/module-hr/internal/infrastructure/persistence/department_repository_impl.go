package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/infrastructure/persistence/sqlc"
)

// DepartmentRepositoryImpl implements domain.DepartmentRepository using sqlc-generated queries.
type DepartmentRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewDepartmentRepository(queries *sqlc.Queries) *DepartmentRepositoryImpl {
	return &DepartmentRepositoryImpl{queries: queries}
}

func (r *DepartmentRepositoryImpl) Create(ctx context.Context, d *entity.Department) (*entity.Department, error) {
	row, err := r.queries.CreateDepartment(ctx, sqlc.CreateDepartmentParams{
		Name: d.Name,
		Code: d.Code,
	})
	if err != nil {
		return nil, fmt.Errorf("insert department: %w", err)
	}
	return hrDepartmentToEntity(row), nil
}

func (r *DepartmentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Department, error) {
	row, err := r.queries.GetDepartmentByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, fmt.Errorf("get department by id: %w", err)
	}
	return hrDepartmentToEntity(row), nil
}

func (r *DepartmentRepositoryImpl) List(ctx context.Context, limit, offset int32) ([]*entity.Department, error) {
	rows, err := r.queries.ListDepartments(ctx, sqlc.ListDepartmentsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	result := make([]*entity.Department, len(rows))
	for i, row := range rows {
		result[i] = hrDepartmentToEntity(row)
	}
	return result, nil
}

func (r *DepartmentRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.queries.CountDepartments(ctx)
}

func hrDepartmentToEntity(d sqlc.HrDepartment) *entity.Department {
	return &entity.Department{
		ID:        pgtypeToUUID(d.ID),
		Name:      d.Name,
		Code:      d.Code,
		CreatedAt: d.CreatedAt.Time,
		UpdatedAt: d.UpdatedAt.Time,
	}
}
