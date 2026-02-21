package service

import (
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
)

func TestSlotsOverlap(t *testing.T) {
	tests := []struct {
		name string
		a    *entity.TimeSlot
		b    *entity.TimeSlot
		want bool
	}{
		{
			name: "same day overlap",
			a:    makeSlot(1, 0, 1, 3),
			b:    makeSlot(2, 0, 2, 4),
			want: true,
		},
		{
			name: "same day adjacent",
			a:    makeSlot(1, 0, 1, 3),
			b:    makeSlot(2, 0, 3, 5),
			want: false,
		},
		{
			name: "same day no overlap",
			a:    makeSlot(1, 0, 1, 2),
			b:    makeSlot(2, 0, 3, 4),
			want: false,
		},
		{
			name: "different day",
			a:    makeSlot(1, 0, 1, 5),
			b:    makeSlot(2, 1, 1, 5),
			want: false,
		},
		{
			name: "identical slots",
			a:    makeSlot(1, 0, 2, 4),
			b:    makeSlot(2, 0, 2, 4),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slotsOverlap(tt.a, tt.b); got != tt.want {
				t.Fatalf("slotsOverlap=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestConstraintCheckerIsConsistent(t *testing.T) {
	teacherID := mustUUID(10)
	roomID := mustUUID(20)
	slotID := mustUUID(30)
	slot := makeSlot(30, 0, 1, 3)

	t.Run("nil slot returns false", func(t *testing.T) {
		cc := openChecker()
		val := makeAssign(10, 20, 99)
		if cc.IsConsistent(mustUUID(1), val, nil, slotsMap(slot)) {
			t.Fatal("expected false for missing slot")
		}
	})

	t.Run("teacher double-booked", func(t *testing.T) {
		cc := openChecker()
		existing := map[string]Assignment{"s1": {TeacherID: teacherID, RoomID: mustUUID(21), SlotID: slotID}}
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		slots := slotsMap(slot)
		if cc.IsConsistent(mustUUID(2), val, existing, slots) {
			t.Fatal("expected false: teacher double-booked")
		}
	})

	t.Run("room double-booked", func(t *testing.T) {
		cc := openChecker()
		existing := map[string]Assignment{"s1": {TeacherID: mustUUID(11), RoomID: roomID, SlotID: slotID}}
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		if cc.IsConsistent(mustUUID(2), val, existing, slotsMap(slot)) {
			t.Fatal("expected false: room double-booked")
		}
	})

	t.Run("specialization mismatch", func(t *testing.T) {
		subjID := mustUUID(1)
		cc := NewConstraintChecker(
			nil,
			map[uuid.UUID]map[string]bool{teacherID: {"math": true}},
			map[uuid.UUID][]string{subjID: {"physics"}},
			nil,
		)
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		if cc.IsConsistent(subjID, val, nil, slotsMap(slot)) {
			t.Fatal("expected false: specialization mismatch")
		}
	})

	t.Run("teacher unavailable wrong day", func(t *testing.T) {
		cc := NewConstraintChecker(
			map[uuid.UUID][]*entity.TimeSlot{teacherID: {makeSlot(99, 1, 1, 3)}},
			nil, nil, nil,
		)
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		if cc.IsConsistent(mustUUID(1), val, nil, slotsMap(slot)) {
			t.Fatal("expected false: wrong day")
		}
	})

	t.Run("no availability data assumes available", func(t *testing.T) {
		cc := NewConstraintChecker(
			map[uuid.UUID][]*entity.TimeSlot{},
			nil, nil, nil,
		)
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		if !cc.IsConsistent(mustUUID(1), val, nil, slotsMap(slot)) {
			t.Fatal("expected true: no availability data = assume available")
		}
	})

	t.Run("all constraints satisfied", func(t *testing.T) {
		cc := openChecker()
		val := Assignment{TeacherID: teacherID, RoomID: roomID, SlotID: slotID}
		if !cc.IsConsistent(mustUUID(1), val, nil, slotsMap(slot)) {
			t.Fatal("expected true: all constraints satisfied")
		}
	})
}

func TestConstraintCheckerSoftConstraints(t *testing.T) {
	t.Run("gap penalty teacher 3 periods far apart", func(t *testing.T) {
		teacherID := mustUUID(10)
		s1 := makeSlot(1, 0, 1, 2)
		s2 := makeSlot(2, 0, 5, 6)
		s3 := makeSlot(3, 0, 9, 10)
		assignment := map[string]Assignment{
			"a": {TeacherID: teacherID, RoomID: mustUUID(20), SlotID: mustUUID(1)},
			"b": {TeacherID: teacherID, RoomID: mustUUID(20), SlotID: mustUUID(2)},
			"c": {TeacherID: teacherID, RoomID: mustUUID(20), SlotID: mustUUID(3)},
		}
		cc := openChecker()
		penalty := cc.EvaluateSoftConstraints(assignment, slotsMap(s1, s2, s3))
		if penalty != 6.0 {
			t.Fatalf("expected penalty=6.0, got=%f", penalty)
		}
	})

	t.Run("load imbalance penalty", func(t *testing.T) {
		tA, tB := mustUUID(1), mustUUID(2)
		s := makeSlot(1, 0, 1, 2)
		assignment := map[string]Assignment{
			"x1": {TeacherID: tA, RoomID: mustUUID(9), SlotID: mustUUID(1)},
			"x2": {TeacherID: tA, RoomID: mustUUID(9), SlotID: mustUUID(1)},
			"x3": {TeacherID: tA, RoomID: mustUUID(9), SlotID: mustUUID(1)},
			"x4": {TeacherID: tB, RoomID: mustUUID(9), SlotID: mustUUID(1)},
		}
		cc := openChecker()
		penalty := cc.EvaluateSoftConstraints(assignment, slotsMap(s))
		if penalty <= 0 {
			t.Fatalf("expected positive imbalance penalty, got=%f", penalty)
		}
	})

	t.Run("balanced load zero imbalance penalty", func(t *testing.T) {
		tA, tB := mustUUID(1), mustUUID(2)
		s := makeSlot(1, 0, 1, 2)
		assignment := map[string]Assignment{
			"x1": {TeacherID: tA, RoomID: mustUUID(9), SlotID: mustUUID(1)},
			"x2": {TeacherID: tB, RoomID: mustUUID(9), SlotID: mustUUID(1)},
		}
		cc := openChecker()
		penalty := cc.EvaluateSoftConstraints(assignment, slotsMap(s))
		if penalty != 0.0 {
			t.Fatalf("expected 0 penalty for balanced load, got=%f", penalty)
		}
	})
}
