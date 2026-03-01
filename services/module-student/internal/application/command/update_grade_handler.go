package command

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
)

// UpdateGradeCommand carries input for updating a grade.
type UpdateGradeCommand struct {
	GradeID      uuid.UUID
	GradeNumeric float64
	GradedBy     uuid.UUID
	Notes        string
}

// UpdateGradeHandler updates a grade and invalidates transcript cache.
type UpdateGradeHandler struct {
	enrollments repository.EnrollmentRepository
	grades      repository.GradeRepository
	cache       cache.Cache
}

func NewUpdateGradeHandler(
	enrollments repository.EnrollmentRepository,
	grades repository.GradeRepository,
	cache cache.Cache,
) *UpdateGradeHandler {
	return &UpdateGradeHandler{
		enrollments: enrollments,
		grades:      grades,
		cache:       cache,
	}
}

func (h *UpdateGradeHandler) Handle(ctx context.Context, cmd UpdateGradeCommand) (*entity.Grade, error) {
	if h.enrollments == nil || h.grades == nil {
		return nil, fmt.Errorf("repositories are required")
	}

	existing, err := h.grades.GetByID(ctx, cmd.GradeID)
	if err != nil {
		return nil, fmt.Errorf("get grade: %w", err)
	}
	enrollment, err := h.enrollments.GetByID(ctx, existing.EnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("get enrollment: %w", err)
	}

	existing.GradeNumeric = cmd.GradeNumeric
	existing.GradedBy = cmd.GradedBy
	existing.Notes = cmd.Notes
	if err := existing.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	updated, err := h.grades.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("update grade: %w", err)
	}
	if h.cache != nil {
		_ = h.cache.Delete(ctx, transcriptCacheKey(enrollment.StudentID))
	}
	return updated, nil
}
