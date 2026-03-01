package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
)

func TestListStudentsHandler(t *testing.T) {
	student := &entity.Student{ID: uuid.New(), StudentCode: "ST001", FullName: "Ada", Email: "ada@example.com", DepartmentID: uuid.New(), EnrollmentYear: 2026, Status: entity.StudentStatusActive}

	tests := []struct {
		name       string
		query      ListStudentsQuery
		repo       *mockStudentRepo
		wantErr    bool
		wantTotal  int64
		wantLength int
	}{
		{
			name:  "success",
			query: ListStudentsQuery{Page: 1, PageSize: 10},
			repo: &mockStudentRepo{
				listResults: []*entity.Student{student},
				totalCount:  1,
			},
			wantTotal:  1,
			wantLength: 1,
		},
		{
			name:  "repo list error",
			query: ListStudentsQuery{Page: 1, PageSize: 10},
			repo:  &mockStudentRepo{listErr: errors.New("db")},
			wantErr: true,
		},
		{
			name:  "repo count error",
			query: ListStudentsQuery{Page: 1, PageSize: 10},
			repo: &mockStudentRepo{
				listResults: []*entity.Student{student},
				countErr:    errors.New("count"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewListStudentsHandler(tt.repo)
			result, err := handler.Handle(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if result.Total != tt.wantTotal {
				t.Fatalf("Total = %d, want %d", result.Total, tt.wantTotal)
			}
			if len(result.Students) != tt.wantLength {
				t.Fatalf("Students length = %d, want %d", len(result.Students), tt.wantLength)
			}
		})
	}
}
