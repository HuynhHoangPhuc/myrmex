package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
)

// ListTeachersQuery paginates active teachers, optionally filtered by department.
type ListTeachersQuery struct {
	Page         int32
	PageSize     int32
	DepartmentID *uuid.UUID
}

// ListTeachersResult wraps the teachers list with total count for pagination.
type ListTeachersResult struct {
	Teachers []*entity.Teacher
	Total    int64
}

// ListTeachersHandler handles paginated teacher listing.
type ListTeachersHandler struct {
	repo repository.TeacherRepository
}

func NewListTeachersHandler(repo repository.TeacherRepository) *ListTeachersHandler {
	return &ListTeachersHandler{repo: repo}
}

func (h *ListTeachersHandler) Handle(ctx context.Context, q ListTeachersQuery) (*ListTeachersResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	var (
		teachers []*entity.Teacher
		err      error
	)
	if q.DepartmentID != nil {
		teachers, err = h.repo.ListByDepartment(ctx, *q.DepartmentID, q.PageSize, offset)
	} else {
		teachers, err = h.repo.List(ctx, q.PageSize, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}

	total, err := h.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count teachers: %w", err)
	}

	return &ListTeachersResult{Teachers: teachers, Total: total}, nil
}
