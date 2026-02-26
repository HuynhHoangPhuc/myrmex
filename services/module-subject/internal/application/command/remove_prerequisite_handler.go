package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// RemovePrerequisiteCommand carries the edge identifiers to remove.
type RemovePrerequisiteCommand struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
}

// RemovePrerequisiteHandler handles RemovePrerequisiteCommand.
type RemovePrerequisiteHandler struct {
	prereqRepo repository.PrerequisiteRepository
	publisher  EventPublisher
}

// NewRemovePrerequisiteHandler constructs a RemovePrerequisiteHandler.
func NewRemovePrerequisiteHandler(prereqRepo repository.PrerequisiteRepository, publisher EventPublisher) *RemovePrerequisiteHandler {
	return &RemovePrerequisiteHandler{prereqRepo: prereqRepo, publisher: publisher}
}

// Handle executes the remove prerequisite use case.
func (h *RemovePrerequisiteHandler) Handle(ctx context.Context, cmd RemovePrerequisiteCommand) error {
	if err := h.prereqRepo.Remove(ctx, cmd.SubjectID, cmd.PrerequisiteID); err != nil {
		return fmt.Errorf("remove prerequisite: %w", err)
	}
	_ = h.publisher.Publish(ctx, "subject.prerequisite.removed", map[string]string{
		"subject_id": cmd.SubjectID.String(), "prerequisite_id": cmd.PrerequisiteID.String(),
	})
	return nil
}
