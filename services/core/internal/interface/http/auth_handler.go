package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myrmex-erp/myrmex/services/core/internal/application/command"
	"github.com/myrmex-erp/myrmex/services/core/internal/application/query"
	"github.com/myrmex-erp/myrmex/services/core/internal/infrastructure/auth"
)

type AuthHandler struct {
	registerHandler *command.RegisterUserHandler
	loginHandler    *query.LoginHandler
	jwtSvc          *auth.JWTService
}

func NewAuthHandler(
	registerHandler *command.RegisterUserHandler,
	loginHandler *query.LoginHandler,
	jwtSvc *auth.JWTService,
) *AuthHandler {
	return &AuthHandler{
		registerHandler: registerHandler,
		loginHandler:    loginHandler,
		jwtSvc:          jwtSvc,
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

	accessToken, _ := h.jwtSvc.GenerateAccessToken(user.ID.String(), string(user.Role))
	refreshToken, _ := h.jwtSvc.GenerateRefreshToken(user.ID.String())

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
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
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

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

	accessToken, _ := h.jwtSvc.GenerateAccessToken(claims.UserID, claims.Role)
	refreshToken, _ := h.jwtSvc.GenerateRefreshToken(claims.UserID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}
