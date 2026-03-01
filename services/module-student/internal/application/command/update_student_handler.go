package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// UpdateStudentCommand carries full student state for update.
type UpdateStudentCommand struct {
	ID             uuid.UUID
	StudentCode    string
	UserID         *uuid.UUID
	FullName       string
	Email          string
	DepartmentID   uuid.UUID
	EnrollmentYear int
	Status         string
	IsActive       bool
}

// UpdateStudentHandler handles student updates.
type UpdateStudentHandler struct {
	repo      repository.StudentRepository
	publisher EventPublisher
}

func NewUpdateStudentHandler(repo repository.StudentRepository, publisher EventPublisher) *UpdateStudentHandler {
	return &UpdateStudentHandler{repo: repo, publisher: publisher}
}

func (h *UpdateStudentHandler) Handle(ctx context.Context, cmd UpdateStudentCommand) (*entity.Student, error) {
	student := &entity.Student{
		ID:             cmd.ID,
		StudentCode:    cmd.StudentCode,
		UserID:         cmd.UserID,
		FullName:       cmd.FullName,
		Email:          cmd.Email,
		DepartmentID:   cmd.DepartmentID,
		EnrollmentYear: cmd.EnrollmentYear,
		Status:         cmd.Status,
		IsActive:       cmd.IsActive,
	}
	if err := student.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	updated, err := h.repo.Update(ctx, student)
	if err != nil {
		return nil, fmt.Errorf("update student: %w", err)
	}
	_ = h.publisher.Publish(ctx, "student.updated", updated)
	return updated, nil
}
