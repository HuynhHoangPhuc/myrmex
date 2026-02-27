package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// protoTime converts a proto Timestamp to an RFC3339 string, returns "" if nil.
func protoTime(ts interface{ AsTime() time.Time }) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

// teacherToJSON converts a proto Teacher to a frontend-compatible JSON map.
// dept may be nil — in that case the department field is omitted from the response.
func teacherToJSON(t *hrv1.Teacher, dept *hrv1.Department) gin.H {
	specs := t.Specializations
	if specs == nil {
		specs = []string{}
	}
	var deptJSON interface{}
	if dept != nil {
		deptJSON = gin.H{"id": dept.Id, "name": dept.Name, "code": dept.Code}
	}
	return gin.H{
		"id":                 t.Id,
		"employee_code":      t.EmployeeCode,
		"full_name":          t.FullName,
		"email":              t.Email,
		"phone":              t.Phone,
		"title":              t.Title,
		"department_id":      t.DepartmentId,
		"department":         deptJSON,
		"max_hours_per_week": t.MaxHoursPerWeek,
		"is_active":          t.IsActive,
		"specializations":    specs,
		"created_at":         protoTime(t.CreatedAt),
		"updated_at":         protoTime(t.UpdatedAt),
	}
}

// hrAvailSlots maps 1-indexed period numbers to frontend 2-hour time strings.
// Matches the 6 slots in teacher-availability-form.tsx: 07–09, 09–11, ... 17–19.
var hrAvailSlots = []struct{ start, end string }{
	{"07:00", "09:00"},
	{"09:00", "11:00"},
	{"11:00", "13:00"},
	{"13:00", "15:00"},
	{"15:00", "17:00"},
	{"17:00", "19:00"},
}

// hrSlotStart returns the start time string for a 1-indexed HR availability period.
func hrSlotStart(p int32) string {
	if p >= 1 && int(p) <= len(hrAvailSlots) {
		return hrAvailSlots[p-1].start
	}
	return ""
}

// hrSlotEnd returns the end time string for a 1-indexed HR availability period.
func hrSlotEnd(p int32) string {
	if p >= 1 && int(p) <= len(hrAvailSlots) {
		return hrAvailSlots[p-1].end
	}
	return ""
}

// hrTimeToSlot converts a start time string (e.g. "07:00") to a 1-indexed period.
// Returns 0 if not found.
func hrTimeToSlot(t string) int32 {
	for i, slot := range hrAvailSlots {
		if slot.start == t {
			return int32(i + 1)
		}
	}
	return 0
}

// availabilityToJSON converts proto TimeSlot list to frontend [{day_of_week, start_time, end_time}] format.
func availabilityToJSON(slots []*hrv1.TimeSlot) []gin.H {
	result := make([]gin.H, 0, len(slots))
	for _, s := range slots {
		start := hrSlotStart(s.StartPeriod)
		end := hrSlotEnd(s.StartPeriod) // end of this slot = start of next slot
		if start == "" {
			continue
		}
		result = append(result, gin.H{
			"day_of_week": s.DayOfWeek,
			"start_time":  start,
			"end_time":    end,
		})
	}
	return result
}

