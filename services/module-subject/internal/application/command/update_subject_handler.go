package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// UpdateSubjectCommand carries partial update fields; nil/zero means "no change".
type UpdateSubjectCommand struct {
	ID           uuid.UUID
	Code         *string
	Name         *string
	Credits      *int32
	Description  *string
	DepartmentID *string
	WeeklyHours  *int32
	IsActive     *bool
}

// UpdateSubjectHandler handles UpdateSubjectCommand.
type UpdateSubjectHandler struct {
	subjectRepo repository.SubjectRepository
	publisher   EventPublisher
}

// NewUpdateSubjectHandler constructs an UpdateSubjectHandler.
func NewUpdateSubjectHandler(subjectRepo repository.SubjectRepository, publisher EventPublisher) *UpdateSubjectHandler {
	return &UpdateSubjectHandler{subjectRepo: subjectRepo, publisher: publisher}
}

// Handle executes the update subject use case.
func (h *UpdateSubjectHandler) Handle(ctx context.Context, cmd UpdateSubjectCommand) (*entity.Subject, error) {
	// Fetch existing to verify it exists.
	existing, err := h.subjectRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}

	// Apply partial updates.
	if cmd.Code != nil {
		existing.Code = *cmd.Code
	}
	if cmd.Name != nil {
		existing.Name = *cmd.Name
	}
	if cmd.Credits != nil {
		existing.Credits = *cmd.Credits
	}
	if cmd.Description != nil {
		existing.Description = *cmd.Description
	}
	if cmd.DepartmentID != nil {
		existing.DepartmentID = *cmd.DepartmentID
	}
	if cmd.WeeklyHours != nil {
		existing.WeeklyHours = *cmd.WeeklyHours
	}
	if cmd.IsActive != nil {
		existing.IsActive = *cmd.IsActive
	}

	if err := existing.Validate(); err != nil {
		return nil, fmt.Errorf("validate subject: %w", err)
	}

	updated, err := h.subjectRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("update subject: %w", err)
	}
	_ = h.publisher.Publish(ctx, "subject.updated", updated)
	return updated, nil
}
