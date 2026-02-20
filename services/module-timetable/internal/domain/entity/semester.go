package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Semester is the aggregate root for a teaching semester.
type Semester struct {
	ID                uuid.UUID
	Name              string
	Year              int
	Term              int
	StartDate         time.Time
	EndDate           time.Time
	OfferedSubjectIDs []uuid.UUID
	CreatedAt         time.Time
}

func (s *Semester) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("semester name is required")
	}
	if s.Year < 2000 || s.Year > 2100 {
		return fmt.Errorf("invalid year: %d", s.Year)
	}
	if s.Term < 1 || s.Term > 3 {
		return fmt.Errorf("term must be 1, 2 or 3")
	}
	if !s.EndDate.After(s.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}
	return nil
}

func (s *Semester) AddOfferedSubject(id uuid.UUID) {
	for _, existing := range s.OfferedSubjectIDs {
		if existing == id {
			return
		}
	}
	s.OfferedSubjectIDs = append(s.OfferedSubjectIDs, id)
}

func (s *Semester) RemoveOfferedSubject(id uuid.UUID) {
	filtered := s.OfferedSubjectIDs[:0]
	for _, existing := range s.OfferedSubjectIDs {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	s.OfferedSubjectIDs = filtered
}