// departmentToJSON converts a proto Department to a frontend-compatible JSON map.
func departmentToJSON(d *hrv1.Department) gin.H {
	return gin.H{
		"id":         d.Id,
		"name":       d.Name,
		"code":       d.Code,
		"created_at": protoTime(d.CreatedAt),
		"updated_at": protoTime(d.UpdatedAt),
	}
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

	// Fetch all departments once and build a lookup map to enrich teacher responses.
	deptMap := h.fetchDepartmentMap(c)

	data := make([]gin.H, len(resp.Teachers))
	for i, t := range resp.Teachers {
		data[i] = teacherToJSON(t, deptMap[t.DepartmentId])
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
}

// fetchDepartmentMap returns a map of department ID → Department proto for inline enrichment.
// Errors are silently swallowed so teacher listing still succeeds without department names.
func (h *HRHandler) fetchDepartmentMap(c *gin.Context) map[string]*hrv1.Department {
	deptResp, err := h.departments.ListDepartments(c.Request.Context(), &hrv1.ListDepartmentsRequest{
		Pagination: &corev1.PaginationRequest{Page: 1, PageSize: 200},
	})
	if err != nil {
		return nil
	}
	m := make(map[string]*hrv1.Department, len(deptResp.Departments))
	for _, d := range deptResp.Departments {
		m[d.Id] = d
	}
	return m
}

func (h *HRHandler) CreateTeacher(c *gin.Context) {
	var body struct {
		FullName        string   `json:"full_name" binding:"required"`
		Email           string   `json:"email" binding:"required"`
		DepartmentId    string   `json:"department_id" binding:"required"`
		Title           string   `json:"title" binding:"required"`
		EmployeeCode    string   `json:"employee_code"`
		MaxHoursPerWeek int32    `json:"max_hours_per_week"`
		Specializations []string `json:"specializations"`
		Phone           string   `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.teachers.CreateTeacher(c.Request.Context(), &hrv1.CreateTeacherRequest{
		FullName:        body.FullName,
		Email:           body.Email,
		DepartmentId:    body.DepartmentId,
		Title:           body.Title,
		EmployeeCode:    body.EmployeeCode,
		MaxHoursPerWeek: body.MaxHoursPerWeek,
		Specializations: body.Specializations,
		Phone:           body.Phone,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	deptMap := h.fetchDepartmentMap(c)
	c.JSON(http.StatusCreated, teacherToJSON(resp.Teacher, deptMap[resp.Teacher.DepartmentId]))
}

func (h *HRHandler) GetTeacher(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.teachers.GetTeacher(c.Request.Context(), &hrv1.GetTeacherRequest{Id: id})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	deptMap := h.fetchDepartmentMap(c)
	result := teacherToJSON(resp.Teacher, deptMap[resp.Teacher.DepartmentId])

	// Enrich with availability slots (best-effort — empty slice on error)
	availResp, _ := h.teachers.ListTeacherAvailability(c.Request.Context(), &hrv1.ListTeacherAvailabilityRequest{TeacherId: id})
	result["availability"] = availabilityToJSON(availResp.GetAvailableSlots())

	c.JSON(http.StatusOK, result)
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
	deptMap := h.fetchDepartmentMap(c)
	c.JSON(http.StatusOK, teacherToJSON(resp.Teacher, deptMap[resp.Teacher.DepartmentId]))
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
	c.JSON(http.StatusOK, gin.H{
		"teacher_id":   resp.TeacherId,
		"availability": availabilityToJSON(resp.AvailableSlots),
	})
}

func (h *HRHandler) UpdateTeacherAvailability(c *gin.Context) {
	var body struct {
		AvailableSlots []struct {
			DayOfWeek int32  `json:"day_of_week"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		} `json:"available_slots"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slots := make([]*hrv1.TimeSlot, 0, len(body.AvailableSlots))
	for _, s := range body.AvailableSlots {
		sp := hrTimeToSlot(s.StartTime)
		if sp == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid start_time: %s", s.StartTime)})
			return
		}
		slots = append(slots, &hrv1.TimeSlot{
			DayOfWeek:   s.DayOfWeek,
			StartPeriod: sp,
			EndPeriod:   sp + 1,
		})
	}

	resp, err := h.teachers.UpdateTeacherAvailability(c.Request.Context(), &hrv1.UpdateTeacherAvailabilityRequest{
		TeacherId:      c.Param("id"),
		AvailableSlots: slots,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"teacher_id":   resp.TeacherId,
		"availability": availabilityToJSON(resp.AvailableSlots),
	})
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

	data := make([]gin.H, len(resp.Departments))
	for i, d := range resp.Departments {
		data[i] = departmentToJSON(d)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
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
	c.JSON(http.StatusCreated, departmentToJSON(resp.Department))
}
