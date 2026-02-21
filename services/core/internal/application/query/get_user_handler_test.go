package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

func TestGetUserHandler_Success(t *testing.T) {
	user := &entity.User{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Email: "a@example.com"}
	repo := &mockUserRepo{getUser: user}
	h := NewGetUserHandler(repo)

	result, err := h.Handle(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != user {
		t.Fatalf("expected user result")
	}
}

func TestGetUserHandler_RepoError(t *testing.T) {
	repo := &mockUserRepo{getErr: errors.New("not found")}
	h := NewGetUserHandler(repo)

	_, err := h.Handle(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000002"))
	if err == nil {
		t.Fatal("expected error")
	}
}
