package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         valueobject.Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if !u.Role.IsValid() {
		return fmt.Errorf("invalid role: %s", u.Role)
	}
	return nil
}

func (u *User) CanLogin() bool {
	return u.IsActive && u.PasswordHash != ""
}
