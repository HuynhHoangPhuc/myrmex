package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// DeleteTeacherCommand carries the teacher ID for soft-delete.
type DeleteTeacherCommand struct {
	ID uuid.UUID
}

// DeleteTeacherHandler handles soft-deletion of teachers (sets is_active=false).
type DeleteTeacherHandler struct {
	repo      repository.TeacherRepository
	publisher EventPublisher
}

func NewDeleteTeacherHandler(repo repository.TeacherRepository, publisher EventPublisher) *DeleteTeacherHandler {
	return &DeleteTeacherHandler{repo: repo, publisher: publisher}
}

func (h *DeleteTeacherHandler) Handle(ctx context.Context, cmd DeleteTeacherCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		return fmt.Errorf("delete teacher: %w", err)
	}
	_ = h.publisher.Publish(ctx, "hr.teacher.deleted", map[string]string{"teacher_id": cmd.ID.String()})
	return nil
}
