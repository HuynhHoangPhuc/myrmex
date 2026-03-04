package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
)

// AdminRoleHandler provides the role management API for admins and super_admins.
type AdminRoleHandler struct {
	updateRoleHandler *command.UpdateUserRoleHandler
}

func NewAdminRoleHandler(updateRoleHandler *command.UpdateUserRoleHandler) *AdminRoleHandler {
	return &AdminRoleHandler{updateRoleHandler: updateRoleHandler}
}

type updateRoleRequest struct {
	Role         string `json:"role" binding:"required"`
	DepartmentID string `json:"department_id"` // optional; required for dept_head/teacher roles
}

// UpdateUserRole handles PATCH /api/users/:id/role.
// Requires admin or super_admin role.
func (h *AdminRoleHandler) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := command.UpdateUserRoleCommand{
		ID:   id,
		Role: req.Role,
	}
	if req.DepartmentID != "" {
		deptID, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department_id"})
			return
		}
		cmd.DepartmentID = &deptID
	}

	user, err := h.updateRoleHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		"id":        user.ID.String(),
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      string(user.Role),
	}
	if user.DepartmentID != nil {
		resp["department_id"] = user.DepartmentID.String()
	}
	c.JSON(http.StatusOK, resp)
}
