package command

import (
	"context"
	"errors"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
)

func TestRegisterUserHandler_Success(t *testing.T) {
	repo := &mockUserRepo{}
	store := &mockEventStore{}
	pwdSvc := auth.NewPasswordService()
	h := NewRegisterUserHandler(repo, store, nil, pwdSvc)

	user, err := h.Handle(context.Background(), RegisterUserCommand{
		Email: "alice@example.com", Password: "secret", FullName: "Alice", Role: "admin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("email mismatch")
	}
	if user.PasswordHash == "secret" {
		t.Error("password must be hashed")
	}
	if repo.created == nil {
		t.Fatal("expected user repo create")
	}
}

func TestRegisterUserHandler_ValidationFailure(t *testing.T) {
	h := NewRegisterUserHandler(&mockUserRepo{}, &mockEventStore{}, nil, auth.NewPasswordService())
	_, err := h.Handle(context.Background(), RegisterUserCommand{
		Email: "", Password: "x", FullName: "", Role: "admin",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegisterUserHandler_InvalidRole(t *testing.T) {
	h := NewRegisterUserHandler(&mockUserRepo{}, &mockEventStore{}, nil, auth.NewPasswordService())
	_, err := h.Handle(context.Background(), RegisterUserCommand{
		Email: "a@b.com", Password: "x", FullName: "A", Role: "superadmin",
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestRegisterUserHandler_RepoError(t *testing.T) {
	repo := &mockUserRepo{createErr: errors.New("db down")}
	h := NewRegisterUserHandler(repo, &mockEventStore{}, nil, auth.NewPasswordService())
	_, err := h.Handle(context.Background(), RegisterUserCommand{
		Email: "a@b.com", Password: "x", FullName: "A", Role: "admin",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestRegisterUserHandler_EventStoreError_StillSucceeds(t *testing.T) {
	repo := &mockUserRepo{}
	store := &mockEventStore{appendErr: errors.New("event store down")}
	h := NewRegisterUserHandler(repo, store, nil, auth.NewPasswordService())
	user, err := h.Handle(context.Background(), RegisterUserCommand{
		Email: "b@c.com", Password: "pw", FullName: "Bob", Role: "viewer",
	})
	if err != nil {
		t.Fatalf("should succeed despite event store error: %v", err)
	}
	if user == nil {
		t.Error("expected user")
	}
}
