package valueobject

import "fmt"

// PrerequisiteType distinguishes mandatory vs advisory prerequisites.
type PrerequisiteType string

const (
	// PrerequisiteTypeHard - student must pass this before enrolling.
	PrerequisiteTypeHard PrerequisiteType = "hard"
	// PrerequisiteTypeSoft - advisory recommendation only.
	PrerequisiteTypeSoft PrerequisiteType = "soft"
)

// ParsePrerequisiteType validates and returns a PrerequisiteType.
func ParsePrerequisiteType(s string) (PrerequisiteType, error) {
	switch PrerequisiteType(s) {
	case PrerequisiteTypeHard, PrerequisiteTypeSoft:
		return PrerequisiteType(s), nil
	default:
		return "", fmt.Errorf("invalid prerequisite type: %q (must be 'hard' or 'soft')", s)
	}
}

func (p PrerequisiteType) IsValid() bool {
	return p == PrerequisiteTypeHard || p == PrerequisiteTypeSoft
}

func (p PrerequisiteType) String() string {
	return string(p)
}
