package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
)

// DeleteSubjectCommand carries the ID of the subject to delete.
type DeleteSubjectCommand struct {
	ID uuid.UUID
}

// DeleteSubjectHandler handles DeleteSubjectCommand.
type DeleteSubjectHandler struct {
	subjectRepo repository.SubjectRepository
	publisher   EventPublisher
}

// NewDeleteSubjectHandler constructs a DeleteSubjectHandler.
func NewDeleteSubjectHandler(subjectRepo repository.SubjectRepository, publisher EventPublisher) *DeleteSubjectHandler {
	return &DeleteSubjectHandler{subjectRepo: subjectRepo, publisher: publisher}
}

// Handle executes the delete subject use case.
func (h *DeleteSubjectHandler) Handle(ctx context.Context, cmd DeleteSubjectCommand) error {
	if err := h.subjectRepo.Delete(ctx, cmd.ID); err != nil {
		return fmt.Errorf("delete subject: %w", err)
	}
	_ = h.publisher.Publish(ctx, "subject.deleted", map[string]string{"subject_id": cmd.ID.String()})
	return nil
}
