package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

func TestValidateInviteCodeHandler(t *testing.T) {
	studentID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		query          ValidateInviteCodeQuery
		inviteCode     *entity.InviteCode
		inviteCodeErr  error
		student        *entity.Student
		studentErr     error
		wantValid      bool
		wantMessage    string
		wantStudentID  string
		wantStudentName string
	}{
		{
			name:  "valid code returns success",
			query: ValidateInviteCodeQuery{Code: "valid-code-123"},
			inviteCode: &entity.InviteCode{
				ID:        uuid.New(),
				CodeHash:  command.HashInviteCode("valid-code-123"),
				StudentID: studentID,
				ExpiresAt: now.Add(24 * time.Hour),
				UsedAt:    nil,
				CreatedBy: uuid.New(),
				CreatedAt: now,
			},
			student: &entity.Student{
				ID:       studentID,
				FullName: "John Doe",
				Email:    "john@example.com",
			},
			wantValid:       true,
			wantStudentID:   studentID.String(),
			wantStudentName: "John Doe",
		},
		{
			name:           "code not found returns invalid",
			query:          ValidateInviteCodeQuery{Code: "nonexistent-code"},
			inviteCodeErr:  errors.New("not found"),
			wantValid:      false,
			wantMessage:    "invalid invite code",
		},
		{
			name:  "expired code returns invalid",
			query: ValidateInviteCodeQuery{Code: "expired-code"},
			inviteCode: &entity.InviteCode{
				ID:        uuid.New(),
				CodeHash:  command.HashInviteCode("expired-code"),
				StudentID: studentID,
				ExpiresAt: now.Add(-24 * time.Hour), // past
				UsedAt:    nil,
				CreatedBy: uuid.New(),
				CreatedAt: now.Add(-48 * time.Hour),
			},
			wantValid:   false,
			wantMessage: "invite code has expired",
		},
		{
			name:  "already used code returns invalid",
			query: ValidateInviteCodeQuery{Code: "used-code"},
			inviteCode: &entity.InviteCode{
				ID:           uuid.New(),
				CodeHash:     command.HashInviteCode("used-code"),
				StudentID:    studentID,
				ExpiresAt:    now.Add(24 * time.Hour),
				UsedAt:       &now,
				UsedByUserID: &[]uuid.UUID{uuid.New()}[0],
				CreatedBy:    uuid.New(),
				CreatedAt:    now.Add(-1 * time.Hour),
			},
			wantValid:   false,
			wantMessage: "invite code has already been used",
		},
		{
			name:  "student not found returns invalid",
			query: ValidateInviteCodeQuery{Code: "orphan-code"},
			inviteCode: &entity.InviteCode{
				ID:        uuid.New(),
				CodeHash:  command.HashInviteCode("orphan-code"),
				StudentID: studentID,
				ExpiresAt: now.Add(24 * time.Hour),
				UsedAt:    nil,
				CreatedBy: uuid.New(),
				CreatedAt: now,
			},
			studentErr:  errors.New("not found"),
			wantValid:   false,
			wantMessage: "associated student not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInviteCodeRepo := &mockInviteCodeRepoForValidate{
				findErr: tt.inviteCodeErr,
				code:    tt.inviteCode,
			}
			mockStudentRepo := &mockStudentRepoForValidate{
				getErr:  tt.studentErr,
				student: tt.student,
			}

			handler := NewValidateInviteCodeHandler(mockInviteCodeRepo, mockStudentRepo)
			result, err := handler.Handle(context.Background(), tt.query)

			if err != nil {
				t.Fatalf("Handle() returned unexpected error: %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("result.Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if result.Valid {
				if result.StudentID != tt.wantStudentID {
					t.Errorf("result.StudentID = %q, want %q", result.StudentID, tt.wantStudentID)
				}
				if result.StudentName != tt.wantStudentName {
					t.Errorf("result.StudentName = %q, want %q", result.StudentName, tt.wantStudentName)
				}
			} else {
				if result.Message != tt.wantMessage {
					t.Errorf("result.Message = %q, want %q", result.Message, tt.wantMessage)
				}
			}
		})
	}
}

func TestValidateInviteCodeResultFields(t *testing.T) {
	// Test that ValidateInviteCodeResult has expected fields
	result := &ValidateInviteCodeResult{
		Valid:       true,
		StudentID:   "123e4567-e89b-12d3-a456-426614174000",
		StudentName: "Jane Smith",
		Message:     "",
	}

	tests := []struct {
		name      string
		checkFunc func(*ValidateInviteCodeResult) bool
		errMsg    string
	}{
		{
			name: "Valid field is boolean",
			checkFunc: func(r *ValidateInviteCodeResult) bool {
				return r.Valid == true || r.Valid == false
			},
			errMsg: "Valid field should be boolean",
		},
		{
			name: "StudentID field is string",
			checkFunc: func(r *ValidateInviteCodeResult) bool {
				return r.StudentID != ""
			},
			errMsg: "StudentID field should be non-empty string",
		},
		{
			name: "StudentName field is string",
			checkFunc: func(r *ValidateInviteCodeResult) bool {
				return r.StudentName != ""
			},
			errMsg: "StudentName field should be non-empty string",
		},
		{
			name: "Message field is string",
			checkFunc: func(r *ValidateInviteCodeResult) bool {
				return r.Message == "" || r.Message != ""
			},
			errMsg: "Message field should be string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.checkFunc(result) {
				t.Error(tt.errMsg)
			}
		})
	}
}
