package command

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteUserHandler_Success(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewDeleteUserHandler(repo)
	id := mustUUID(20)

	err := h.Handle(context.Background(), DeleteUserCommand{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedID != id {
		t.Errorf("expected delete id to be recorded")
	}
}

func TestDeleteUserHandler_RepoError(t *testing.T) {
	repo := &mockUserRepo{deleteErr: errors.New("delete failed")}
	h := NewDeleteUserHandler(repo)

	err := h.Handle(context.Background(), DeleteUserCommand{ID: mustUUID(21)})
	if err == nil {
		t.Fatal("expected error")
	}
}
