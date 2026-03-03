package entity

import (
	"time"

	"github.com/google/uuid"
)

// InviteCode represents a single-use student registration invite code.
type InviteCode struct {
	ID           uuid.UUID
	CodeHash     string     // SHA-256 hex of plaintext code
	StudentID    uuid.UUID
	ExpiresAt    time.Time
	UsedAt       *time.Time
	UsedByUserID *uuid.UUID
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
}

func (ic *InviteCode) IsExpired() bool { return time.Now().After(ic.ExpiresAt) }
func (ic *InviteCode) IsUsed() bool    { return ic.UsedAt != nil }
func (ic *InviteCode) IsValid() bool   { return !ic.IsExpired() && !ic.IsUsed() }
