package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestHashInviteCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantLen  int // SHA-256 hex is 64 chars
		wantHash string
	}{
		{
			name:    "deterministic hash same input same output",
			code:    "abc123def456",
			wantLen: 64,
		},
		{
			name:    "different inputs produce different hashes",
			code:    "xyz789",
			wantLen: 64,
		},
		{
			name:    "empty string produces hash",
			code:    "",
			wantLen: 64,
		},
		{
			name:    "32-char hex code produces hash",
			code:    "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
			wantLen: 64,
		},
	}

	// Test determinism: same input should always produce same hash
	firstRun := make(map[string]string)
	secondRun := make(map[string]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashInviteCode(tt.code)

			// Verify hash length
			if len(hash) != tt.wantLen {
				t.Errorf("HashInviteCode length = %d, want %d", len(hash), tt.wantLen)
			}

			// Verify hash is valid hex string (only 0-9, a-f)
			for _, ch := range hash {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
					t.Errorf("HashInviteCode contains non-hex character: %c", ch)
				}
			}

			// Store for determinism check
			firstRun[tt.code] = hash
		})
	}

	// Run again and verify determinism
	for _, tt := range tests {
		hash := HashInviteCode(tt.code)
		secondRun[tt.code] = hash
	}

	for code, hash1 := range firstRun {
		hash2 := secondRun[code]
		if hash1 != hash2 {
			t.Errorf("HashInviteCode not deterministic for code %q: got %s, then %s", code, hash1, hash2)
		}
	}

	// Verify different inputs produce different hashes
	hash1 := HashInviteCode("code1")
	hash2 := HashInviteCode("code2")
	if hash1 == hash2 {
		t.Errorf("Different inputs produced same hash: %s", hash1)
	}
}

func TestCreateInviteCodeHandler(t *testing.T) {
	studentID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name        string
		cmd         CreateInviteCodeCommand
		studentErr  error
		invalidateErr error
		createErr   error
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "success",
			cmd: CreateInviteCodeCommand{
				StudentID:    studentID,
				CreatedBy:    createdBy,
				ExpiresHours: 48,
			},
			wantErr: false,
		},
		{
			name: "student not found",
			cmd: CreateInviteCodeCommand{
				StudentID:    studentID,
				CreatedBy:    createdBy,
				ExpiresHours: 48,
			},
			studentErr: errors.New("not found"),
			wantErr:    true,
		},
		{
			name: "invalidate existing codes fails",
			cmd: CreateInviteCodeCommand{
				StudentID:    studentID,
				CreatedBy:    createdBy,
				ExpiresHours: 48,
			},
			invalidateErr: errors.New("db error"),
			wantErr:       true,
		},
		{
			name: "persist invite code fails",
			cmd: CreateInviteCodeCommand{
				StudentID:    studentID,
				CreatedBy:    createdBy,
				ExpiresHours: 48,
			},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStudentRepo := &mockStudentRepoForInviteCode{getErr: tt.studentErr}
			mockInviteCodeRepo := &mockInviteCodeRepoForCreate{
				invalidateErr: tt.invalidateErr,
				createErr:     tt.createErr,
			}

			handler := NewCreateInviteCodeHandler(
				mockStudentRepo,
				mockInviteCodeRepo,
				NewNoopPublisher(),
			)

			result, err := handler.Handle(context.Background(), tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				// Verify result structure
				if result.StudentID != studentID {
					t.Errorf("result StudentID = %v, want %v", result.StudentID, studentID)
				}
				if len(result.Code) != 32 {
					t.Errorf("result Code length = %d, want 32", len(result.Code))
				}
				// Verify code is hex string
				for _, ch := range result.Code {
					if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
						t.Errorf("Code contains non-hex character: %c", ch)
					}
				}
				if result.ExpiresAt == "" {
					t.Errorf("result ExpiresAt is empty")
				}
			}
		})
	}
}
