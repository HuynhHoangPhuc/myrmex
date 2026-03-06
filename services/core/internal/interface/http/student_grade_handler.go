package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
)

// gradeToJSON converts a proto Grade to a frontend-compatible JSON map.
func gradeToJSON(g *studentv1.Grade) gin.H {
	return gin.H{
		"id":            g.Id,
		"enrollment_id": g.EnrollmentId,
		"grade_numeric": g.GradeNumeric,
		"grade_letter":  g.GradeLetter,
		"graded_by":     g.GradedBy,
		"graded_at":     protoTime(g.GradedAt),
		"notes":         g.Notes,
	}
}

// GetStudentTranscript GET /students/:id/transcript
func (h *StudentHandler) GetStudentTranscript(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.students.GetStudentTranscript(c.Request.Context(), &studentv1.GetStudentTranscriptRequest{
		StudentId: id,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GenerateInviteCode generates a single-use invite code for a student (admin only).
func (h *StudentHandler) GenerateInviteCode(c *gin.Context) {
	createdBy, _ := c.Get("user_id")
	resp, err := h.students.CreateInviteCode(c.Request.Context(), &studentv1.CreateInviteCodeRequest{
		StudentId: c.Param("id"),
		CreatedBy: createdBy.(string),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":       resp.Code,
		"student_id": resp.StudentId,
		"expires_at": resp.ExpiresAt,
	})
}

// AssignGrade assigns a numeric grade to an approved enrollment (admin/teacher only).
func (h *StudentHandler) AssignGrade(c *gin.Context) {
	var body struct {
		EnrollmentID string  `json:"enrollment_id" binding:"required"`
		GradeNumeric float64 `json:"grade_numeric" binding:"required,min=0,max=10"`
		Notes        string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gradedBy, _ := c.Get("user_id")
	resp, err := h.students.AssignGrade(c.Request.Context(), &studentv1.AssignGradeRequest{
		EnrollmentId: body.EnrollmentID,
		GradeNumeric: body.GradeNumeric,
		GradedBy:     gradedBy.(string),
		Notes:        body.Notes,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gradeToJSON(resp.Grade))
}

// UpdateGrade updates an existing grade (admin/teacher only).
func (h *StudentHandler) UpdateGrade(c *gin.Context) {
	var body struct {
		GradeNumeric float64 `json:"grade_numeric" binding:"required,min=0,max=10"`
		Notes        string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gradedBy, _ := c.Get("user_id")
	resp, err := h.students.UpdateGrade(c.Request.Context(), &studentv1.UpdateGradeRequest{
		GradeId:      c.Param("id"),
		GradeNumeric: body.GradeNumeric,
		GradedBy:     gradedBy.(string),
		Notes:        body.Notes,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gradeToJSON(resp.Grade))
}
