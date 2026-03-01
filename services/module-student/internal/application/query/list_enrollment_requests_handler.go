package query

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
)

// ListEnrollmentRequestsQuery paginates enrollment requests with optional filters.
type ListEnrollmentRequestsQuery struct {
	Page       int32
	PageSize   int32
	StudentID  *uuid.UUID
	SemesterID *uuid.UUID
	Status     *string
}

// ListEnrollmentRequestsResult wraps list data and total.
type ListEnrollmentRequestsResult struct {
	Enrollments []*entity.EnrollmentRequest
	Total       int64
}

// ListEnrollmentRequestsHandler handles paginated enrollment listing.
type ListEnrollmentRequestsHandler struct {
	repo repository.EnrollmentRepository
}

func NewListEnrollmentRequestsHandler(repo repository.EnrollmentRepository) *ListEnrollmentRequestsHandler {
	return &ListEnrollmentRequestsHandler{repo: repo}
}

func (h *ListEnrollmentRequestsHandler) Handle(ctx context.Context, q ListEnrollmentRequestsQuery) (*ListEnrollmentRequestsResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	enrollments, err := h.repo.List(ctx, q.StudentID, q.SemesterID, q.Status, q.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list enrollment requests: %w", err)
	}
	total, err := h.repo.Count(ctx, q.StudentID, q.SemesterID, q.Status)
	if err != nil {
		return nil, fmt.Errorf("count enrollment requests: %w", err)
	}

	return &ListEnrollmentRequestsResult{Enrollments: enrollments, Total: total}, nil
}
