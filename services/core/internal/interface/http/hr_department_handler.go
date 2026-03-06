package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
)

// departmentToJSON converts a proto Department to a frontend-compatible JSON map.
func departmentToJSON(d *hrv1.Department) gin.H {
	return gin.H{
		"id":         d.Id,
		"name":       d.Name,
		"code":       d.Code,
		"created_at": protoTime(d.CreatedAt),
		"updated_at": protoTime(d.UpdatedAt),
	}
}

// ListDepartments GET /departments?page=&page_size=
func (h *HRHandler) ListDepartments(c *gin.Context) {
	page, pageSize := parsePage(c)
	resp, err := h.departments.ListDepartments(c.Request.Context(), &hrv1.ListDepartmentsRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	data := make([]gin.H, len(resp.Departments))
	for i, d := range resp.Departments {
		data[i] = departmentToJSON(d)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateDepartment POST /departments
func (h *HRHandler) CreateDepartment(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.departments.CreateDepartment(c.Request.Context(), &hrv1.CreateDepartmentRequest{
		Name: body.Name,
		Code: body.Code,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, departmentToJSON(resp.Department))
}
