package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
)

// ListStudentsQuery paginates active students with optional filters.
type ListStudentsQuery struct {
	Page         int32
	PageSize     int32
	DepartmentID *uuid.UUID
	Status       *string
}

// ListStudentsResult wraps list data and total.
type ListStudentsResult struct {
	Students []*entity.Student
	Total    int64
}

// ListStudentsHandler handles paginated student listing.
type ListStudentsHandler struct {
	repo repository.StudentRepository
}

func NewListStudentsHandler(repo repository.StudentRepository) *ListStudentsHandler {
	return &ListStudentsHandler{repo: repo}
}

func (h *ListStudentsHandler) Handle(ctx context.Context, q ListStudentsQuery) (*ListStudentsResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	students, err := h.repo.List(ctx, q.DepartmentID, q.Status, q.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	total, err := h.repo.Count(ctx, q.DepartmentID, q.Status)
	if err != nil {
		return nil, fmt.Errorf("count students: %w", err)
	}

	return &ListStudentsResult{Students: students, Total: total}, nil
}
