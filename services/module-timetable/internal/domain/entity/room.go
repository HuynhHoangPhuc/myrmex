package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// Room represents a physical teaching room available for scheduling.
type Room struct {
	ID       uuid.UUID
	Name     string
	Capacity int
	Type     string
	Features []string
	IsActive bool
}

func (r *Room) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("room name is required")
	}
	if r.Capacity <= 0 {
		return fmt.Errorf("room capacity must be positive")
	}
	if r.Type == "" {
		r.Type = "classroom"
	}
	return nil
}
