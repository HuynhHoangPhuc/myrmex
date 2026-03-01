package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
)

// StudentHandler proxies admin student CRUD requests to the student gRPC service.
type StudentHandler struct {
	students studentv1.StudentServiceClient
}

func NewStudentHandler(students studentv1.StudentServiceClient) *StudentHandler {
	return &StudentHandler{students: students}
}

func studentToJSON(s *studentv1.Student) gin.H {
	return gin.H{
		"id":              s.Id,
		"student_code":    s.StudentCode,
		"user_id":         s.UserId,
		"full_name":       s.FullName,
		"email":           s.Email,
		"department_id":   s.DepartmentId,
		"enrollment_year": s.EnrollmentYear,
		"status":          s.Status,
		"is_active":       s.IsActive,
		"created_at":      protoTime(s.CreatedAt),
		"updated_at":      protoTime(s.UpdatedAt),
	}
}

func (h *StudentHandler) ListStudents(c *gin.Context) {
	page, pageSize := parsePage(c)
	req := &studentv1.ListStudentsRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	}
	if departmentID := c.Query("department_id"); departmentID != "" {
		req.DepartmentId = &departmentID
	}
	if statusFilter := c.Query("status"); statusFilter != "" {
		req.Status = &statusFilter
	}

	resp, err := h.students.ListStudents(c.Request.Context(), req)
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	data := make([]gin.H, len(resp.Students))
	for i, student := range resp.Students {
		data[i] = studentToJSON(student)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var body struct {
		StudentCode    string `json:"student_code" binding:"required"`
		FullName       string `json:"full_name" binding:"required"`
		Email          string `json:"email" binding:"required,email"`
		DepartmentID   string `json:"department_id" binding:"required"`
		EnrollmentYear int32  `json:"enrollment_year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.students.CreateStudent(c.Request.Context(), &studentv1.CreateStudentRequest{
		StudentCode:    body.StudentCode,
		FullName:       body.FullName,
		Email:          body.Email,
		DepartmentId:   body.DepartmentID,
		EnrollmentYear: body.EnrollmentYear,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, studentToJSON(resp.Student))
}

func (h *StudentHandler) GetStudent(c *gin.Context) {
	resp, err := h.students.GetStudent(c.Request.Context(), &studentv1.GetStudentRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studentToJSON(resp.Student))
}

func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	var body struct {
		FullName     *string `json:"full_name"`
		Email        *string `json:"email"`
		DepartmentID *string `json:"department_id"`
		Status       *string `json:"status"`
		IsActive     *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.students.UpdateStudent(c.Request.Context(), &studentv1.UpdateStudentRequest{
		Id:           c.Param("id"),
		FullName:     body.FullName,
		Email:        body.Email,
		DepartmentId: body.DepartmentID,
		Status:       body.Status,
		IsActive:     body.IsActive,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studentToJSON(resp.Student))
}

func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	_, err := h.students.DeleteStudent(c.Request.Context(), &studentv1.DeleteStudentRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

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
