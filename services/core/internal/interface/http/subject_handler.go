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

// ListPrerequisites GET /subjects/:id/prerequisites — returns array directly
func (h *SubjectHandler) ListPrerequisites(c *gin.Context) {
	resp, err := h.prerequisites.ListPrerequisites(c.Request.Context(), &subjectv1.ListPrerequisitesRequest{SubjectId: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	prereqs := make([]gin.H, len(resp.Prerequisites))
	for i, p := range resp.Prerequisites {
		prereqs[i] = gin.H{
			"subject_id":       p.SubjectId,
			"prerequisite_id":  p.PrerequisiteId,
			"prerequisite_type": "hard", // default; proto doesn't carry type yet
		}
	}
	c.JSON(http.StatusOK, prereqs)
}

// AddPrerequisite POST /subjects/:id/prerequisites
func (h *SubjectHandler) AddPrerequisite(c *gin.Context) {
	var body struct {
		PrerequisiteID string `json:"prerequisite_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.prerequisites.AddPrerequisite(c.Request.Context(), &subjectv1.AddPrerequisiteRequest{
		SubjectId:      c.Param("id"),
		PrerequisiteId: body.PrerequisiteID,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"subject_id":        resp.Prerequisite.SubjectId,
		"prerequisite_id":   resp.Prerequisite.PrerequisiteId,
		"prerequisite_type": "hard",
	})
}

// RemovePrerequisite DELETE /subjects/:id/prerequisites/:prereqId
func (h *SubjectHandler) RemovePrerequisite(c *gin.Context) {
	_, err := h.prerequisites.RemovePrerequisite(c.Request.Context(), &subjectv1.RemovePrerequisiteRequest{
		SubjectId:      c.Param("id"),
		PrerequisiteId: c.Param("prereqId"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ValidateDAG GET /subjects/dag/validate
func (h *SubjectHandler) ValidateDAG(c *gin.Context) {
	resp, err := h.prerequisites.ValidateDAG(c.Request.Context(), &subjectv1.ValidateDAGRequest{})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_valid": resp.IsValid, "cycle_path": resp.CyclePath})
}

// TopologicalSort GET /subjects/dag/topological-sort
func (h *SubjectHandler) TopologicalSort(c *gin.Context) {
	resp, err := h.prerequisites.TopologicalSort(c.Request.Context(), &subjectv1.TopologicalSortRequest{})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subject_ids": resp.SubjectIds})
}
