package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

// SubjectRepository defines persistence operations for Subject aggregates.
type SubjectRepository interface {
	Create(ctx context.Context, subject *entity.Subject) (*entity.Subject, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subject, error)
	GetByCode(ctx context.Context, code string) (*entity.Subject, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.Subject, error)
	ListByDepartment(ctx context.Context, departmentID string, limit, offset int32) ([]*entity.Subject, error)
	Count(ctx context.Context) (int64, error)
	CountByDepartment(ctx context.Context, departmentID string) (int64, error)
	Update(ctx context.Context, subject *entity.Subject) (*entity.Subject, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// ListAllIDs returns all subject UUIDs (used by DAG service for graph traversal).
	ListAllIDs(ctx context.Context) ([]uuid.UUID, error)
}
