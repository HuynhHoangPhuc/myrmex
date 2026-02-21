package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// CreateTeacherCommand carries input for creating a teacher.
type CreateTeacherCommand struct {
	EmployeeCode    string
	FullName        string
	Email           string
	Phone           string
	Title           string
	DepartmentID    *uuid.UUID
	MaxHoursPerWeek int
}

// CreateTeacherHandler handles teacher creation.
type CreateTeacherHandler struct {
	repo repository.TeacherRepository
}

func NewCreateTeacherHandler(repo repository.TeacherRepository) *CreateTeacherHandler {
	return &CreateTeacherHandler{repo: repo}
}

func (h *CreateTeacherHandler) Handle(ctx context.Context, cmd CreateTeacherCommand) (*entity.Teacher, error) {
	teacher := &entity.Teacher{
		EmployeeCode:    cmd.EmployeeCode,
		FullName:        cmd.FullName,
		Email:           cmd.Email,
		Phone:           cmd.Phone,
		Title:           cmd.Title,
		DepartmentID:    cmd.DepartmentID,
		MaxHoursPerWeek: cmd.MaxHoursPerWeek,
		IsActive:        true,
	}
	if err := teacher.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	created, err := h.repo.Create(ctx, teacher)
	if err != nil {
		return nil, fmt.Errorf("create teacher: %w", err)
	}
	return created, nil
}
