package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
)

type AuthHandler struct {
	registerHandler   *command.RegisterUserHandler
	loginHandler      *query.LoginHandler
	jwtSvc            *auth.JWTService
	userRepo          repository.UserRepository   // for refresh token re-validation
	updateUserHandler *command.UpdateUserHandler
	deleteUserHandler *command.DeleteUserHandler
	studentClient     studentv1.StudentServiceClient // nil if student module not configured
}

func NewAuthHandler(
	registerHandler *command.RegisterUserHandler,
	loginHandler *query.LoginHandler,
	jwtSvc *auth.JWTService,
	userRepo repository.UserRepository,
	updateUserHandler *command.UpdateUserHandler,
	deleteUserHandler *command.DeleteUserHandler,
	studentClient studentv1.StudentServiceClient,
) *AuthHandler {
	return &AuthHandler{
		registerHandler:   registerHandler,
		loginHandler:      loginHandler,
		jwtSvc:            jwtSvc,
		userRepo:          userRepo,
		updateUserHandler: updateUserHandler,
		deleteUserHandler: deleteUserHandler,
		studentClient:     studentClient,
	}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.registerHandler.Handle(c.Request.Context(), command.RegisterUserCommand{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     "viewer",
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// New users have no dept scope yet
	accessToken, _ := h.jwtSvc.GenerateAccessToken(user.ID.String(), string(user.Role), "", "")
	refreshToken, _ := h.jwtSvc.GenerateRefreshToken(user.ID.String())

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":         user.ID.String(),
			"email":      user.Email,
			"full_name":  user.FullName,
			"role":       string(user.Role),
			"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.loginHandler.Handle(c.Request.Context(), query.LoginQuery{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
		"user": gin.H{
			"id":         result.UserID,
			"email":      result.Email,
			"full_name":  result.FullName,
			"role":       result.Role,
			"created_at": result.CreatedAt,
		},
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh validates the refresh token, re-fetches the user from DB (to get current
// role and dept scope), then issues a new access + refresh token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.jwtSvc.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Re-fetch user to get current role and department scope (refresh token only carries userID)
	ctx := c.Request.Context()
	userID, parseErr := uuid.Parse(claims.UserID)
	if parseErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or inactive"})
		return
	}

	deptID, teacherID := query.ResolveTokenClaims(ctx, user, h.userRepo)

	accessToken, _ := h.jwtSvc.GenerateAccessToken(user.ID.String(), string(user.Role), deptID, teacherID)
	refreshToken, _ := h.jwtSvc.GenerateRefreshToken(user.ID.String())

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

type registerStudentRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FullName   string `json:"full_name" binding:"required"`
	InviteCode string `json:"invite_code" binding:"required"`
}

// RegisterStudent creates a user account and links it to a student record via invite code.
// Flow: validate code → create user (viewer) → redeem code (link + upgrade to student).
// Rollback: deletes user if code redemption fails.
func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	if h.studentClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "student service not available"})
		return
	}

	var req registerStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Step 1: validate invite code before creating any user.
	validateResp, err := h.studentClient.ValidateInviteCode(ctx, &studentv1.ValidateInviteCodeRequest{
		Code: req.InviteCode,
	})
	if err != nil || !validateResp.Valid {
		msg := "invalid invite code"
		if validateResp != nil && validateResp.Message != "" {
			msg = validateResp.Message
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Step 2: create user with viewer role (safe default before redemption).
	user, err := h.registerHandler.Handle(ctx, command.RegisterUserCommand{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     "viewer",
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Step 3: atomically redeem code — links user to student record.
	_, redeemErr := h.studentClient.RedeemInviteCode(ctx, &studentv1.RedeemInviteCodeRequest{
		Code:   req.InviteCode,
		UserId: user.ID.String(),
	})
	if redeemErr != nil {
		// Rollback: delete the viewer user to avoid orphaned accounts.
		_ = h.deleteUserHandler.Handle(ctx, command.DeleteUserCommand{ID: user.ID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite code could not be redeemed: " + redeemErr.Error()})
		return
	}

	// Step 4: upgrade user role to student.
	updated, err := h.updateUserHandler.Handle(ctx, command.UpdateUserCommand{
		ID:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     "student",
		IsActive: true,
	})
	if err != nil {
		// Non-fatal: user exists as viewer, code was claimed. Log but continue.
		updated = user
	}

	// Students have no dept scope
	accessToken, _ := h.jwtSvc.GenerateAccessToken(updated.ID.String(), string(updated.Role), "", "")
	refreshToken, _ := h.jwtSvc.GenerateRefreshToken(updated.ID.String())

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":         updated.ID.String(),
			"email":      updated.Email,
			"full_name":  updated.FullName,
			"role":       string(updated.Role),
			"created_at": updated.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
