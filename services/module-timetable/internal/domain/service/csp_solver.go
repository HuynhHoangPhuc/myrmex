package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
)

// ErrNoFeasibleSolution is returned when AC-3 wipes a domain before search.
var ErrNoFeasibleSolution = errors.New("no feasible schedule solution found")

// ScheduleVariable represents a subject that must be assigned to a
// (teacher, room, time-slot) tuple.
type ScheduleVariable struct {
	SubjectID               uuid.UUID
	SubjectCode             string
	WeeklyHours             int
	RequiredSpecializations []string
}

// Assignment is a concrete (teacher, room, slot) tuple for one subject.
type Assignment struct {
	TeacherID uuid.UUID
	RoomID    uuid.UUID
	SlotID    uuid.UUID
}

// SolverResult holds the output of a CSP solve run.
type SolverResult struct {
	Entries        []*entity.ScheduleEntry
	Score          float64
	HardViolations int
	SoftPenalty    float64
	IsPartial      bool // true when solver timed out before full assignment
	Duration       time.Duration
}

// CSPSolver performs backtracking search with AC-3 pre-processing,
// MRV variable selection, and LCV value ordering.
type CSPSolver struct {
	variables  []ScheduleVariable
	domains    map[string][]Assignment       // subject_id_str -> valid assignments
	slots      map[uuid.UUID]*entity.TimeSlot // slot_id -> TimeSlot (for overlap checks)
	checker    *ConstraintChecker
	bestSoFar  map[string]Assignment // best partial assignment seen during search
}

// NewCSPSolver constructs a solver ready to call Solve.
func NewCSPSolver(
	variables []ScheduleVariable,
	domains map[string][]Assignment,
	slots map[uuid.UUID]*entity.TimeSlot,
	checker *ConstraintChecker,
) *CSPSolver {
	return &CSPSolver{
		variables: variables,
		domains:   cloneDomains(domains),
		slots:     slots,
		checker:   checker,
		bestSoFar: map[string]Assignment{},
	}
}

// Solve runs AC-3 pre-processing then backtracking search.
// It respects ctx cancellation and returns the best partial result on timeout.
func (csp *CSPSolver) Solve(ctx context.Context) (*SolverResult, error) {
	start := time.Now()

	// AC-3: prune infeasible values before search begins
	if !ac3Propagate(csp.variables, csp.domains, csp.checker, nil) {
		return nil, ErrNoFeasibleSolution
	}

	assignment := make(map[string]Assignment, len(csp.variables))
	final := csp.backtrack(ctx, assignment)

	isPartial := false
	if final == nil {
		// Timeout or dead-end — return best partial if available
		if len(csp.bestSoFar) == 0 {
			return nil, ErrNoFeasibleSolution
		}
		final = csp.bestSoFar
		isPartial = true
	}

	penalty := csp.checker.EvaluateSoftConstraints(final, csp.slots)
	entries := assignmentToEntries(final)

	return &SolverResult{
		Entries:     entries,
		Score:       100.0 - penalty,
		SoftPenalty: penalty,
		IsPartial:   isPartial,
		Duration:    time.Since(start),
	}, nil
}

// backtrack performs recursive backtracking with MRV + LCV.
// Returns the complete assignment on success, nil on failure/timeout.
func (csp *CSPSolver) backtrack(ctx context.Context, assignment map[string]Assignment) map[string]Assignment {
	// Honour context cancellation — return best partial found so far
	if ctx.Err() != nil {
		if len(csp.bestSoFar) > 0 {
			return csp.bestSoFar
		}
		return nil
	}

	// Track the most complete partial assignment seen
	if len(assignment) > len(csp.bestSoFar) {
		csp.bestSoFar = copyAssignment(assignment)
	}

	// Base case: all variables assigned
	if len(assignment) == len(csp.variables) {
		return assignment
	}

	// MRV: pick variable with smallest remaining domain
	variable := selectMRV(csp.variables, csp.domains, assignment)
	if variable == nil {
		return nil
	}

	// LCV: try values ordered by least constraining first
	for _, value := range orderLCV(variable, csp.variables, csp.domains, assignment, csp.checker, nil) {
		if !csp.checker.IsConsistent(variable.SubjectID, value, assignment, csp.slots) {
			continue
		}

		assignment[variable.SubjectID.String()] = value

		// Forward checking: save domains, propagate, recurse, restore
		savedDomains := cloneDomains(csp.domains)
		if csp.propagateAfterAssignment(variable.SubjectID, value) {
			result := csp.backtrack(ctx, assignment)
			if result != nil {
				return result
			}
		}

		csp.domains = savedDomains
		delete(assignment, variable.SubjectID.String())
	}

	return nil
}

// propagateAfterAssignment removes values from remaining domains that now
// conflict with the newly assigned (teacher, room, slot) tuple.
// Returns false if any domain becomes empty (forward-checking failure).
func (csp *CSPSolver) propagateAfterAssignment(subjectID uuid.UUID, val Assignment) bool {
	assigned := subjectID.String()
	for _, v := range csp.variables {
		key := v.SubjectID.String()
		if key == assigned {
			continue
		}
		if _, done := csp.domains[key]; !done {
			continue
		}
		newDomain := csp.domains[key][:0:0]
		for _, candidate := range csp.domains[key] {
			if !csp.checker.Conflicts(assigned, val, key, candidate, csp.slots) {
				newDomain = append(newDomain, candidate)
			}
		}
		if len(newDomain) == 0 {
			return false
		}
		csp.domains[key] = newDomain
	}
	return true
}

// --- helpers ---

func copyAssignment(src map[string]Assignment) map[string]Assignment {
	dst := make(map[string]Assignment, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneDomains(src map[string][]Assignment) map[string][]Assignment {
	dst := make(map[string][]Assignment, len(src))
	for k, vals := range src {
		cp := make([]Assignment, len(vals))
		copy(cp, vals)
		dst[k] = cp
	}
	return dst
}

func assignmentToEntries(assignment map[string]Assignment) []*entity.ScheduleEntry {
	entries := make([]*entity.ScheduleEntry, 0, len(assignment))
	for subjectIDStr, a := range assignment {
		subjectID, _ := uuid.Parse(subjectIDStr)
		entries = append(entries, &entity.ScheduleEntry{
			SubjectID: subjectID,
			TeacherID: a.TeacherID,
			RoomID:    a.RoomID,
			TimeSlotID: a.SlotID,
		})
	}
	return entries
}
