package command

import (
	"context"
	"errors"
	"testing"
)

func TestCreateTeacherHandler(t *testing.T) {
	tests := []struct {
		name    string
		cmd     CreateTeacherCommand
		repoErr error
		wantErr bool
	}{
		{
			name:    "success",
			cmd:     CreateTeacherCommand{EmployeeCode: "T001", FullName: "Alice", Email: "alice@example.com", MaxHoursPerWeek: 20},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "missing full name",
			cmd:     CreateTeacherCommand{EmployeeCode: "T002", Email: "bob@example.com"},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "repo error",
			cmd:     CreateTeacherCommand{EmployeeCode: "T003", FullName: "Carol", Email: "carol@example.com", MaxHoursPerWeek: 10},
			repoErr: errors.New("db"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCreateTeacherHandler(&mockTeacherRepo{createErr: tt.repoErr})
			_, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
