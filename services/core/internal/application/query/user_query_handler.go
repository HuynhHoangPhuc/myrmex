package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/core/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/core/internal/domain/repository"
	"github.com/myrmex-erp/myrmex/services/core/internal/infrastructure/auth"
)

type GetUserHandler struct {
	userRepo repository.UserRepository
}

func NewGetUserHandler(userRepo repository.UserRepository) *GetUserHandler {
	return &GetUserHandler{userRepo: userRepo}
}

func (h *GetUserHandler) Handle(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return h.userRepo.GetByID(ctx, id)
}

type ListUsersQuery struct {
	Page     int32
	PageSize int32
}

type ListUsersResult struct {
	Users []*entity.User
	Total int64
}

type ListUsersHandler struct {
	userRepo repository.UserRepository
}

func NewListUsersHandler(userRepo repository.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{userRepo: userRepo}
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	users, err := h.userRepo.List(ctx, q.PageSize, offset)
	if err != nil {
		return nil, err
	}
	total, err := h.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &ListUsersResult{Users: users, Total: total}, nil
}

type LoginQuery struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type LoginHandler struct {
	userRepo    repository.UserRepository
	jwtSvc      *auth.JWTService
	passwordSvc *auth.PasswordService
}

func NewLoginHandler(userRepo repository.UserRepository, jwtSvc *auth.JWTService, passwordSvc *auth.PasswordService) *LoginHandler {
	return &LoginHandler{userRepo: userRepo, jwtSvc: jwtSvc, passwordSvc: passwordSvc}
}

func (h *LoginHandler) Handle(ctx context.Context, q LoginQuery) (*LoginResult, error) {
	user, err := h.userRepo.GetByEmail(ctx, q.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !user.CanLogin() {
		return nil, fmt.Errorf("account is inactive")
	}
	if err := h.passwordSvc.Verify(user.PasswordHash, q.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, err := h.jwtSvc.GenerateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err := h.jwtSvc.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}
