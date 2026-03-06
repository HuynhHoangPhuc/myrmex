package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
)

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

	// Resolve IDs to subject objects so the AI agent can present meaningful names
	subjectMap := h.buildSubjectMap(c.Request.Context())
	subjects := make([]gin.H, 0, len(resp.SubjectIds))
	for _, id := range resp.SubjectIds {
		if s, ok := subjectMap[id]; ok {
			subjects = append(subjects, gin.H{
				"id":      s.Id,
				"code":    s.Code,
				"name":    s.Name,
				"credits": s.Credits,
			})
		} else {
			subjects = append(subjects, gin.H{"id": id})
		}
	}
	c.JSON(http.StatusOK, gin.H{"subject_ids": resp.SubjectIds, "subjects": subjects})
}
