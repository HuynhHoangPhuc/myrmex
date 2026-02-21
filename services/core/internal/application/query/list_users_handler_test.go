package query

import (
	"context"
	"errors"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

func TestListUsersHandler_DefaultPaging(t *testing.T) {
	repo := &mockUserRepo{
		listUsers: []*entity.User{{Email: "a@example.com"}},
		countTotal: 1,
	}
	h := NewListUsersHandler(repo)

	result, err := h.Handle(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastListLimit != 20 {
		t.Errorf("expected default limit 20, got %d", repo.lastListLimit)
	}
	if repo.lastListOffset != 0 {
		t.Errorf("expected default offset 0, got %d", repo.lastListOffset)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Users) != 1 {
		t.Fatalf("expected one user")
	}
}

func TestListUsersHandler_CustomPaging(t *testing.T) {
	repo := &mockUserRepo{listUsers: []*entity.User{}}
	h := NewListUsersHandler(repo)

	_, err := h.Handle(context.Background(), ListUsersQuery{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastListLimit != 5 {
		t.Errorf("expected limit 5, got %d", repo.lastListLimit)
	}
	if repo.lastListOffset != 5 {
		t.Errorf("expected offset 5, got %d", repo.lastListOffset)
	}
}

func TestListUsersHandler_ListError(t *testing.T) {
	repo := &mockUserRepo{listErr: errors.New("list failed")}
	h := NewListUsersHandler(repo)

	_, err := h.Handle(context.Background(), ListUsersQuery{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListUsersHandler_CountError(t *testing.T) {
	repo := &mockUserRepo{countErr: errors.New("count failed")}
	h := NewListUsersHandler(repo)

	_, err := h.Handle(context.Background(), ListUsersQuery{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatal("expected error")
	}
}
