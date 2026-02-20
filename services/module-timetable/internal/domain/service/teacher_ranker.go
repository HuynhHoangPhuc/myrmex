package service

import (
	"sort"

	"github.com/google/uuid"
)

// TeacherInfo holds denormalised teacher data used for ranking.
type TeacherInfo struct {
	ID              uuid.UUID
	FullName        string
	Specializations []string
	MaxHoursPerWeek int
}

// TeacherRank holds a scored teacher suggestion for a given subject/slot.
type TeacherRank struct {
	TeacherID uuid.UUID
	Name      string
	Score     float64
	Reasons   []string
}

// TeacherRanker ranks teachers for a given subject based on specialization,
// availability, and current load in the partial assignment.
type TeacherRanker struct {
	checker *ConstraintChecker
}

func NewTeacherRanker(checker *ConstraintChecker) *TeacherRanker {
	return &TeacherRanker{checker: checker}
}

// RankForSubject returns teachers sorted by suitability (highest score first).
func (r *TeacherRanker) RankForSubject(
	subjectID uuid.UUID,
	teachers []TeacherInfo,
	currentAssignment map[string]Assignment,
) []TeacherRank {
	ranks := make([]TeacherRank, 0, len(teachers))

	for _, t := range teachers {
		score := 0.0
		var reasons []string

		// +30: teacher has a matching specialization
		if r.checker.teacherHasSpecialization(t.ID, subjectID) {
			score += 30
			reasons = append(reasons, "specialization match")
		}

		// +up to 20: based on remaining capacity (fewer assigned hours = more headroom)
		assigned := countAssignedHours(t.ID, currentAssignment)
		headroom := t.MaxHoursPerWeek - assigned
		if headroom > 0 {
			if headroom > 20 {
				headroom = 20
			}
			score += float64(headroom)
			reasons = append(reasons, "available capacity")
		}

		// +5: availability data was provided (teacher is reachable)
		if _, ok := r.checker.teacherAvailability[t.ID]; ok {
			score += 5
			reasons = append(reasons, "availability defined")
		}

		ranks = append(ranks, TeacherRank{
			TeacherID: t.ID,
			Name:      t.FullName,
			Score:     score,
			Reasons:   reasons,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Score > ranks[j].Score
	})
	return ranks
}

func countAssignedHours(teacherID uuid.UUID, assignment map[string]Assignment) int {
	count := 0
	for _, a := range assignment {
		if a.TeacherID == teacherID {
			count++
		}
	}
	return count
}
