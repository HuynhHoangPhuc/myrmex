package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// ListPrerequisitesQuery carries the subject ID to list prerequisites for.
type ListPrerequisitesQuery struct {
	SubjectID uuid.UUID
}

// ListPrerequisitesHandler handles ListPrerequisitesQuery.
type ListPrerequisitesHandler struct {
	prereqRepo repository.PrerequisiteRepository
}

// NewListPrerequisitesHandler constructs a ListPrerequisitesHandler.
func NewListPrerequisitesHandler(prereqRepo repository.PrerequisiteRepository) *ListPrerequisitesHandler {
	return &ListPrerequisitesHandler{prereqRepo: prereqRepo}
}

// Handle returns all prerequisites for the given subject.
func (h *ListPrerequisitesHandler) Handle(ctx context.Context, q ListPrerequisitesQuery) ([]*entity.Prerequisite, error) {
	prereqs, err := h.prereqRepo.ListBySubject(ctx, q.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("list prerequisites: %w", err)
	}
	return prereqs, nil
}
