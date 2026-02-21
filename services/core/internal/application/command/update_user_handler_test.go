package command

import (
	"context"
	"errors"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
)

func TestUpdateUserHandler_SuccessWithRole(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewUpdateUserHandler(repo)
	userID := mustUUID(10)

	user, err := h.Handle(context.Background(), UpdateUserCommand{
		ID: userID, FullName: "Alice", Email: "alice@example.com", Role: "admin", IsActive: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected update to be called")
	}
	if repo.updated.ID != userID {
		t.Errorf("id mismatch")
	}
	if repo.updated.Role != valueobject.RoleAdmin {
		t.Errorf("role mismatch: got %q", repo.updated.Role)
	}
	if user == nil {
		t.Fatal("expected user")
	}
}

func TestUpdateUserHandler_SuccessWithoutRole(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewUpdateUserHandler(repo)
	userID := mustUUID(11)

	_, err := h.Handle(context.Background(), UpdateUserCommand{
		ID: userID, FullName: "Bob", Email: "bob@example.com", Role: "", IsActive: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected update to be called")
	}
	if repo.updated.Role != "" {
		t.Errorf("expected empty role")
	}
}

func TestUpdateUserHandler_RepoError(t *testing.T) {
	repo := &mockUserRepo{updateErr: errors.New("update failed")}
	h := NewUpdateUserHandler(repo)

	_, err := h.Handle(context.Background(), UpdateUserCommand{
		ID: mustUUID(12), FullName: "Chris", Email: "chris@example.com", Role: "admin", IsActive: true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
