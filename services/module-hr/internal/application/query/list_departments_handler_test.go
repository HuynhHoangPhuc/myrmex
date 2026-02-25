package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
)

// mockDeptRepo implements repository.DepartmentRepository for testing.
type mockDeptRepo struct {
	depts    []*entity.Department
	listErr  error
	total    int64
	countErr error
}

func (m *mockDeptRepo) Create(_ context.Context, d *entity.Department) (*entity.Department, error) {
	return d, nil
}

func (m *mockDeptRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Department, error) {
	return nil, nil
}

func (m *mockDeptRepo) List(_ context.Context, _, _ int32) ([]*entity.Department, error) {
	return m.depts, m.listErr
}

func (m *mockDeptRepo) Count(_ context.Context) (int64, error) {
	return m.total, m.countErr
}

func TestListDepartmentsHandler_Success(t *testing.T) {
	dept := &entity.Department{ID: uuid.New(), Name: "CS", Code: "CS"}
	repo := &mockDeptRepo{depts: []*entity.Department{dept}, total: 1}
	h := NewListDepartmentsHandler(repo)

	result, err := h.Handle(context.Background(), ListDepartmentsQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Departments) != 1 {
		t.Fatalf("expected 1 department, got %d", len(result.Departments))
	}
	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
}

func TestListDepartmentsHandler_DefaultPagination(t *testing.T) {
	repo := &mockDeptRepo{depts: []*entity.Department{}}
	h := NewListDepartmentsHandler(repo)

	// page=0 and pageSize=0 should use defaults (page=1, pageSize=20)
	result, err := h.Handle(context.Background(), ListDepartmentsQuery{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestListDepartmentsHandler_ListError(t *testing.T) {
	repo := &mockDeptRepo{listErr: errors.New("db down")}
	h := NewListDepartmentsHandler(repo)

	_, err := h.Handle(context.Background(), ListDepartmentsQuery{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatal("expected error from list")
	}
}

func TestListDepartmentsHandler_CountError(t *testing.T) {
	repo := &mockDeptRepo{depts: []*entity.Department{}, countErr: errors.New("count failed")}
	h := NewListDepartmentsHandler(repo)

	_, err := h.Handle(context.Background(), ListDepartmentsQuery{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatal("expected error from count")
	}
}

func TestListDepartmentsHandler_EmptyResult(t *testing.T) {
	repo := &mockDeptRepo{depts: []*entity.Department{}, total: 0}
	h := NewListDepartmentsHandler(repo)

	result, err := h.Handle(context.Background(), ListDepartmentsQuery{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Departments) != 0 {
		t.Fatalf("expected 0 departments, got %d", len(result.Departments))
	}
}
