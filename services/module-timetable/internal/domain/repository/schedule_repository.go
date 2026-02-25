package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
)

// ScheduleRepository defines persistence operations for Schedule aggregates.
type ScheduleRepository interface {
	Create(ctx context.Context, s *entity.Schedule) (*entity.Schedule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error)
	ListBySemester(ctx context.Context, semesterID uuid.UUID) ([]*entity.Schedule, error)
	// ListPaged returns schedules with optional semester filter; pass uuid.Nil to list all.
	ListPaged(ctx context.Context, semesterID uuid.UUID, limit, offset int32) ([]*entity.Schedule, error)
	// CountSchedules returns total count; pass uuid.Nil to count all.
	CountSchedules(ctx context.Context, semesterID uuid.UUID) (int64, error)
	UpdateResult(ctx context.Context, id uuid.UUID, score float64, hardViolations int, softPenalty float64) (*entity.Schedule, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status valueobject.ScheduleStatus) (*entity.Schedule, error)

	CreateEntry(ctx context.Context, e *entity.ScheduleEntry) (*entity.ScheduleEntry, error)
	GetEntry(ctx context.Context, id uuid.UUID) (*entity.ScheduleEntry, error)
	ListEntries(ctx context.Context, scheduleID uuid.UUID) ([]*entity.ScheduleEntry, error)
	UpdateEntry(ctx context.Context, e *entity.ScheduleEntry) (*entity.ScheduleEntry, error)
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	DeleteAllEntries(ctx context.Context, scheduleID uuid.UUID) error

	AppendEvent(ctx context.Context, aggregateID uuid.UUID, aggregateType, eventType string, payload json.RawMessage) error
}
