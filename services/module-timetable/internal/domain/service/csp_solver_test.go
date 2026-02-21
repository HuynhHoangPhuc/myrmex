package service

import (
	"context"
	"errors"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
)

func TestCSPSolverSingleSubjectTrivial(t *testing.T) {
	subj := makeVar(1)
	slot := makeSlot(1, 0, 1, 3)
	assign := makeAssign(10, 20, 1)
	domains := map[string][]Assignment{subj.SubjectID.String(): {assign}}
	solver := NewCSPSolver([]ScheduleVariable{subj}, domains, slotsMap(slot), openChecker())
	result, err := solver.Solve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsPartial {
		t.Fatal("expected full solution")
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestCSPSolverEmptyDomainAC3Wipeout(t *testing.T) {
	subj := makeVar(1)
	domains := map[string][]Assignment{subj.SubjectID.String(): {}}
	solver := NewCSPSolver([]ScheduleVariable{subj}, domains, nil, openChecker())
	_, err := solver.Solve(context.Background())
	if !errors.Is(err, ErrNoFeasibleSolution) {
		t.Fatalf("expected ErrNoFeasibleSolution, got %v", err)
	}
}

func TestCSPSolverThreeSubjectsFullSolution(t *testing.T) {
	vars := []ScheduleVariable{makeVar(1), makeVar(2), makeVar(3)}
	slots := []*entity.TimeSlot{
		makeSlot(1, 0, 1, 2),
		makeSlot(2, 0, 3, 4),
		makeSlot(3, 0, 5, 6),
	}
	domains := map[string][]Assignment{
		mustUUID(1).String(): {makeAssign(10, 20, 1)},
		mustUUID(2).String(): {makeAssign(11, 21, 2)},
		mustUUID(3).String(): {makeAssign(12, 22, 3)},
	}
	solver := NewCSPSolver(vars, domains, slotsMap(slots...), openChecker())
	result, err := solver.Solve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsPartial {
		t.Fatal("expected complete solution")
	}
	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}
}

func TestCSPSolverCancelledContextReturnsPartialOrError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	subj := makeVar(1)
	assign := makeAssign(10, 20, 1)
	slot := makeSlot(1, 0, 1, 3)
	domains := map[string][]Assignment{subj.SubjectID.String(): {assign}}
	solver := NewCSPSolver([]ScheduleVariable{subj}, domains, slotsMap(slot), openChecker())
	result, err := solver.Solve(ctx)
	if err != nil && !errors.Is(err, ErrNoFeasibleSolution) {
		t.Fatalf("unexpected error type: %v", err)
	}
	if err == nil && !result.IsPartial {
		t.Fatal("cancelled solve should return partial or error")
	}
}

func TestCSPSolverConflictingAssignmentsReturnPartialOrError(t *testing.T) {
	v1 := makeVar(1)
	v2 := makeVar(2)
	slot := makeSlot(1, 0, 1, 3)
	conflict := makeAssign(10, 20, 1)
	domains := map[string][]Assignment{
		v1.SubjectID.String(): {conflict},
		v2.SubjectID.String(): {conflict},
	}
	solver := NewCSPSolver([]ScheduleVariable{v1, v2}, domains, slotsMap(slot), openChecker())
	result, err := solver.Solve(context.Background())
	if err != nil && !errors.Is(err, ErrNoFeasibleSolution) {
		t.Fatalf("unexpected error type: %v", err)
	}
	if err == nil && !result.IsPartial {
		t.Fatal("expected partial result or error for conflicting assignments")
	}
}
