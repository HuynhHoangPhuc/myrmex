package query

import (
	"context"
	"fmt"

	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-subject/internal/domain/repository"
)

// ListSubjectsQuery carries pagination and optional department filter.
type ListSubjectsQuery struct {
	Limit        int32
	Offset       int32
	DepartmentID string // empty means no filter
}

// ListSubjectsResult bundles subjects with total count for pagination.
type ListSubjectsResult struct {
	Subjects []*entity.Subject
	Total    int64
}

// ListSubjectsHandler handles ListSubjectsQuery.
type ListSubjectsHandler struct {
	subjectRepo repository.SubjectRepository
}

// NewListSubjectsHandler constructs a ListSubjectsHandler.
func NewListSubjectsHandler(subjectRepo repository.SubjectRepository) *ListSubjectsHandler {
	return &ListSubjectsHandler{subjectRepo: subjectRepo}
}

// Handle returns a paginated list of subjects.
func (h *ListSubjectsHandler) Handle(ctx context.Context, q ListSubjectsQuery) (*ListSubjectsResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	var (
		subjects []*entity.Subject
		total    int64
		err      error
	)

	if q.DepartmentID != "" {
		subjects, err = h.subjectRepo.ListByDepartment(ctx, q.DepartmentID, limit, q.Offset)
		if err != nil {
			return nil, fmt.Errorf("list subjects by department: %w", err)
		}
		total, err = h.subjectRepo.CountByDepartment(ctx, q.DepartmentID)
	} else {
		subjects, err = h.subjectRepo.List(ctx, limit, q.Offset)
		if err != nil {
			return nil, fmt.Errorf("list subjects: %w", err)
		}
		total, err = h.subjectRepo.Count(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("count subjects: %w", err)
	}

	return &ListSubjectsResult{Subjects: subjects, Total: total}, nil
}
