package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCreateStudentHandler(t *testing.T) {
	deptID := uuid.New()

	tests := []struct {
		name    string
		cmd     CreateStudentCommand
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			cmd: CreateStudentCommand{
				StudentCode:    "ST001",
				FullName:       "Ada Lovelace",
				Email:          "ada@example.com",
				DepartmentID:   deptID,
				EnrollmentYear: 2026,
			},
		},
		{
			name: "missing full name",
			cmd: CreateStudentCommand{
				StudentCode:    "ST002",
				Email:          "ada@example.com",
				DepartmentID:   deptID,
				EnrollmentYear: 2026,
			},
			wantErr: true,
		},
		{
			name: "repo error",
			cmd: CreateStudentCommand{
				StudentCode:    "ST003",
				FullName:       "Ada Lovelace",
				Email:          "ada@example.com",
				DepartmentID:   deptID,
				EnrollmentYear: 2026,
			},
			repoErr: errors.New("db"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCreateStudentHandler(&mockStudentRepo{createErr: tt.repoErr}, NewNoopPublisher())
			_, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
