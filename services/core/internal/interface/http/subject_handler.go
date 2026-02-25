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
	c.JSON(http.StatusOK, gin.H{"subjects": resp.Subjects, "pagination": resp.Pagination})
}

// CreateSubject POST /subjects
func (h *SubjectHandler) CreateSubject(c *gin.Context) {
	var body struct {
		Code         string `json:"code" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Credits      int32  `json:"credits" binding:"required"`
		DepartmentID string `json:"department_id" binding:"required"`
		Description  string `json:"description"`
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
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"subject": resp.Subject})
}

// GetSubject GET /subjects/:id
func (h *SubjectHandler) GetSubject(c *gin.Context) {
	resp, err := h.subjects.GetSubject(c.Request.Context(), &subjectv1.GetSubjectRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subject": resp.Subject})
}

// UpdateSubject PATCH /subjects/:id
func (h *SubjectHandler) UpdateSubject(c *gin.Context) {
	var body struct {
		Code         *string `json:"code"`
		Name         *string `json:"name"`
		Credits      *int32  `json:"credits"`
		DepartmentID *string `json:"department_id"`
		Description  *string `json:"description"`
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
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subject": resp.Subject})
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

// ListPrerequisites GET /subjects/:id/prerequisites
func (h *SubjectHandler) ListPrerequisites(c *gin.Context) {
	resp, err := h.prerequisites.ListPrerequisites(c.Request.Context(), &subjectv1.ListPrerequisitesRequest{SubjectId: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"prerequisites": resp.Prerequisites})
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
	c.JSON(http.StatusCreated, gin.H{"prerequisite": resp.Prerequisite})
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
