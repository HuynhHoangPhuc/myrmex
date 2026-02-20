package valueobject

// ConstraintType differentiates hard constraints (must not violate) from soft
// constraints (penalised if violated).
type ConstraintType string

const (
	// Hard constraints — any violation makes the schedule infeasible.
	ConstraintTeacherConflict       ConstraintType = "teacher_conflict"
	ConstraintRoomConflict          ConstraintType = "room_conflict"
	ConstraintSpecializationMissing ConstraintType = "specialization_missing"
	ConstraintTeacherUnavailable    ConstraintType = "teacher_unavailable"

	// Soft constraints — violations accumulate a penalty score.
	ConstraintTeacherGap      ConstraintType = "teacher_gap"
	ConstraintLoadImbalance   ConstraintType = "load_imbalance"
	ConstraintPreferenceBreak ConstraintType = "preference_break"
)

// IsHard returns true for hard constraints.
func (c ConstraintType) IsHard() bool {
	switch c {
	case ConstraintTeacherConflict,
		ConstraintRoomConflict,
		ConstraintSpecializationMissing,
		ConstraintTeacherUnavailable:
		return true
	}
	return false
}
