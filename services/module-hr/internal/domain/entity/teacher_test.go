package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestTeacher_Validate(t *testing.T) {
	departmentID := uuid.New()

	tests := []struct {
		name    string
		teacher Teacher
		wantErr bool
		wantMax int
	}{
		{
			name:    "valid",
			teacher: Teacher{FullName: "Alice", Email: "alice@example.com", EmployeeCode: "T001", MaxHoursPerWeek: 20, DepartmentID: &departmentID},
			wantErr: false,
			wantMax: 20,
		},
		{
			name:    "zero max defaults to 20",
			teacher: Teacher{FullName: "Bob", Email: "bob@example.com", EmployeeCode: "T002", MaxHoursPerWeek: 0},
			wantErr: false,
			wantMax: 20,
		},
		{
			name:    "missing full name",
			teacher: Teacher{Email: "charlie@example.com", EmployeeCode: "T003"},
			wantErr: true,
			wantMax: 0,
		},
		{
			name:    "missing email",
			teacher: Teacher{FullName: "Dana", EmployeeCode: "T004"},
			wantErr: true,
			wantMax: 0,
		},
		{
			name:    "missing employee code",
			teacher: Teacher{FullName: "Evan", Email: "evan@example.com"},
			wantErr: true,
			wantMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.teacher.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.teacher.MaxHoursPerWeek != tt.wantMax {
				t.Fatalf("MaxHoursPerWeek = %d, want %d", tt.teacher.MaxHoursPerWeek, tt.wantMax)
			}
		})
	}
}
