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
		prereqType := p.Type
		if prereqType == "" {
			prereqType = "hard"
		}
		prereqs[i] = gin.H{
			"subject_id":        p.SubjectId,
			"prerequisite_id":   p.PrerequisiteId,
			"prerequisite_type": prereqType,
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
	prereqType := resp.Prerequisite.Type
	if prereqType == "" {
		prereqType = "hard"
	}
	c.JSON(http.StatusCreated, gin.H{
		"subject_id":        resp.Prerequisite.SubjectId,
		"prerequisite_id":   resp.Prerequisite.PrerequisiteId,
		"prerequisite_type": prereqType,
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

// FullDAG GET /subjects/dag/full — returns all nodes + edges in one response
func (h *SubjectHandler) FullDAG(c *gin.Context) {
	resp, err := h.prerequisites.GetFullDAG(c.Request.Context(), &subjectv1.GetFullDAGRequest{})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	nodes := make([]gin.H, len(resp.Nodes))
	for i, n := range resp.Nodes {
		nodes[i] = gin.H{
			"id":            n.Id,
			"code":          n.Code,
			"name":          n.Name,
			"credits":       n.Credits,
			"department_id": n.DepartmentId,
			"weekly_hours":  n.WeeklyHours,
			"is_active":     n.IsActive,
		}
	}
	edges := make([]gin.H, len(resp.Edges))
	for i, e := range resp.Edges {
		edges[i] = gin.H{
			"source_id": e.SourceId,
			"target_id": e.TargetId,
			"type":      e.Type,
			"priority":  e.Priority,
		}
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "edges": edges})
}

// CheckConflicts POST /subjects/dag/check-conflicts — finds subjects with missing hard prerequisites
func (h *SubjectHandler) CheckConflicts(c *gin.Context) {
	var body struct {
		SubjectIDs []string `json:"subject_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.prerequisites.CheckPrerequisiteConflicts(c.Request.Context(), &subjectv1.CheckConflictsRequest{
		SubjectIds: body.SubjectIDs,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	conflicts := make([]gin.H, len(resp.Conflicts))
	for i, cd := range resp.Conflicts {
		missing := make([]gin.H, len(cd.Missing))
		for j, m := range cd.Missing {
			missing[j] = gin.H{
				"id":   m.Id,
				"name": m.Name,
				"code": m.Code,
				"type": m.Type,
			}
		}
		conflicts[i] = gin.H{
			"subject_id":   cd.SubjectId,
			"subject_name": cd.SubjectName,
			"missing":      missing,
		}
	}
	c.JSON(http.StatusOK, gin.H{"conflicts": conflicts})
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
