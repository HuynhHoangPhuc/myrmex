package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// DeleteStudentCommand carries the student ID for soft delete.
type DeleteStudentCommand struct {
	ID uuid.UUID
}

// DeleteStudentHandler handles soft delete.
type DeleteStudentHandler struct {
	repo      repository.StudentRepository
	publisher EventPublisher
}

func NewDeleteStudentHandler(repo repository.StudentRepository, publisher EventPublisher) *DeleteStudentHandler {
	return &DeleteStudentHandler{repo: repo, publisher: publisher}
}

func (h *DeleteStudentHandler) Handle(ctx context.Context, cmd DeleteStudentCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		return fmt.Errorf("delete student: %w", err)
	}
	_ = h.publisher.Publish(ctx, "student.deleted", map[string]string{"student_id": cmd.ID.String()})
	return nil
}
