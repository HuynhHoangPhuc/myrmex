package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
)

// SubjectHandler proxies HTTP requests to the subject gRPC service.
type SubjectHandler struct {
	subjects      subjectv1.SubjectServiceClient
	prerequisites subjectv1.PrerequisiteServiceClient
}

func NewSubjectHandler(subjects subjectv1.SubjectServiceClient, prerequisites subjectv1.PrerequisiteServiceClient) *SubjectHandler {
	return &SubjectHandler{subjects: subjects, prerequisites: prerequisites}
}

// subjectToJSON converts a proto Subject to a frontend-compatible JSON map.
func subjectToJSON(s *subjectv1.Subject) gin.H {
	return gin.H{
		"id":            s.Id,
		"code":          s.Code,
		"name":          s.Name,
		"credits":       s.Credits,
		"department_id": s.DepartmentId,
		"description":   s.Description,
		"weekly_hours":  s.WeeklyHours,
		"is_active":     s.IsActive,
		"prerequisites": []gin.H{}, // populated separately via /prerequisites endpoint
		"created_at":    protoTime(s.CreatedAt),
		"updated_at":    protoTime(s.UpdatedAt),
	}
}

// ListSubjects GET /subjects?page=&page_size=&department_id=
func (h *SubjectHandler) ListSubjects(c *gin.Context) {
	page, pageSize := parsePage(c)
	req := &subjectv1.ListSubjectsRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	}
	if deptID := c.Query("department_id"); deptID != "" {
		req.DepartmentId = &deptID
	}

	resp, err := h.subjects.ListSubjects(c.Request.Context(), req)
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	data := make([]gin.H, len(resp.Subjects))
	for i, s := range resp.Subjects {
		data[i] = subjectToJSON(s)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateSubject POST /subjects
func (h *SubjectHandler) CreateSubject(c *gin.Context) {
	var body struct {
		Code         string `json:"code" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Credits      int32  `json:"credits" binding:"required"`
		DepartmentID string `json:"department_id" binding:"required"`
		Description  string `json:"description"`
		WeeklyHours  int32  `json:"weekly_hours"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.subjects.CreateSubject(c.Request.Context(), &subjectv1.CreateSubjectRequest{
		Code:         body.Code,
		Name:         body.Name,
		Credits:      body.Credits,
		DepartmentId: body.DepartmentID,
		Description:  body.Description,
		WeeklyHours:  body.WeeklyHours,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, subjectToJSON(resp.Subject))
}

// GetSubject GET /subjects/:id
func (h *SubjectHandler) GetSubject(c *gin.Context) {
	resp, err := h.subjects.GetSubject(c.Request.Context(), &subjectv1.GetSubjectRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subjectToJSON(resp.Subject))
}

// UpdateSubject PATCH /subjects/:id
func (h *SubjectHandler) UpdateSubject(c *gin.Context) {
	var body struct {
		Code         *string `json:"code"`
		Name         *string `json:"name"`
		Credits      *int32  `json:"credits"`
		DepartmentID *string `json:"department_id"`
		Description  *string `json:"description"`
		WeeklyHours  *int32  `json:"weekly_hours"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.subjects.UpdateSubject(c.Request.Context(), &subjectv1.UpdateSubjectRequest{
		Id:           c.Param("id"),
		Code:         body.Code,
		Name:         body.Name,
		Credits:      body.Credits,
		DepartmentId: body.DepartmentID,
		Description:  body.Description,
		WeeklyHours:  body.WeeklyHours,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subjectToJSON(resp.Subject))
}

// DeleteSubject DELETE /subjects/:id
func (h *SubjectHandler) DeleteSubject(c *gin.Context) {
	_, err := h.subjects.DeleteSubject(c.Request.Context(), &subjectv1.DeleteSubjectRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
