package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
)

func TestLoginHandler_Success(t *testing.T) {
	passwordSvc := auth.NewPasswordService()
	hash, err := passwordSvc.Hash("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.User{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		Email:        "user@example.com",
		PasswordHash: hash,
		Role:         valueobject.RoleAdmin,
		IsActive:     true,
	}
	repo := &mockUserRepo{getByEmailUser: user}
	jwtSvc := NewJWTServiceForTest("secret")
	h := NewLoginHandler(repo, jwtSvc, passwordSvc)

	result, err := h.Handle(context.Background(), LoginQuery{Email: "user@example.com", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("expected expiresIn 900, got %d", result.ExpiresIn)
	}
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	passwordSvc := auth.NewPasswordService()
	hash, err := passwordSvc.Hash("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repo := &mockUserRepo{getByEmailUser: &entity.User{PasswordHash: hash, IsActive: true}}
	jwtSvc := NewJWTServiceForTest("secret")
	h := NewLoginHandler(repo, jwtSvc, passwordSvc)

	_, err = h.Handle(context.Background(), LoginQuery{Email: "user@example.com", Password: "wrong"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoginHandler_InactiveAccount(t *testing.T) {
	passwordSvc := auth.NewPasswordService()
	hash, err := passwordSvc.Hash("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repo := &mockUserRepo{getByEmailUser: &entity.User{PasswordHash: hash, IsActive: false}}
	jwtSvc := NewJWTServiceForTest("secret")
	h := NewLoginHandler(repo, jwtSvc, passwordSvc)

	_, err = h.Handle(context.Background(), LoginQuery{Email: "user@example.com", Password: "secret"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoginHandler_RepoError(t *testing.T) {
	repo := &mockUserRepo{getByEmailErr: errors.New("db down")}
	jwtSvc := NewJWTServiceForTest("secret")
	h := NewLoginHandler(repo, jwtSvc, auth.NewPasswordService())

	_, err := h.Handle(context.Background(), LoginQuery{Email: "user@example.com", Password: "secret"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func NewJWTServiceForTest(secret string) *auth.JWTService {
	return auth.NewJWTService(secret, 15*time.Minute, 7*24*time.Hour)
}
