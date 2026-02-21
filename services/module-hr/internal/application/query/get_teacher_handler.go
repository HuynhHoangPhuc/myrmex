package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// GetTeacherQuery carries the teacher ID to look up.
type GetTeacherQuery struct {
	ID uuid.UUID
}

// GetTeacherHandler fetches a single teacher by ID with specializations.
type GetTeacherHandler struct {
	repo repository.TeacherRepository
}

func NewGetTeacherHandler(repo repository.TeacherRepository) *GetTeacherHandler {
	return &GetTeacherHandler{repo: repo}
}

func (h *GetTeacherHandler) Handle(ctx context.Context, q GetTeacherQuery) (*entity.Teacher, error) {
	teacher, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get teacher: %w", err)
	}
	specs, err := h.repo.ListSpecializations(ctx, teacher.ID)
	if err != nil {
		return nil, fmt.Errorf("list specializations: %w", err)
	}
	teacher.Specializations = specs
	return teacher, nil
}
