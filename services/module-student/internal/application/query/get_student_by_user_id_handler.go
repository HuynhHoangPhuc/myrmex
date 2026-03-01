package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// GetStudentByUserIDQuery carries user_id to look up the associated student.
type GetStudentByUserIDQuery struct {
	UserID uuid.UUID
}

// GetStudentByUserIDHandler fetches a student by their linked user_id.
type GetStudentByUserIDHandler struct {
	repo repository.StudentRepository
}

func NewGetStudentByUserIDHandler(repo repository.StudentRepository) *GetStudentByUserIDHandler {
	return &GetStudentByUserIDHandler{repo: repo}
}

func (h *GetStudentByUserIDHandler) Handle(ctx context.Context, q GetStudentByUserIDQuery) (*entity.Student, error) {
	student, err := h.repo.GetByUserID(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("get student by user_id: %w", err)
	}
	return student, nil
}
