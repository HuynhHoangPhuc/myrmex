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
	repo repository.TeacherRepository
}

func NewDeleteTeacherHandler(repo repository.TeacherRepository) *DeleteTeacherHandler {
	return &DeleteTeacherHandler{repo: repo}
}

func (h *DeleteTeacherHandler) Handle(ctx context.Context, cmd DeleteTeacherCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		return fmt.Errorf("delete teacher: %w", err)
	}
	return nil
}
