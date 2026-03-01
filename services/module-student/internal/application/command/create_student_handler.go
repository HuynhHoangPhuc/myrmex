package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// CreateStudentCommand carries input for creating a student.
type CreateStudentCommand struct {
	StudentCode    string
	FullName       string
	Email          string
	DepartmentID   uuid.UUID
	EnrollmentYear int
}

// CreateStudentHandler handles student creation.
type CreateStudentHandler struct {
	repo      repository.StudentRepository
	publisher EventPublisher
}

func NewCreateStudentHandler(repo repository.StudentRepository, publisher EventPublisher) *CreateStudentHandler {
	return &CreateStudentHandler{repo: repo, publisher: publisher}
}

func (h *CreateStudentHandler) Handle(ctx context.Context, cmd CreateStudentCommand) (*entity.Student, error) {
	student := &entity.Student{
		StudentCode:    cmd.StudentCode,
		FullName:       cmd.FullName,
		Email:          cmd.Email,
		DepartmentID:   cmd.DepartmentID,
		EnrollmentYear: cmd.EnrollmentYear,
		Status:         entity.StudentStatusActive,
		IsActive:       true,
	}
	if err := student.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	created, err := h.repo.Create(ctx, student)
	if err != nil {
		return nil, fmt.Errorf("create student: %w", err)
	}
	_ = h.publisher.Publish(ctx, "student.created", created)
	return created, nil
}
