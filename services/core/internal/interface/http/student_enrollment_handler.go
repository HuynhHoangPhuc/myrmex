package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
)

// enrollmentToJSON converts a proto EnrollmentRequest to a frontend-compatible JSON map.
func enrollmentToJSON(e *studentv1.EnrollmentRequest) gin.H {
	return gin.H{
		"id":                 e.Id,
		"student_id":         e.StudentId,
		"semester_id":        e.SemesterId,
		"offered_subject_id": e.OfferedSubjectId,
		"subject_id":         e.SubjectId,
		"status":             e.Status,
		"request_note":       e.RequestNote,
		"admin_note":         e.AdminNote,
		"requested_at":       protoTime(e.RequestedAt),
		"reviewed_at":        protoTime(e.ReviewedAt),
		"reviewed_by":        e.ReviewedBy,
	}
}

// ListEnrollments GET /enrollments?page=&page_size=&student_id=&semester_id=&status=&subject_id=
func (h *StudentHandler) ListEnrollments(c *gin.Context) {
	page, pageSize := parsePage(c)
	subjectIDFilter := c.Query("subject_id") // client-side filter (proto doesn't support it)

	req := &studentv1.ListEnrollmentRequestsRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	}
	// When filtering by subject, fetch a large batch so client-side count is accurate
	if subjectIDFilter != "" {
		req.Pagination = &corev1.PaginationRequest{Page: 1, PageSize: 500}
	}
	if studentID := c.Query("student_id"); studentID != "" {
		req.StudentId = &studentID
	}
	if semesterID := c.Query("semester_id"); semesterID != "" {
		req.SemesterId = &semesterID
	}
	if statusFilter := c.Query("status"); statusFilter != "" {
		req.Status = &statusFilter
	}

	resp, err := h.students.ListEnrollmentRequests(c.Request.Context(), req)
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	data := make([]gin.H, len(resp.Enrollments))
	for i, e := range resp.Enrollments {
		data[i] = enrollmentToJSON(e)
	}

	// Apply client-side subject_id filter
	if subjectIDFilter != "" {
		filtered := make([]gin.H, 0, len(data))
		for _, entry := range data {
			if sid, _ := entry["subject_id"].(string); sid == subjectIDFilter {
				filtered = append(filtered, entry)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"data":      filtered,
			"total":     int64(len(filtered)),
			"page":      1,
			"page_size": len(filtered),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
}

// ReviewEnrollment PATCH /enrollments/:id/review
func (h *StudentHandler) ReviewEnrollment(c *gin.Context) {
	var body struct {
		Approve   bool   `json:"approve"`
		AdminNote string `json:"admin_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reviewedBy, _ := c.Get("user_id")
	reviewedByStr, _ := reviewedBy.(string)

	resp, err := h.students.ReviewEnrollment(c.Request.Context(), &studentv1.ReviewEnrollmentRequest{
		Id:         c.Param("id"),
		Approve:    body.Approve,
		AdminNote:  body.AdminNote,
		ReviewedBy: reviewedByStr,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollmentToJSON(resp.Enrollment))
}
