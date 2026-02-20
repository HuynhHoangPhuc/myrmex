package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Department is a simple lookup entity grouping teachers.
type Department struct {
	ID        uuid.UUID
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks required fields.
func (d *Department) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}
	if d.Code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}
