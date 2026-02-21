package command

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// CreateSubjectCommand carries the data needed to create a new subject.
type CreateSubjectCommand struct {
	Code         string
	Name         string
	Credits      int32
	Description  string
	DepartmentID string
	WeeklyHours  int32
}

// CreateSubjectHandler handles CreateSubjectCommand and returns the created Subject.
type CreateSubjectHandler struct {
	subjectRepo repository.SubjectRepository
}

// NewCreateSubjectHandler constructs a CreateSubjectHandler.
func NewCreateSubjectHandler(subjectRepo repository.SubjectRepository) *CreateSubjectHandler {
	return &CreateSubjectHandler{subjectRepo: subjectRepo}
}

// Handle executes the create subject use case.
func (h *CreateSubjectHandler) Handle(ctx context.Context, cmd CreateSubjectCommand) (*entity.Subject, error) {
	subject := &entity.Subject{
		Code:         cmd.Code,
		Name:         cmd.Name,
		Credits:      cmd.Credits,
		Description:  cmd.Description,
		DepartmentID: cmd.DepartmentID,
		WeeklyHours:  cmd.WeeklyHours,
		IsActive:     true,
	}

	if err := subject.Validate(); err != nil {
		return nil, fmt.Errorf("validate subject: %w", err)
	}

	created, err := h.subjectRepo.Create(ctx, subject)
	if err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return created, nil
}
