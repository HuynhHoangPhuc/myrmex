package service

import (
	"testing"
)

func TestAC3PropagateEmptyDomainWipeout(t *testing.T) {
	v1, v2 := makeVar(1), makeVar(2)
	conflictAssign := makeAssign(10, 20, 1)
	domains := map[string][]Assignment{
		mustUUID(1).String(): {conflictAssign},
		mustUUID(2).String(): {conflictAssign},
	}
	result := ac3Propagate([]ScheduleVariable{v1, v2}, domains, openChecker(), nil)
	if !result {
		t.Fatal("AC3 with nil slots should not wipe domains")
	}
}

func TestAC3PropagateValidDomainsReturnsTrue(t *testing.T) {
	v1, v2 := makeVar(1), makeVar(2)
	domains := map[string][]Assignment{
		mustUUID(1).String(): {makeAssign(10, 20, 1)},
		mustUUID(2).String(): {makeAssign(11, 21, 2)},
	}
	if !ac3Propagate([]ScheduleVariable{v1, v2}, domains, openChecker(), nil) {
		t.Fatal("expected true for non-conflicting domains")
	}
}
