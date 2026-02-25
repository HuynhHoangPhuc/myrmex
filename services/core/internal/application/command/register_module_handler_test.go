package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
)

type mockModuleRepo struct {
	registered *entity.ModuleRegistration
	registerErr error
}

func (m *mockModuleRepo) Register(_ context.Context, mod *entity.ModuleRegistration) (*entity.ModuleRegistration, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	mod.ID = uuid.New()
	m.registered = mod
	return mod, nil
}

func (m *mockModuleRepo) Unregister(_ context.Context, _ string) error { return nil }
func (m *mockModuleRepo) GetByName(_ context.Context, _ string) (*entity.ModuleRegistration, error) {
	return nil, nil
}
func (m *mockModuleRepo) List(_ context.Context) ([]*entity.ModuleRegistration, error) {
	return nil, nil
}
func (m *mockModuleRepo) UpdateHealth(_ context.Context, _ string, _ entity.HealthStatus) error {
	return nil
}

func TestRegisterModuleHandler_Success(t *testing.T) {
	repo := &mockModuleRepo{}
	h := NewRegisterModuleHandler(repo)

	result, err := h.Handle(context.Background(), RegisterModuleCommand{
		Name:        "hr",
		Version:     "v1.0",
		GRPCAddress: "localhost:50051",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Name != "hr" {
		t.Fatalf("name mismatch: got %s", result.Name)
	}
	if repo.registered == nil {
		t.Fatal("expected repo.Register to be called")
	}
}

func TestRegisterModuleHandler_ValidationFailure_MissingName(t *testing.T) {
	h := NewRegisterModuleHandler(&mockModuleRepo{})
	_, err := h.Handle(context.Background(), RegisterModuleCommand{
		Name:        "",
		Version:     "v1.0",
		GRPCAddress: "localhost:50051",
	})
	if err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestRegisterModuleHandler_ValidationFailure_MissingVersion(t *testing.T) {
	h := NewRegisterModuleHandler(&mockModuleRepo{})
	_, err := h.Handle(context.Background(), RegisterModuleCommand{
		Name:        "hr",
		Version:     "",
		GRPCAddress: "localhost:50051",
	})
	if err == nil {
		t.Fatal("expected validation error for missing version")
	}
}

func TestRegisterModuleHandler_ValidationFailure_MissingAddress(t *testing.T) {
	h := NewRegisterModuleHandler(&mockModuleRepo{})
	_, err := h.Handle(context.Background(), RegisterModuleCommand{
		Name:        "hr",
		Version:     "v1.0",
		GRPCAddress: "",
	})
	if err == nil {
		t.Fatal("expected validation error for missing grpc address")
	}
}

func TestRegisterModuleHandler_RepoError(t *testing.T) {
	repo := &mockModuleRepo{registerErr: errors.New("db down")}
	h := NewRegisterModuleHandler(repo)
	_, err := h.Handle(context.Background(), RegisterModuleCommand{
		Name:        "hr",
		Version:     "v1.0",
		GRPCAddress: "localhost:50051",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
