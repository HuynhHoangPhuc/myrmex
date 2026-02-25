package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

func TestListSubjectsHandler_Success_AllSubjects(t *testing.T) {
	subj := &entity.Subject{ID: uuid.New(), Code: "CS101", Name: "Intro", Credits: 3, IsActive: true}
	repo := &mockSubjectRepo{
		listAll:    []*entity.Subject{subj},
		totalCount: 1,
	}
	h := NewListSubjectsHandler(repo)

	result, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(result.Subjects))
	}
	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
}

func TestListSubjectsHandler_DefaultLimit(t *testing.T) {
	repo := &mockSubjectRepo{listAll: []*entity.Subject{}}
	h := NewListSubjectsHandler(repo)

	// limit=0 should default to 20
	result, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestListSubjectsHandler_FilterByDepartment(t *testing.T) {
	subj := &entity.Subject{ID: uuid.New(), Code: "CS101", Name: "Intro", Credits: 3, DepartmentID: "eng"}
	repo := &mockSubjectRepo{
		listByDept: []*entity.Subject{subj},
		deptCount:  1,
	}
	h := NewListSubjectsHandler(repo)

	result, err := h.Handle(context.Background(), ListSubjectsQuery{
		Limit: 10, DepartmentID: "eng",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Subjects) != 1 {
		t.Fatalf("expected 1 subject by dept, got %d", len(result.Subjects))
	}
	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
}

func TestListSubjectsHandler_ListError(t *testing.T) {
	repo := &mockSubjectRepo{listAllErr: errors.New("db down")}
	h := NewListSubjectsHandler(repo)

	_, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 10})
	if err == nil {
		t.Fatal("expected error from list")
	}
}

func TestListSubjectsHandler_CountError(t *testing.T) {
	repo := &mockSubjectRepo{
		listAll:  []*entity.Subject{},
		countErr: errors.New("count failed"),
	}
	h := NewListSubjectsHandler(repo)

	_, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 10})
	if err == nil {
		t.Fatal("expected error from count")
	}
}

func TestListSubjectsHandler_DeptListError(t *testing.T) {
	repo := &mockSubjectRepo{listByDeptErr: errors.New("dept list failed")}
	h := NewListSubjectsHandler(repo)

	_, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 10, DepartmentID: "eng"})
	if err == nil {
		t.Fatal("expected error from dept list")
	}
}

func TestListSubjectsHandler_DeptCountError(t *testing.T) {
	repo := &mockSubjectRepo{
		listByDept:   []*entity.Subject{},
		deptCountErr: errors.New("dept count failed"),
	}
	h := NewListSubjectsHandler(repo)

	_, err := h.Handle(context.Background(), ListSubjectsQuery{Limit: 10, DepartmentID: "eng"})
	if err == nil {
		t.Fatal("expected error from dept count")
	}
}

func TestGetSubjectHandler_Success(t *testing.T) {
	subj := &entity.Subject{ID: uuid.New(), Code: "CS101", Name: "Intro", Credits: 3}
	repo := &mockSubjectRepo{getByID: subj}
	h := NewGetSubjectHandler(repo)

	result, err := h.Handle(context.Background(), GetSubjectQuery{ID: subj.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != "CS101" {
		t.Fatalf("code mismatch: got %s", result.Code)
	}
}

func TestGetSubjectHandler_NotFound(t *testing.T) {
	repo := &mockSubjectRepo{getByIDErr: errNotFound}
	h := NewGetSubjectHandler(repo)

	_, err := h.Handle(context.Background(), GetSubjectQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}
