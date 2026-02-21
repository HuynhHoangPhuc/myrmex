package service

import "testing"

func TestSelectMRVPicksSmallestDomain(t *testing.T) {
	v1 := makeVar(1)
	v2 := makeVar(2)
	v3 := makeVar(3)
	domains := map[string][]Assignment{
		mustUUID(1).String(): {makeAssign(1, 1, 1), makeAssign(2, 2, 2)},
		mustUUID(2).String(): {makeAssign(3, 3, 3)},
		mustUUID(3).String(): {makeAssign(4, 4, 4), makeAssign(5, 5, 5), makeAssign(6, 6, 6)},
	}
	selected := selectMRV([]ScheduleVariable{v1, v2, v3}, domains, map[string]Assignment{})
	if selected == nil {
		t.Fatal("expected selection")
	}
	if selected.SubjectID != mustUUID(2) {
		t.Fatalf("MRV should pick v2 (domain size 1), got %v", selected.SubjectID)
	}
}

func TestSelectMRVSkipsAssigned(t *testing.T) {
	v1 := makeVar(1)
	v2 := makeVar(2)
	domains := map[string][]Assignment{
		mustUUID(1).String(): {makeAssign(1, 1, 1)},
		mustUUID(2).String(): {makeAssign(2, 2, 2), makeAssign(3, 3, 3)},
	}
	assignment := map[string]Assignment{mustUUID(1).String(): makeAssign(1, 1, 1)}
	selected := selectMRV([]ScheduleVariable{v1, v2}, domains, assignment)
	if selected == nil {
		t.Fatal("expected selection")
	}
	if selected.SubjectID != mustUUID(2) {
		t.Fatal("should skip assigned v1, select v2")
	}
}

func TestSelectMRVAllAssignedReturnsNil(t *testing.T) {
	v1 := makeVar(1)
	domains := map[string][]Assignment{mustUUID(1).String(): {makeAssign(1, 1, 1)}}
	assignment := map[string]Assignment{mustUUID(1).String(): makeAssign(1, 1, 1)}
	if selectMRV([]ScheduleVariable{v1}, domains, assignment) != nil {
		t.Fatal("expected nil when all variables assigned")
	}
}

func TestOrderLCVSingleValueReturnsSame(t *testing.T) {
	v := makeVar(1)
	domains := map[string][]Assignment{mustUUID(1).String(): {makeAssign(1, 1, 1)}}
	result := orderLCV(&v, []ScheduleVariable{v}, domains, nil, openChecker(), nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result))
	}
}

func TestOrderLCVMultiValueStableWhenScoresEqual(t *testing.T) {
	v := makeVar(1)
	domains := map[string][]Assignment{
		mustUUID(1).String(): {
			makeAssign(1, 1, 1),
			makeAssign(2, 2, 2),
		},
	}
	result := orderLCV(&v, []ScheduleVariable{v}, domains, nil, openChecker(), nil)
	if len(result) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result))
	}
	if result[0] != domains[mustUUID(1).String()][0] || result[1] != domains[mustUUID(1).String()][1] {
		t.Fatal("expected order to remain stable when scores are equal")
	}
}
