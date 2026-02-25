package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

// availabilityMockRepo extends mockTeacherRepo with availability tracking.
type availabilityMockRepo struct {
	mockTeacherRepo
	deleteAvailErr  error
	upsertAvailErr  error
	upsertedSlots   []*entity.Availability
}

func (m *availabilityMockRepo) DeleteAvailability(_ context.Context, _ uuid.UUID) error {
	return m.deleteAvailErr
}

func (m *availabilityMockRepo) UpsertAvailability(_ context.Context, a *entity.Availability) (*entity.Availability, error) {
	if m.upsertAvailErr != nil {
		return nil, m.upsertAvailErr
	}
	m.upsertedSlots = append(m.upsertedSlots, a)
	return a, nil
}

func TestUpdateAvailabilityHandler_Success(t *testing.T) {
	teacherID := uuid.New()
	repo := &availabilityMockRepo{}
	h := NewUpdateAvailabilityHandler(repo)

	slots, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: teacherID,
		Slots: []AvailabilitySlot{
			{DayOfWeek: 1, StartPeriod: 1, EndPeriod: 3},
			{DayOfWeek: 3, StartPeriod: 5, EndPeriod: 7},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
}

func TestUpdateAvailabilityHandler_EmptySlots(t *testing.T) {
	teacherID := uuid.New()
	repo := &availabilityMockRepo{}
	h := NewUpdateAvailabilityHandler(repo)

	slots, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: teacherID,
		Slots:     []AvailabilitySlot{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots, got %d", len(slots))
	}
}

func TestUpdateAvailabilityHandler_InvalidSlot_DayOutOfRange(t *testing.T) {
	repo := &availabilityMockRepo{}
	h := NewUpdateAvailabilityHandler(repo)

	_, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: uuid.New(),
		Slots:     []AvailabilitySlot{{DayOfWeek: 7, StartPeriod: 1, EndPeriod: 3}},
	})
	if err == nil {
		t.Fatal("expected validation error for day out of range")
	}
}

func TestUpdateAvailabilityHandler_InvalidSlot_StartAfterEnd(t *testing.T) {
	repo := &availabilityMockRepo{}
	h := NewUpdateAvailabilityHandler(repo)

	_, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: uuid.New(),
		Slots:     []AvailabilitySlot{{DayOfWeek: 1, StartPeriod: 5, EndPeriod: 3}},
	})
	if err == nil {
		t.Fatal("expected validation error for start >= end")
	}
}

func TestUpdateAvailabilityHandler_DeleteError(t *testing.T) {
	repo := &availabilityMockRepo{deleteAvailErr: errors.New("delete failed")}
	h := NewUpdateAvailabilityHandler(repo)

	_, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: uuid.New(),
		Slots:     []AvailabilitySlot{{DayOfWeek: 0, StartPeriod: 1, EndPeriod: 2}},
	})
	if err == nil {
		t.Fatal("expected error from delete")
	}
}

func TestUpdateAvailabilityHandler_UpsertError(t *testing.T) {
	repo := &availabilityMockRepo{upsertAvailErr: errors.New("upsert failed")}
	h := NewUpdateAvailabilityHandler(repo)

	_, err := h.Handle(context.Background(), UpdateAvailabilityCommand{
		TeacherID: uuid.New(),
		Slots:     []AvailabilitySlot{{DayOfWeek: 2, StartPeriod: 1, EndPeriod: 4}},
	})
	if err == nil {
		t.Fatal("expected error from upsert")
	}
}
