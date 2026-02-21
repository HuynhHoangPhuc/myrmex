package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

func TestListTeachersHandler(t *testing.T) {
	departmentID := uuid.New()
	teacher := &entity.Teacher{ID: uuid.New(), FullName: "Alice", Email: "alice@example.com", EmployeeCode: "T001"}

	tests := []struct {
		name       string
		query      ListTeachersQuery
		repo       *mockTeacherRepo
		wantErr    bool
		wantTotal  int64
		wantLength int
	}{
		{
			name:  "success",
			query: ListTeachersQuery{Page: 1, PageSize: 10},
			repo: &mockTeacherRepo{
				listAll:    []*entity.Teacher{teacher},
				totalCount: 1,
			},
			wantErr:    false,
			wantTotal:  1,
			wantLength: 1,
		},
		{
			name:  "empty result",
			query: ListTeachersQuery{Page: 1, PageSize: 10, DepartmentID: &departmentID},
			repo: &mockTeacherRepo{
				listByDepartment: []*entity.Teacher{},
				totalCount:       0,
			},
			wantErr:    false,
			wantTotal:  0,
			wantLength: 0,
		},
		{
			name:  "repo list error",
			query: ListTeachersQuery{Page: 1, PageSize: 10},
			repo: &mockTeacherRepo{
				listAllErr: errors.New("db"),
			},
			wantErr: true,
		},
		{
			name:  "repo count error",
			query: ListTeachersQuery{Page: 1, PageSize: 10},
			repo: &mockTeacherRepo{
				listAll:  []*entity.Teacher{teacher},
				countErr: errors.New("count"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewListTeachersHandler(tt.repo)
			result, err := handler.Handle(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if result == nil {
				t.Fatalf("expected result, got nil")
			}
			if result.Total != tt.wantTotal {
				t.Fatalf("Total = %d, want %d", result.Total, tt.wantTotal)
			}
			if len(result.Teachers) != tt.wantLength {
				t.Fatalf("Teachers length = %d, want %d", len(result.Teachers), tt.wantLength)
			}
		})
	}
}
