package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
)

// GenerationStatus summarises the current state of an async generation job.
type GenerationStatus struct {
	Schedule    *entity.Schedule
	IsComplete  bool
	IsPartial   bool
	IsFailed    bool
}

// GetGenerationStatusHandler polls the schedule record to determine job state.
type GetGenerationStatusHandler struct {
	repo repository.ScheduleRepository
}

func NewGetGenerationStatusHandler(repo repository.ScheduleRepository) *GetGenerationStatusHandler {
	return &GetGenerationStatusHandler{repo: repo}
}

func (h *GetGenerationStatusHandler) Handle(ctx context.Context, scheduleID uuid.UUID) (*GenerationStatus, error) {
	schedule, err := h.repo.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}

	// score == -1 is the sentinel written by markFailed
	isFailed := schedule.Score < 0
	isComplete := schedule.GeneratedAt != nil && !isFailed
	isPartial := isComplete && schedule.Status == valueobject.ScheduleStatusDraft && schedule.Score < 100

	return &GenerationStatus{
		Schedule:   schedule,
		IsComplete: isComplete,
		IsPartial:  isPartial,
		IsFailed:   isFailed,
	}, nil
}
