package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// DashboardHandler aggregates stats from all module gRPC services.
type DashboardHandler struct {
	teachers    hrv1.TeacherServiceClient
	departments hrv1.DepartmentServiceClient
	subjects    subjectv1.SubjectServiceClient
	semesters   timetablev1.SemesterServiceClient
}

func NewDashboardHandler(
	teachers hrv1.TeacherServiceClient,
	departments hrv1.DepartmentServiceClient,
	subjects subjectv1.SubjectServiceClient,
	semesters timetablev1.SemesterServiceClient,
) *DashboardHandler {
	return &DashboardHandler{
		teachers:    teachers,
		departments: departments,
		subjects:    subjects,
		semesters:   semesters,
	}
}

// Stats returns aggregated counts for dashboard display.
func (h *DashboardHandler) Stats(c *gin.Context) {
	ctx := c.Request.Context()
	// Use page_size=1 to get total counts cheaply via pagination metadata
	smallPage := &corev1.PaginationRequest{Page: 1, PageSize: 1}

	var (
		totalTeachers    int32
		totalDepts       int32
		totalSubjects    int32
		activeSemesters  int32
	)

	if h.teachers != nil {
		if resp, err := h.teachers.ListTeachers(ctx, &hrv1.ListTeachersRequest{Pagination: smallPage}); err == nil && resp.Pagination != nil {
			totalTeachers = resp.Pagination.Total
		}
	}

	if h.departments != nil {
		if resp, err := h.departments.ListDepartments(ctx, &hrv1.ListDepartmentsRequest{Pagination: smallPage}); err == nil && resp.Pagination != nil {
			totalDepts = resp.Pagination.Total
		}
	}

	if h.subjects != nil {
		if resp, err := h.subjects.ListSubjects(ctx, &subjectv1.ListSubjectsRequest{Pagination: smallPage}); err == nil && resp.Pagination != nil {
			totalSubjects = resp.Pagination.Total
		}
	}

	if h.semesters != nil {
		if resp, err := h.semesters.ListSemesters(ctx, &timetablev1.ListSemestersRequest{Pagination: smallPage}); err == nil && resp.Pagination != nil {
			activeSemesters = resp.Pagination.Total
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_teachers":    totalTeachers,
		"total_departments": totalDepts,
		"total_subjects":    totalSubjects,
		"active_semesters":  activeSemesters,
	})
}
