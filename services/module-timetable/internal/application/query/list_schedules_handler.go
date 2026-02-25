package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
)

// ListSchedulesQuery requests schedules with optional semester filter and pagination.
type ListSchedulesQuery struct {
	// SemesterID is optional; pass uuid.Nil to list all schedules.
	SemesterID uuid.UUID
	Page       int32 // 1-based; default 1
	PageSize   int32 // default 25
}

// ListSchedulesResult holds paginated schedules and total count.
type ListSchedulesResult struct {
	Schedules []*entity.Schedule
	Total     int64
	Page      int32
	PageSize  int32
}

// ListSchedulesHandler executes the ListSchedules read use case.
type ListSchedulesHandler struct {
	repo repository.ScheduleRepository
}

func NewListSchedulesHandler(repo repository.ScheduleRepository) *ListSchedulesHandler {
	return &ListSchedulesHandler{repo: repo}
}

func (h *ListSchedulesHandler) Handle(ctx context.Context, q ListSchedulesQuery) (*ListSchedulesResult, error) {
	page := q.Page
	if page <= 0 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	schedules, err := h.repo.ListPaged(ctx, q.SemesterID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list schedules paged: %w", err)
	}
	total, err := h.repo.CountSchedules(ctx, q.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("count schedules: %w", err)
	}
	return &ListSchedulesResult{
		Schedules: schedules,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}
