package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
)

// HRHandler proxies HTTP requests to the HR gRPC microservice.
type HRHandler struct {
	teachers    hrv1.TeacherServiceClient
	departments hrv1.DepartmentServiceClient
}

func NewHRHandler(teachers hrv1.TeacherServiceClient, departments hrv1.DepartmentServiceClient) *HRHandler {
	return &HRHandler{teachers: teachers, departments: departments}
}

// parsePage reads ?page= and ?page_size= query params with safe defaults and upper bound.
func parsePage(c *gin.Context) (int32, int32) {
	page := int32(1)
	pageSize := int32(20)
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = int32(v)
	}
	if v, err := strconv.Atoi(c.Query("page_size")); err == nil && v > 0 {
		pageSize = int32(v)
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// --- Teacher handlers ---

func (h *HRHandler) ListTeachers(c *gin.Context) {
	page, pageSize := parsePage(c)
	req := &hrv1.ListTeachersRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	}
	if deptID := c.Query("department_id"); deptID != "" {
		req.DepartmentId = &deptID
	}

	resp, err := h.teachers.ListTeachers(c.Request.Context(), req)
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teachers": resp.Teachers, "pagination": resp.Pagination})
}

func (h *HRHandler) CreateTeacher(c *gin.Context) {
	var body struct {
		FullName     string `json:"full_name" binding:"required"`
		Email        string `json:"email" binding:"required"`
		DepartmentId string `json:"department_id" binding:"required"`
		Title        string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.teachers.CreateTeacher(c.Request.Context(), &hrv1.CreateTeacherRequest{
		FullName:     body.FullName,
		Email:        body.Email,
		DepartmentId: body.DepartmentId,
		Title:        body.Title,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"teacher": resp.Teacher})
}

func (h *HRHandler) GetTeacher(c *gin.Context) {
	resp, err := h.teachers.GetTeacher(c.Request.Context(), &hrv1.GetTeacherRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teacher": resp.Teacher})
}

func (h *HRHandler) UpdateTeacher(c *gin.Context) {
	var body struct {
		FullName     *string `json:"full_name"`
		Email        *string `json:"email"`
		DepartmentId *string `json:"department_id"`
		Title        *string `json:"title"`
		IsActive     *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.teachers.UpdateTeacher(c.Request.Context(), &hrv1.UpdateTeacherRequest{
		Id:           c.Param("id"),
		FullName:     body.FullName,
		Email:        body.Email,
		DepartmentId: body.DepartmentId,
		Title:        body.Title,
		IsActive:     body.IsActive,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teacher": resp.Teacher})
}

func (h *HRHandler) DeleteTeacher(c *gin.Context) {
	_, err := h.teachers.DeleteTeacher(c.Request.Context(), &hrv1.DeleteTeacherRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HRHandler) GetTeacherAvailability(c *gin.Context) {
	resp, err := h.teachers.ListTeacherAvailability(c.Request.Context(), &hrv1.ListTeacherAvailabilityRequest{
		TeacherId: c.Param("id"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teacher_id": resp.TeacherId, "available_slots": resp.AvailableSlots})
}

func (h *HRHandler) UpdateTeacherAvailability(c *gin.Context) {
	var body struct {
		AvailableSlots []*hrv1.TimeSlot `json:"available_slots" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.teachers.UpdateTeacherAvailability(c.Request.Context(), &hrv1.UpdateTeacherAvailabilityRequest{
		TeacherId:      c.Param("id"),
		AvailableSlots: body.AvailableSlots,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teacher_id": resp.TeacherId, "available_slots": resp.AvailableSlots})
}

// --- Department handlers ---

func (h *HRHandler) ListDepartments(c *gin.Context) {
	page, pageSize := parsePage(c)
	resp, err := h.departments.ListDepartments(c.Request.Context(), &hrv1.ListDepartmentsRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"departments": resp.Departments, "pagination": resp.Pagination})
}

func (h *HRHandler) CreateDepartment(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.departments.CreateDepartment(c.Request.Context(), &hrv1.CreateDepartmentRequest{
		Name: body.Name,
		Code: body.Code,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"department": resp.Department})
}
