package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// TimetableHandler proxies HTTP requests to the Timetable gRPC microservice.
type TimetableHandler struct {
	timetable timetablev1.TimetableServiceClient
	semesters timetablev1.SemesterServiceClient
	natsJS    jetstream.JetStream // nil-safe — SSE endpoint returns 503 if nil
}

func NewTimetableHandler(
	timetable timetablev1.TimetableServiceClient,
	semesters timetablev1.SemesterServiceClient,
	natsJS jetstream.JetStream,
) *TimetableHandler {
	return &TimetableHandler{timetable: timetable, semesters: semesters, natsJS: natsJS}
}

// --- Semester handlers ---

func (h *TimetableHandler) ListSemesters(c *gin.Context) {
	page, pageSize := parsePage(c)
	resp, err := h.semesters.ListSemesters(c.Request.Context(), &timetablev1.ListSemestersRequest{
		Pagination: &corev1.PaginationRequest{Page: page, PageSize: pageSize},
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"semesters": resp.Semesters, "pagination": resp.Pagination})
}

func (h *TimetableHandler) CreateSemester(c *gin.Context) {
	var body struct {
		Name      string `json:"name" binding:"required"`
		Year      int32  `json:"year" binding:"required"`
		Term      int32  `json:"term" binding:"required"`
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, body.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use RFC3339"})
		return
	}
	endDate, err := time.Parse(time.RFC3339, body.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use RFC3339"})
		return
	}

	resp, err := h.semesters.CreateSemester(c.Request.Context(), &timetablev1.CreateSemesterRequest{
		Name:      body.Name,
		Year:      body.Year,
		Term:      body.Term,
		StartDate: timestamppb.New(startDate),
		EndDate:   timestamppb.New(endDate),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"semester": resp.Semester})
}

func (h *TimetableHandler) GetSemester(c *gin.Context) {
	resp, err := h.semesters.GetSemester(c.Request.Context(), &timetablev1.GetSemesterRequest{
		Id: c.Param("id"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"semester": resp.Semester})
}

func (h *TimetableHandler) AddOfferedSubject(c *gin.Context) {
	var body struct {
		SubjectID string `json:"subject_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.semesters.AddOfferedSubject(c.Request.Context(), &timetablev1.AddOfferedSubjectRequest{
		SemesterId: c.Param("id"),
		SubjectId:  body.SubjectID,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"semester": resp.Semester})
}

func (h *TimetableHandler) RemoveOfferedSubject(c *gin.Context) {
	_, err := h.semesters.RemoveOfferedSubject(c.Request.Context(), &timetablev1.RemoveOfferedSubjectRequest{
		SemesterId: c.Param("id"),
		SubjectId:  c.Param("subjectId"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
