package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

func TestGetStudentHandler(t *testing.T) {
	studentID := uuid.New()
	student := &entity.Student{ID: studentID, StudentCode: "ST001", FullName: "Ada", Email: "ada@example.com", DepartmentID: uuid.New(), EnrollmentYear: 2026, Status: entity.StudentStatusActive}

	tests := []struct {
		name    string
		repo    *mockStudentRepo
		wantErr bool
	}{
		{name: "success", repo: &mockStudentRepo{student: student}},
		{name: "repo error", repo: &mockStudentRepo{getErr: errors.New("not found")}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewGetStudentHandler(tt.repo)
			result, err := handler.Handle(context.Background(), GetStudentQuery{ID: studentID})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if result == nil {
				t.Fatalf("expected student, got nil")
			}
		})
	}
}
