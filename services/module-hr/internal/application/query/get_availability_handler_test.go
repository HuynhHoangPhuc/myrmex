package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

// availabilityQueryRepo extends mockTeacherRepo with ListAvailability support.
type availabilityQueryRepo struct {
	mockTeacherRepo
	slots   []*entity.Availability
	listErr error
}

func (m *availabilityQueryRepo) ListAvailability(_ context.Context, _ uuid.UUID) ([]*entity.Availability, error) {
	return m.slots, m.listErr
}

func TestGetAvailabilityHandler_Success(t *testing.T) {
	teacherID := uuid.New()
	slots := []*entity.Availability{
		{TeacherID: teacherID, DayOfWeek: 1, StartPeriod: 1, EndPeriod: 3},
		{TeacherID: teacherID, DayOfWeek: 3, StartPeriod: 5, EndPeriod: 7},
	}
	repo := &availabilityQueryRepo{slots: slots}
	h := NewGetAvailabilityHandler(repo)

	result, err := h.Handle(context.Background(), GetAvailabilityQuery{TeacherID: teacherID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(result))
	}
}

func TestGetAvailabilityHandler_Empty(t *testing.T) {
	repo := &availabilityQueryRepo{slots: []*entity.Availability{}}
	h := NewGetAvailabilityHandler(repo)

	result, err := h.Handle(context.Background(), GetAvailabilityQuery{TeacherID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 slots, got %d", len(result))
	}
}

func TestGetAvailabilityHandler_RepoError(t *testing.T) {
	repo := &availabilityQueryRepo{listErr: errors.New("db error")}
	h := NewGetAvailabilityHandler(repo)

	_, err := h.Handle(context.Background(), GetAvailabilityQuery{TeacherID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
