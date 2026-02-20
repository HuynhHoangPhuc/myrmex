package valueobject

import (
	"fmt"
	"strings"
)

// Specialization is a validated tag attached to a teacher (e.g., "mathematics", "physics").
type Specialization string

// NewSpecialization validates and creates a Specialization value object.
func NewSpecialization(s string) (Specialization, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "", fmt.Errorf("specialization cannot be empty")
	}
	if len(s) > 100 {
		return "", fmt.Errorf("specialization exceeds 100 characters")
	}
	return Specialization(s), nil
}

func (s Specialization) String() string {
	return string(s)
}
