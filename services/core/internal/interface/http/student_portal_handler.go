package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
)

// StudentPortalHandler serves student self-service portal routes (/api/student/*).
// All handlers resolve the student record from the JWT user_id via ResolveStudentMiddleware.
type StudentPortalHandler struct {
	students studentv1.StudentServiceClient
}

func NewStudentPortalHandler(students studentv1.StudentServiceClient) *StudentPortalHandler {
	return &StudentPortalHandler{students: students}
}

// ResolveStudentMiddleware resolves the student_id from the JWT user_id once per request.
// Stores student_id and student proto in gin context for downstream handlers.
func (h *StudentPortalHandler) ResolveStudentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		resp, err := h.students.GetStudentByUserID(c.Request.Context(),
			&studentv1.GetStudentByUserIDRequest{UserId: userID.(string)})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no student profile linked — contact admin"})
			c.Abort()
			return
		}
		c.Set("student_id", resp.Student.Id)
		c.Set("student", resp.Student)
		c.Next()
	}
}

func (h *StudentPortalHandler) GetMyProfile(c *gin.Context) {
	student, _ := c.Get("student")
	c.JSON(http.StatusOK, studentToJSON(student.(*studentv1.Student)))
}

func (h *StudentPortalHandler) ListMyEnrollments(c *gin.Context) {
	studentID, _ := c.Get("student_id")
	req := &studentv1.GetStudentEnrollmentsRequest{StudentId: studentID.(string)}
	if semesterID := c.Query("semester_id"); semesterID != "" {
		req.SemesterId = &semesterID
	}
	resp, err := h.students.GetStudentEnrollments(c.Request.Context(), req)
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	enrollments := make([]gin.H, len(resp.Enrollments))
	for i, e := range resp.Enrollments {
		enrollments[i] = enrollmentToJSON(e)
	}
	c.JSON(http.StatusOK, gin.H{"enrollments": enrollments})
}

func (h *StudentPortalHandler) RequestMyEnrollment(c *gin.Context) {
	studentID, _ := c.Get("student_id")
	var body struct {
		SemesterID        string `json:"semester_id" binding:"required"`
		OfferedSubjectID  string `json:"offered_subject_id" binding:"required"`
		SubjectID         string `json:"subject_id" binding:"required"`
		RequestNote       string `json:"request_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.students.RequestEnrollment(c.Request.Context(), &studentv1.RequestEnrollmentRequest{
		StudentId:        studentID.(string),
		SemesterId:       body.SemesterID,
		OfferedSubjectId: body.OfferedSubjectID,
		SubjectId:        body.SubjectID,
		RequestNote:      body.RequestNote,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, enrollmentToJSON(resp.Enrollment))
}

func (h *StudentPortalHandler) CheckMyPrerequisites(c *gin.Context) {
	studentID, _ := c.Get("student_id")
	subjectID := c.Query("subject_id")
	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject_id is required"})
		return
	}
	resp, err := h.students.CheckPrerequisites(c.Request.Context(), &studentv1.CheckPrerequisitesRequest{
		StudentId: studentID.(string),
		SubjectId: subjectID,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"can_enroll": resp.CanEnroll,
		"missing":    resp.Missing,
	})
}

func (h *StudentPortalHandler) GetMyTranscript(c *gin.Context) {
	studentID, _ := c.Get("student_id")
	resp, err := h.students.GetStudentTranscript(c.Request.Context(), &studentv1.GetStudentTranscriptRequest{
		StudentId: studentID.(string),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ExportTranscript is stubbed — PDF export deferred to a future phase.
func (h *StudentPortalHandler) ExportTranscript(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "PDF export not yet implemented"})
}
