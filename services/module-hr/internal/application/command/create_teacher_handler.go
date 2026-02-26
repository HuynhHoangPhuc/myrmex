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
	Specializations []string
}

// CreateTeacherHandler handles teacher creation.
type CreateTeacherHandler struct {
	repo      repository.TeacherRepository
	publisher EventPublisher
}

func NewCreateTeacherHandler(repo repository.TeacherRepository, publisher EventPublisher) *CreateTeacherHandler {
	return &CreateTeacherHandler{repo: repo, publisher: publisher}
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
	for _, spec := range cmd.Specializations {
		if err := h.repo.AddSpecialization(ctx, created.ID, spec); err != nil {
			return nil, fmt.Errorf("add specialization %q: %w", spec, err)
		}
	}
	created.Specializations = cmd.Specializations

	// Fire-and-forget event publish
	_ = h.publisher.Publish(ctx, "hr.teacher.created", created)

	return created, nil
}
