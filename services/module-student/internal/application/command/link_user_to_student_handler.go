package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// LinkUserToStudentCommand associates an auth user with a student record.
type LinkUserToStudentCommand struct {
	StudentID uuid.UUID
	UserID    uuid.UUID
}

// LinkUserToStudentHandler handles linking a user account to a student record.
type LinkUserToStudentHandler struct {
	repo      repository.StudentRepository
	publisher EventPublisher
}

func NewLinkUserToStudentHandler(repo repository.StudentRepository, publisher EventPublisher) *LinkUserToStudentHandler {
	return &LinkUserToStudentHandler{repo: repo, publisher: publisher}
}

func (h *LinkUserToStudentHandler) Handle(ctx context.Context, cmd LinkUserToStudentCommand) (*entity.Student, error) {
	updated, err := h.repo.LinkUser(ctx, cmd.StudentID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("link user to student: %w", err)
	}
	_ = h.publisher.Publish(ctx, "student.user_linked", updated)
	return updated, nil
}
