package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInviteCodeIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "future expiry is not expired",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "past expiry is expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "far future expiry is not expired",
			expiresAt: time.Now().Add(48 * time.Hour),
			want:      false,
		},
		{
			name:      "far past expiry is expired",
			expiresAt: time.Now().Add(-24 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := &InviteCode{
				ID:        uuid.New(),
				CodeHash:  "dummy-hash",
				StudentID: uuid.New(),
				ExpiresAt: tt.expiresAt,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
			}

			got := ic.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInviteCodeIsUsed(t *testing.T) {
	tests := []struct {
		name   string
		usedAt *time.Time
		want   bool
	}{
		{
			name:   "UsedAt is nil returns false",
			usedAt: nil,
			want:   false,
		},
		{
			name:   "UsedAt is set returns true",
			usedAt: &[]time.Time{time.Now()}[0],
			want:   true,
		},
		{
			name:   "past time returns true",
			usedAt: &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := &InviteCode{
				ID:        uuid.New(),
				CodeHash:  "dummy-hash",
				StudentID: uuid.New(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UsedAt:    tt.usedAt,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
			}

			got := ic.IsUsed()
			if got != tt.want {
				t.Errorf("IsUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInviteCodeIsValid(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	tests := []struct {
		name      string
		expiresAt time.Time
		usedAt    *time.Time
		want      bool
	}{
		{
			name:      "not expired and not used is valid",
			expiresAt: now.Add(1 * time.Hour),
			usedAt:    nil,
			want:      true,
		},
		{
			name:      "expired and not used is invalid",
			expiresAt: now.Add(-1 * time.Hour),
			usedAt:    nil,
			want:      false,
		},
		{
			name:      "not expired and used is invalid",
			expiresAt: now.Add(1 * time.Hour),
			usedAt:    &now,
			want:      false,
		},
		{
			name:      "expired and used is invalid",
			expiresAt: now.Add(-1 * time.Hour),
			usedAt:    &now,
			want:      false,
		},
		{
			name:      "far future and not used is valid",
			expiresAt: now.Add(48 * time.Hour),
			usedAt:    nil,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := &InviteCode{
				ID:           uuid.New(),
				CodeHash:     "dummy-hash",
				StudentID:    uuid.New(),
				ExpiresAt:    tt.expiresAt,
				UsedAt:       tt.usedAt,
				UsedByUserID: &userID,
				CreatedBy:    uuid.New(),
				CreatedAt:    now,
			}

			got := ic.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
