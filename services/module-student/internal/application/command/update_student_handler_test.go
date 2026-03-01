package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

func TestUpdateStudentHandler(t *testing.T) {
	studentID := uuid.New()
	deptID := uuid.New()

	tests := []struct {
		name    string
		cmd     UpdateStudentCommand
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			cmd: UpdateStudentCommand{
				ID:             studentID,
				StudentCode:    "ST001",
				FullName:       "Updated",
				Email:          "updated@example.com",
				DepartmentID:   deptID,
				EnrollmentYear: 2026,
				Status:         entity.StudentStatusSuspended,
				IsActive:       true,
			},
		},
		{
			name: "repo error",
			cmd: UpdateStudentCommand{
				ID:             studentID,
				StudentCode:    "ST001",
				FullName:       "Updated",
				Email:          "updated@example.com",
				DepartmentID:   deptID,
				EnrollmentYear: 2026,
				Status:         entity.StudentStatusActive,
				IsActive:       true,
			},
			repoErr: errors.New("db"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewUpdateStudentHandler(&mockStudentRepo{updateErr: tt.repoErr}, NewNoopPublisher())
			_, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
