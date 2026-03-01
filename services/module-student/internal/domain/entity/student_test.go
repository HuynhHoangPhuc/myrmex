package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestStudentValidate(t *testing.T) {
	valid := Student{
		StudentCode:    "ST-001",
		FullName:       "Ada Lovelace",
		Email:          "ada@example.com",
		DepartmentID:   uuid.New(),
		EnrollmentYear: 2026,
		Status:         StudentStatusActive,
	}

	tests := []struct {
		name    string
		student Student
		wantErr bool
	}{
		{name: "valid", student: valid, wantErr: false},
		{name: "missing student code", student: Student{FullName: "Ada", Email: "ada@example.com", DepartmentID: uuid.New(), EnrollmentYear: 2026, Status: StudentStatusActive}, wantErr: true},
		{name: "invalid year", student: Student{StudentCode: "ST-001", FullName: "Ada", Email: "ada@example.com", DepartmentID: uuid.New(), EnrollmentYear: 1999, Status: StudentStatusActive}, wantErr: true},
		{name: "invalid status", student: Student{StudentCode: "ST-001", FullName: "Ada", Email: "ada@example.com", DepartmentID: uuid.New(), EnrollmentYear: 2026, Status: "unknown"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.student.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
