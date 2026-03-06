package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
)

// buildSubjectMap fetches all subjects and returns a map of ID → Subject proto (best-effort).
func (h *SubjectHandler) buildSubjectMap(ctx context.Context) map[string]*subjectv1.Subject {
	resp, err := h.subjects.ListSubjects(ctx, &subjectv1.ListSubjectsRequest{
		Pagination: &corev1.PaginationRequest{Page: 1, PageSize: 500},
	})
	if err != nil {
		return nil
	}
	m := make(map[string]*subjectv1.Subject, len(resp.Subjects))
	for _, s := range resp.Subjects {
		m[s.Id] = s
	}
	return m
}

// ListPrerequisites GET /subjects/:id/prerequisites — returns array with enriched subject names
func (h *SubjectHandler) ListPrerequisites(c *gin.Context) {
	resp, err := h.prerequisites.ListPrerequisites(c.Request.Context(), &subjectv1.ListPrerequisitesRequest{SubjectId: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Resolve UUIDs to names so the AI agent can present meaningful information
	subjectMap := h.buildSubjectMap(c.Request.Context())

	prereqs := make([]gin.H, len(resp.Prerequisites))
	for i, p := range resp.Prerequisites {
		prereqType := p.Type
		if prereqType == "" {
			prereqType = "hard"
		}
		entry := gin.H{
			"subject_id":        p.SubjectId,
			"prerequisite_id":   p.PrerequisiteId,
			"prerequisite_type": prereqType,
		}
		if s, ok := subjectMap[p.SubjectId]; ok {
			entry["subject_name"] = s.Name
			entry["subject_code"] = s.Code
		}
		if prereq, ok := subjectMap[p.PrerequisiteId]; ok {
			entry["prerequisite_name"] = prereq.Name
			entry["prerequisite_code"] = prereq.Code
		}
		prereqs[i] = entry
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
