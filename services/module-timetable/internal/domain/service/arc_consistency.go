package service

import "github.com/google/uuid"

// arc represents a directed constraint between two CSP variables.
type arc struct {
	from string // subject ID string
	to   string // subject ID string
}

// ac3Propagate runs the AC-3 arc-consistency algorithm to prune domains before
// the backtracking search begins. Returns false when a domain becomes empty
// (problem is infeasible).
func ac3Propagate(
	variables []ScheduleVariable,
	domains map[string][]Assignment,
	checker *ConstraintChecker,
	slots map[uuid.UUID]*ScheduleVariable,
) bool {
	// Build initial queue: all arcs between every pair of variables
	queue := buildAllArcs(variables)

	for len(queue) > 0 {
		a := queue[0]
		queue = queue[1:]

		if removeInconsistentValues(a.from, a.to, domains, checker) {
			if len(domains[a.from]) == 0 {
				return false // domain wipeout → infeasible
			}
			// Re-add arcs pointing to a.from so they are re-checked
			for _, v := range variables {
				key := v.SubjectID.String()
				if key != a.to && key != a.from {
					queue = append(queue, arc{from: key, to: a.from})
				}
			}
		}
	}
	return true
}

// removeInconsistentValues removes values from domain[xi] that have no support
// in domain[xj]. Returns true when at least one value was removed.
func removeInconsistentValues(
	xi, xj string,
	domains map[string][]Assignment,
	checker *ConstraintChecker,
) bool {
	removed := false
	newDomain := domains[xi][:0:0] // zero-length slice, same backing capacity

	for _, valI := range domains[xi] {
		hasSupport := false
		for _, valJ := range domains[xj] {
			if !checker.Conflicts(xi, valI, xj, valJ, nil) {
				hasSupport = true
				break
			}
		}
		if hasSupport {
			newDomain = append(newDomain, valI)
		} else {
			removed = true
		}
	}
	domains[xi] = newDomain
	return removed
}

func buildAllArcs(variables []ScheduleVariable) []arc {
	arcs := make([]arc, 0, len(variables)*(len(variables)-1))
	for i := range variables {
		for j := range variables {
			if i != j {
				arcs = append(arcs, arc{
					from: variables[i].SubjectID.String(),
					to:   variables[j].SubjectID.String(),
				})
			}
		}
	}
	return arcs
}
