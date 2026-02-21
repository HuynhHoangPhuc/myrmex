package entity

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/valueobject"
)

// Prerequisite represents a directed dependency edge between two subjects.
// SubjectID depends on PrerequisiteID (i.e. PrerequisiteID must be completed first).
type Prerequisite struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
	Type           valueobject.PrerequisiteType
	// Priority ranks multiple prerequisites (1 = highest priority).
	Priority int32
}

// Validate enforces invariants on the Prerequisite entity.
func (p *Prerequisite) Validate() error {
	if p.SubjectID == p.PrerequisiteID {
		return fmt.Errorf("a subject cannot be its own prerequisite")
	}
	if !p.Type.IsValid() {
		return fmt.Errorf("invalid prerequisite type: %s", p.Type)
	}
	if p.Priority < 1 || p.Priority > 5 {
		return fmt.Errorf("priority must be between 1 and 5, got %d", p.Priority)
	}
	return nil
}
