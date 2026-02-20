package service

import "github.com/google/uuid"

// selectMRV picks the unassigned variable (subject) with the Minimum Remaining Values
// in its domain — fewest valid (teacher, room, slot) combinations left.
// This reduces backtracking by tackling the most constrained subject first.
func selectMRV(
	variables []ScheduleVariable,
	domains map[string][]Assignment,
	assignment map[string]Assignment,
) *ScheduleVariable {
	best := -1
	bestSize := -1
	for i, v := range variables {
		key := v.SubjectID.String()
		if _, assigned := assignment[key]; assigned {
			continue
		}
		size := len(domains[key])
		if bestSize < 0 || size < bestSize {
			bestSize = size
			best = i
		}
	}
	if best < 0 {
		return nil
	}
	return &variables[best]
}

// orderLCV sorts the domain values of a variable by Least Constraining Value —
// values that rule out the fewest options for neighbouring unassigned variables
// are tried first, keeping the search space as open as possible.
func orderLCV(
	variable *ScheduleVariable,
	variables []ScheduleVariable,
	domains map[string][]Assignment,
	assignment map[string]Assignment,
	checker *ConstraintChecker,
	slots map[uuid.UUID]*ScheduleVariable, // unused but kept for interface symmetry
) []Assignment {
	values := domains[variable.SubjectID.String()]
	if len(values) <= 1 {
		return values
	}

	type scored struct {
		val   Assignment
		score int // number of values eliminated from neighbours
	}

	ranked := make([]scored, len(values))
	for i, val := range values {
		eliminated := 0
		for _, neighbour := range variables {
			nKey := neighbour.SubjectID.String()
			if _, done := assignment[nKey]; done {
				continue
			}
			if nKey == variable.SubjectID.String() {
				continue
			}
			for _, nVal := range domains[nKey] {
				// Count how many neighbour values conflict with this choice
				if checker.Conflicts(
					variable.SubjectID.String(), val,
					nKey, nVal,
					nil, // slots map not needed for teacher/room conflict check
				) {
					eliminated++
				}
			}
		}
		ranked[i] = scored{val, eliminated}
	}

	// Sort ascending: values that eliminate fewer neighbours come first
	for i := 1; i < len(ranked); i++ {
		for j := i; j > 0 && ranked[j].score < ranked[j-1].score; j-- {
			ranked[j], ranked[j-1] = ranked[j-1], ranked[j]
		}
	}

	result := make([]Assignment, len(ranked))
	for i, s := range ranked {
		result[i] = s.val
	}
	return result
}
