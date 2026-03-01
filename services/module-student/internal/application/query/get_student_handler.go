package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// GetStudentQuery carries student ID to look up.
type GetStudentQuery struct {
	ID uuid.UUID
}

// GetStudentHandler fetches a student by ID.
type GetStudentHandler struct {
	repo repository.StudentRepository
}

func NewGetStudentHandler(repo repository.StudentRepository) *GetStudentHandler {
	return &GetStudentHandler{repo: repo}
}

func (h *GetStudentHandler) Handle(ctx context.Context, q GetStudentQuery) (*entity.Student, error) {
	student, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	return student, nil
}
