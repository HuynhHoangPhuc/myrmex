package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// UpdateTeacherCommand carries partial update fields.
type UpdateTeacherCommand struct {
	ID              uuid.UUID
	FullName        string
	Email           string
	Phone           string
	Title           string
	DepartmentID    *uuid.UUID
	MaxHoursPerWeek int
	IsActive        bool
}

// UpdateTeacherHandler handles teacher updates.
type UpdateTeacherHandler struct {
	repo repository.TeacherRepository
}

func NewUpdateTeacherHandler(repo repository.TeacherRepository) *UpdateTeacherHandler {
	return &UpdateTeacherHandler{repo: repo}
}

func (h *UpdateTeacherHandler) Handle(ctx context.Context, cmd UpdateTeacherCommand) (*entity.Teacher, error) {
	teacher := &entity.Teacher{
		ID:              cmd.ID,
		FullName:        cmd.FullName,
		Email:           cmd.Email,
		Phone:           cmd.Phone,
		Title:           cmd.Title,
		DepartmentID:    cmd.DepartmentID,
		MaxHoursPerWeek: cmd.MaxHoursPerWeek,
		IsActive:        cmd.IsActive,
	}
	updated, err := h.repo.Update(ctx, teacher)
	if err != nil {
		return nil, fmt.Errorf("update teacher: %w", err)
	}
	return updated, nil
}
