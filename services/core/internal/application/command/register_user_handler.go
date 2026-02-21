package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/valueobject"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"

	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/pkg/eventstore"
	pkgnats "github.com/HuynhHoangPhuc/myrmex/pkg/nats"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	FullName string
	Role     string
}

type RegisterUserHandler struct {
	userRepo    repository.UserRepository
	eventStore  eventstore.EventStore
	publisher   *pkgnats.Publisher
	passwordSvc *auth.PasswordService
}

func NewRegisterUserHandler(
	userRepo repository.UserRepository,
	eventStore eventstore.EventStore,
	publisher *pkgnats.Publisher,
	passwordSvc *auth.PasswordService,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:    userRepo,
		eventStore:  eventStore,
		publisher:   publisher,
		passwordSvc: passwordSvc,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*entity.User, error) {
	role, err := valueobject.ParseRole(cmd.Role)
	if err != nil {
		return nil, err
	}

	hash, err := h.passwordSvc.Hash(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		Email:        cmd.Email,
		PasswordHash: hash,
		FullName:     cmd.FullName,
		Role:         role,
		IsActive:     true,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}

	created, err := h.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Append event
	eventData, _ := json.Marshal(map[string]string{
		"email": created.Email, "full_name": created.FullName, "role": string(created.Role),
	})
	_ = h.eventStore.Append(ctx, created.ID, 0, []eventstore.Event{{
		AggregateID:   created.ID,
		AggregateType: "user",
		EventType:     "user.created",
		Data:          eventData,
	}})

	// Publish to NATS
	if h.publisher != nil {
		_ = h.publisher.Publish(ctx, "core.user.created", eventData)
	}

	return created, nil
}

type UpdateUserCommand struct {
	ID       uuid.UUID
	FullName string
	Email    string
	Role     string
	IsActive bool
}

type UpdateUserHandler struct {
	userRepo repository.UserRepository
}

func NewUpdateUserHandler(userRepo repository.UserRepository) *UpdateUserHandler {
	return &UpdateUserHandler{userRepo: userRepo}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (*entity.User, error) {
	user := &entity.User{
		ID:       cmd.ID,
		FullName: cmd.FullName,
		Email:    cmd.Email,
		IsActive: cmd.IsActive,
	}
	if cmd.Role != "" {
		role, err := valueobject.ParseRole(cmd.Role)
		if err != nil {
			return nil, err
		}
		user.Role = role
	}
	return h.userRepo.Update(ctx, user)
}

type DeleteUserCommand struct {
	ID uuid.UUID
}

type DeleteUserHandler struct {
	userRepo repository.UserRepository
}

func NewDeleteUserHandler(userRepo repository.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{userRepo: userRepo}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) error {
	return h.userRepo.Delete(ctx, cmd.ID)
}
