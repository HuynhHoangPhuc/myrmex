package query

import (
	"context"
	"fmt"

	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/domain/repository"
)

// ListDepartmentsQuery paginates departments.
type ListDepartmentsQuery struct {
	Page     int32
	PageSize int32
}

// ListDepartmentsResult wraps the list with total count.
type ListDepartmentsResult struct {
	Departments []*entity.Department
	Total       int64
}

// ListDepartmentsHandler handles paginated department listing.
type ListDepartmentsHandler struct {
	repo repository.DepartmentRepository
}

func NewListDepartmentsHandler(repo repository.DepartmentRepository) *ListDepartmentsHandler {
	return &ListDepartmentsHandler{repo: repo}
}

func (h *ListDepartmentsHandler) Handle(ctx context.Context, q ListDepartmentsQuery) (*ListDepartmentsResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	depts, err := h.repo.List(ctx, q.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	total, err := h.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count departments: %w", err)
	}

	return &ListDepartmentsResult{Departments: depts, Total: total}, nil
}
