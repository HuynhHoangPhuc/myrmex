package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/entity"
)

// PrerequisiteRepository defines persistence operations for prerequisite edges.
type PrerequisiteRepository interface {
	Add(ctx context.Context, prereq *entity.Prerequisite) (*entity.Prerequisite, error)
	Remove(ctx context.Context, subjectID, prerequisiteID uuid.UUID) error
	ListBySubject(ctx context.Context, subjectID uuid.UUID) ([]*entity.Prerequisite, error)
	// ListAll returns every edge in the graph (used for full DAG traversal).
	ListAll(ctx context.Context) ([]*entity.Prerequisite, error)
	Get(ctx context.Context, subjectID, prerequisiteID uuid.UUID) (*entity.Prerequisite, error)
}
