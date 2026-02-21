package service

import (
	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
)

// ConstraintChecker evaluates hard and soft scheduling constraints.
// All data is pre-loaded before the CSP solver starts (no IO during search).
type ConstraintChecker struct {
	// teacher_id -> list of available (day, startPeriod, endPeriod) slots
	teacherAvailability map[uuid.UUID][]*entity.TimeSlot
	// teacher_id -> set of specialization strings
	teacherSpecs map[uuid.UUID]map[string]bool
	// subject_id -> required specialization strings
	subjectSpecs map[uuid.UUID][]string
	// teacher_id -> max weekly hours
	teacherMaxHours map[uuid.UUID]int
}

func NewConstraintChecker(
	availability map[uuid.UUID][]*entity.TimeSlot,
	teacherSpecs map[uuid.UUID]map[string]bool,
	subjectSpecs map[uuid.UUID][]string,
	teacherMaxHours map[uuid.UUID]int,
) *ConstraintChecker {
	return &ConstraintChecker{
		teacherAvailability: availability,
		teacherSpecs:        teacherSpecs,
		subjectSpecs:        subjectSpecs,
		teacherMaxHours:     teacherMaxHours,
	}
}

// IsConsistent checks all hard constraints for a proposed assignment.
// Returns false if any hard constraint is violated.
func (cc *ConstraintChecker) IsConsistent(
	subjectID uuid.UUID,
	val Assignment,
	current map[string]Assignment,
	slots map[uuid.UUID]*entity.TimeSlot,
) bool {
	slot := slots[val.SlotID]
	if slot == nil {
		return false
	}

	for _, existing := range current {
		existingSlot := slots[existing.SlotID]
		if existingSlot == nil {
			continue
		}
		// Hard constraint 1: teacher not double-booked in overlapping period
		if existing.TeacherID == val.TeacherID && slotsOverlap(slot, existingSlot) {
			return false
		}
		// Hard constraint 2: room not double-booked in overlapping period
		if existing.RoomID == val.RoomID && slotsOverlap(slot, existingSlot) {
			return false
		}
	}

	// Hard constraint 3: teacher has required specialization
	if !cc.teacherHasSpecialization(val.TeacherID, subjectID) {
		return false
	}

	// Hard constraint 4: teacher available at this time slot
	if !cc.teacherAvailableAt(val.TeacherID, slot) {
		return false
	}

	return true
}

// Conflicts checks whether two assignments for different subjects conflict (used by AC-3).
func (cc *ConstraintChecker) Conflicts(
	xi string, valI Assignment,
	xj string, valJ Assignment,
	slots map[uuid.UUID]*entity.TimeSlot,
) bool {
	slotI := slots[valI.SlotID]
	slotJ := slots[valJ.SlotID]
	if slotI == nil || slotJ == nil {
		return false
	}
	if !slotsOverlap(slotI, slotJ) {
		return false
	}
	// Same teacher at same overlapping time = conflict
	if valI.TeacherID == valJ.TeacherID {
		return true
	}
	// Same room at same overlapping time = conflict
	if valI.RoomID == valJ.RoomID {
		return true
	}
	return false
}

// EvaluateSoftConstraints computes a total penalty score for a complete assignment.
func (cc *ConstraintChecker) EvaluateSoftConstraints(
	assignment map[string]Assignment,
	slots map[uuid.UUID]*entity.TimeSlot,
) float64 {
	penalty := 0.0
	penalty += cc.teacherGapPenalty(assignment, slots) * 2.0
	penalty += cc.loadImbalancePenalty(assignment) * 1.5
	return penalty
}

// --- helpers ---

func (cc *ConstraintChecker) teacherHasSpecialization(teacherID, subjectID uuid.UUID) bool {
	required, ok := cc.subjectSpecs[subjectID]
	if !ok || len(required) == 0 {
		return true // no requirement → any teacher is fine
	}
	specs := cc.teacherSpecs[teacherID]
	for _, req := range required {
		if specs[req] {
			return true
		}
	}
	return false
}

func (cc *ConstraintChecker) teacherAvailableAt(teacherID uuid.UUID, slot *entity.TimeSlot) bool {
	available, ok := cc.teacherAvailability[teacherID]
	if !ok || len(available) == 0 {
		return true // no availability data → assume available
	}
	for _, avail := range available {
		if avail.DayOfWeek == slot.DayOfWeek &&
			avail.StartPeriod <= slot.StartPeriod &&
			avail.EndPeriod >= slot.EndPeriod {
			return true
		}
	}
	return false
}

// teacherGapPenalty penalises teachers who have large idle gaps between classes.
func (cc *ConstraintChecker) teacherGapPenalty(
	assignment map[string]Assignment,
	slots map[uuid.UUID]*entity.TimeSlot,
) float64 {
	// group assigned slots per teacher per day
	type dayKey struct {
		teacherID uuid.UUID
		day       int
	}
	byDay := map[dayKey][]int{}
	for _, a := range assignment {
		slot := slots[a.SlotID]
		if slot == nil {
			continue
		}
		k := dayKey{a.TeacherID, slot.DayOfWeek}
		byDay[k] = append(byDay[k], slot.StartPeriod)
	}
	penalty := 0.0
	for _, periods := range byDay {
		if len(periods) < 2 {
			continue
		}
		min, max := periods[0], periods[0]
		for _, p := range periods[1:] {
			if p < min {
				min = p
			}
			if p > max {
				max = p
			}
		}
		span := max - min
		assigned := len(periods)
		gap := span - assigned
		if gap > 2 {
			penalty += float64(gap - 2)
		}
	}
	return penalty
}

// loadImbalancePenalty penalises uneven distribution of teaching hours.
func (cc *ConstraintChecker) loadImbalancePenalty(assignment map[string]Assignment) float64 {
	hours := map[uuid.UUID]int{}
	for _, a := range assignment {
		hours[a.TeacherID]++
	}
	if len(hours) == 0 {
		return 0
	}
	total := 0
	for _, h := range hours {
		total += h
	}
	avg := float64(total) / float64(len(hours))
	penalty := 0.0
	for _, h := range hours {
		diff := float64(h) - avg
		if diff < 0 {
			diff = -diff
		}
		penalty += diff
	}
	return penalty
}

func slotsOverlap(a, b *entity.TimeSlot) bool {
	if a.DayOfWeek != b.DayOfWeek {
		return false
	}
	return a.StartPeriod < b.EndPeriod && b.StartPeriod < a.EndPeriod
}
