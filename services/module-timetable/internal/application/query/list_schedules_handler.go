package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
)

// ListSchedulesQuery requests all schedules for a semester.
type ListSchedulesQuery struct {
	SemesterID uuid.UUID
}

// ListSchedulesHandler executes the ListSchedules read use case.
type ListSchedulesHandler struct {
	repo repository.ScheduleRepository
}

func NewListSchedulesHandler(repo repository.ScheduleRepository) *ListSchedulesHandler {
	return &ListSchedulesHandler{repo: repo}
}

func (h *ListSchedulesHandler) Handle(ctx context.Context, q ListSchedulesQuery) ([]*entity.Schedule, error) {
	schedules, err := h.repo.ListBySemester(ctx, q.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("list schedules for semester %s: %w", q.SemesterID, err)
	}
	return schedules, nil
}
