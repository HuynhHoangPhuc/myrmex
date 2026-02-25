package command

import (
	"context"
	"errors"
	"testing"
)

func TestCreateSubjectHandler_Success(t *testing.T) {
	repo := &mockSubjectRepo{}
	h := NewCreateSubjectHandler(repo)

	subject, err := h.Handle(context.Background(), CreateSubjectCommand{
		Code:        "CS101",
		Name:        "Intro to CS",
		Credits:     3,
		WeeklyHours: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject == nil {
		t.Fatal("expected subject, got nil")
	}
	if subject.Code != "CS101" {
		t.Fatalf("code mismatch: got %s", subject.Code)
	}
	if !subject.IsActive {
		t.Fatal("expected IsActive=true")
	}
	if repo.created == nil {
		t.Fatal("expected repo.Create to be called")
	}
}

func TestCreateSubjectHandler_MissingCode(t *testing.T) {
	h := NewCreateSubjectHandler(&mockSubjectRepo{})
	_, err := h.Handle(context.Background(), CreateSubjectCommand{
		Code: "", Name: "Intro", Credits: 3,
	})
	if err == nil {
		t.Fatal("expected validation error for missing code")
	}
}

func TestCreateSubjectHandler_MissingName(t *testing.T) {
	h := NewCreateSubjectHandler(&mockSubjectRepo{})
	_, err := h.Handle(context.Background(), CreateSubjectCommand{
		Code: "CS101", Name: "", Credits: 3,
	})
	if err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestCreateSubjectHandler_NegativeCredits(t *testing.T) {
	h := NewCreateSubjectHandler(&mockSubjectRepo{})
	_, err := h.Handle(context.Background(), CreateSubjectCommand{
		Code: "CS101", Name: "Intro", Credits: -1,
	})
	if err == nil {
		t.Fatal("expected validation error for negative credits")
	}
}

func TestCreateSubjectHandler_RepoError(t *testing.T) {
	repo := &mockSubjectRepo{createErr: errors.New("db down")}
	h := NewCreateSubjectHandler(repo)
	_, err := h.Handle(context.Background(), CreateSubjectCommand{
		Code: "CS101", Name: "Intro", Credits: 3,
	})
	if err == nil {
		t.Fatal("expected repo error")
	}
}

func TestUpdateSubjectHandler_Success(t *testing.T) {
	existing := newTestSubject("CS101", "Intro", 3)
	repo := &mockSubjectRepo{getByID: existing}
	h := NewUpdateSubjectHandler(repo)

	newName := "Updated Name"
	result, err := h.Handle(context.Background(), UpdateSubjectCommand{
		ID:   existing.ID,
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Updated Name" {
		t.Fatalf("name not updated: got %s", result.Name)
	}
}

func TestUpdateSubjectHandler_NotFound(t *testing.T) {
	repo := &mockSubjectRepo{getByIDErr: errors.New("not found")}
	h := NewUpdateSubjectHandler(repo)
	_, err := h.Handle(context.Background(), UpdateSubjectCommand{})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestUpdateSubjectHandler_ValidationError(t *testing.T) {
	existing := newTestSubject("CS101", "Intro", 3)
	repo := &mockSubjectRepo{getByID: existing}
	h := NewUpdateSubjectHandler(repo)

	emptyCode := ""
	_, err := h.Handle(context.Background(), UpdateSubjectCommand{
		ID:   existing.ID,
		Code: &emptyCode,
	})
	if err == nil {
		t.Fatal("expected validation error for empty code")
	}
}

func TestUpdateSubjectHandler_RepoUpdateError(t *testing.T) {
	existing := newTestSubject("CS101", "Intro", 3)
	repo := &mockSubjectRepo{getByID: existing, updateErr: errors.New("db error")}
	h := NewUpdateSubjectHandler(repo)

	newName := "New Name"
	_, err := h.Handle(context.Background(), UpdateSubjectCommand{
		ID:   existing.ID,
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected update error")
	}
}

func TestDeleteSubjectHandler_Success(t *testing.T) {
	repo := &mockSubjectRepo{}
	h := NewDeleteSubjectHandler(repo)
	err := h.Handle(context.Background(), DeleteSubjectCommand{ID: newTestSubject("CS101", "Intro", 3).ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSubjectHandler_RepoError(t *testing.T) {
	repo := &mockSubjectRepo{deleteErr: errors.New("db error")}
	h := NewDeleteSubjectHandler(repo)
	err := h.Handle(context.Background(), DeleteSubjectCommand{ID: newTestSubject("CS101", "Intro", 3).ID})
	if err == nil {
		t.Fatal("expected delete error")
	}
}
