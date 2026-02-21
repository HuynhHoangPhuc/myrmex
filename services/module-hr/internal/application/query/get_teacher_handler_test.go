package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

func TestGetTeacherHandler(t *testing.T) {
	teacherID := uuid.New()
	teacher := &entity.Teacher{ID: teacherID, FullName: "Alice", Email: "alice@example.com", EmployeeCode: "T001"}

	tests := []struct {
		name      string
		repo      *mockTeacherRepo
		wantErr   bool
		wantSpecs []string
	}{
		{
			name: "success",
			repo: &mockTeacherRepo{
				teacher:         teacher,
				specializations: []string{"math"},
			},
			wantErr:   false,
			wantSpecs: []string{"math"},
		},
		{
			name: "repo error",
			repo: &mockTeacherRepo{
				getErr: errors.New("not found"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewGetTeacherHandler(tt.repo)
			result, err := handler.Handle(context.Background(), GetTeacherQuery{ID: teacherID})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if result == nil {
				t.Fatalf("expected teacher, got nil")
			}
			if len(result.Specializations) != len(tt.wantSpecs) {
				t.Fatalf("Specializations length = %d, want %d", len(result.Specializations), len(tt.wantSpecs))
			}
		})
	}
}
