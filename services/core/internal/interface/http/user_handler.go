package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/myrmex-erp/myrmex/services/core/internal/application/command"
	"github.com/myrmex-erp/myrmex/services/core/internal/application/query"
	"github.com/myrmex-erp/myrmex/services/core/internal/domain/entity"
)

type UserHandler struct {
	getUserHandler    *query.GetUserHandler
	listUsersHandler  *query.ListUsersHandler
	updateUserHandler *command.UpdateUserHandler
	deleteUserHandler *command.DeleteUserHandler
}

func NewUserHandler(
	getUser *query.GetUserHandler,
	listUsers *query.ListUsersHandler,
	updateUser *command.UpdateUserHandler,
	deleteUser *command.DeleteUserHandler,
) *UserHandler {
	return &UserHandler{
		getUserHandler:    getUser,
		listUsersHandler:  listUsers,
		updateUserHandler: updateUser,
		deleteUserHandler: deleteUser,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.getUserHandler.Handle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, userResponse(user))
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var q struct {
		Page     int32 `form:"page"`
		PageSize int32 `form:"page_size"`
	}
	_ = c.ShouldBindQuery(&q)

	result, err := h.listUsersHandler.Handle(c.Request.Context(), query.ListUsersQuery{
		Page: q.Page, PageSize: q.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users := make([]gin.H, len(result.Users))
	for i, u := range result.Users {
		users[i] = userResponse(u)
	}
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": result.Total,
		"page":  q.Page,
	})
}

type updateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := command.UpdateUserCommand{
		ID:       id,
		FullName: req.FullName,
		Email:    req.Email,
		Role:     req.Role,
	}
	if req.IsActive != nil {
		cmd.IsActive = *req.IsActive
	}

	user, err := h.updateUserHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userResponse(user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.deleteUserHandler.Handle(c.Request.Context(), command.DeleteUserCommand{ID: id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func userResponse(u *entity.User) gin.H {
	return gin.H{
		"id":         u.ID.String(),
		"email":      u.Email,
		"full_name":  u.FullName,
		"role":       string(u.Role),
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}
