package entity

import (
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
)

// Grade stores the evaluated result for one enrollment.
type Grade struct {
    ID          uuid.UUID
    EnrollmentID uuid.UUID
    GradeNumeric float64
    GradeLetter  string
    GradedBy     uuid.UUID
    GradedAt     time.Time
    Notes        string
}

func (g *Grade) Validate() error {
    if g == nil {
        return fmt.Errorf("grade is required")
    }
    if g.EnrollmentID == uuid.Nil {
        return fmt.Errorf("enrollment_id is required")
    }
    if g.GradedBy == uuid.Nil {
        return fmt.Errorf("graded_by is required")
    }
    if g.GradeNumeric < 0 || g.GradeNumeric > 10 {
        return fmt.Errorf("grade_numeric must be between 0 and 10")
    }
    g.Notes = strings.TrimSpace(g.Notes)
    return nil
}
