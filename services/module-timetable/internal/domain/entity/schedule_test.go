package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
)

func TestSchedule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sched   Schedule
		wantErr bool
	}{
		{
			name:    "valid draft",
			sched:   Schedule{Name: "Auto-2026-1", Status: valueobject.ScheduleStatusDraft},
			wantErr: false,
		},
		{
			name:    "valid published",
			sched:   Schedule{Name: "Auto-2026-1", Status: valueobject.ScheduleStatusPublished},
			wantErr: false,
		},
		{
			name:    "missing name",
			sched:   Schedule{Name: "", Status: valueobject.ScheduleStatusDraft},
			wantErr: true,
		},
		{
			name:    "invalid status",
			sched:   Schedule{Name: "S1", Status: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sched.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchedule_Publish_Success(t *testing.T) {
	s := &Schedule{
		ID:             uuid.New(),
		Name:           "S1",
		Status:         valueobject.ScheduleStatusDraft,
		HardViolations: 0,
	}
	if err := s.Publish(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s.Status != valueobject.ScheduleStatusPublished {
		t.Fatalf("expected Published, got %s", s.Status)
	}
}

func TestSchedule_Publish_WithHardViolations(t *testing.T) {
	s := &Schedule{
		ID:             uuid.New(),
		Name:           "S1",
		Status:         valueobject.ScheduleStatusDraft,
		HardViolations: 2,
	}
	if err := s.Publish(); err == nil {
		t.Fatal("expected error due to hard violations")
	}
}

func TestSchedule_Publish_AlreadyPublished(t *testing.T) {
	s := &Schedule{
		ID:     uuid.New(),
		Name:   "S1",
		Status: valueobject.ScheduleStatusPublished,
	}
	if err := s.Publish(); err == nil {
		t.Fatal("expected error for non-draft status")
	}
}

func TestSchedule_Publish_Archived(t *testing.T) {
	s := &Schedule{
		ID:     uuid.New(),
		Name:   "S1",
		Status: valueobject.ScheduleStatusArchived,
	}
	if err := s.Publish(); err == nil {
		t.Fatal("expected error for archived status")
	}
}
