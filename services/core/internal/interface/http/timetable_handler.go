package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// TimetableHandler proxies HTTP requests to the Timetable gRPC microservice.
type TimetableHandler struct {
	timetable timetablev1.TimetableServiceClient
	semesters timetablev1.SemesterServiceClient
	subjects  subjectv1.SubjectServiceClient // nil-safe — used for UUID-to-name resolution
	natsJS    jetstream.JetStream            // nil-safe — SSE endpoint returns 503 if nil
}

func NewTimetableHandler(
	timetable timetablev1.TimetableServiceClient,
	semesters timetablev1.SemesterServiceClient,
	subjects subjectv1.SubjectServiceClient,
	natsJS jetstream.JetStream,
) *TimetableHandler {
	return &TimetableHandler{timetable: timetable, semesters: semesters, subjects: subjects, natsJS: natsJS}
}

// buildSubjectMap fetches all subjects and returns a map of ID → Subject proto (best-effort).
func (h *TimetableHandler) buildSubjectMap(ctx context.Context) map[string]*subjectv1.Subject {
	if h.subjects == nil {
		return nil
	}
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

// semesterToJSON converts a proto Semester to a frontend-compatible JSON map.
// academic_year is derived from the year and term numbers.
// is_active is inferred by checking whether today falls within the semester dates.
func semesterToJSON(s *timetablev1.Semester) gin.H {
	startDate := protoTime(s.StartDate)
	endDate := protoTime(s.EndDate)

	// Derive academic_year string from year and term (e.g. "2026 Term 1")
	academicYear := fmt.Sprintf("%d Term %d", s.Year, s.Term)

	// Compute is_active: semester is active if today is between start and end dates
	isActive := false
	if s.StartDate != nil && s.EndDate != nil {
		now := time.Now()
		isActive = !now.Before(s.StartDate.AsTime()) && !now.After(s.EndDate.AsTime())
	}

	offeredSubjectIDs := s.OfferedSubjectIds
	if offeredSubjectIDs == nil {
		offeredSubjectIDs = []string{}
	}
	return gin.H{
		"id":                  s.Id,
		"name":                s.Name,
		"year":                s.Year,
		"term":                s.Term,
		"academic_year":       academicYear,
		"start_date":          startDate,
		"end_date":            endDate,
		"is_active":           isActive,
		"offered_subject_ids": offeredSubjectIDs,
		"time_slots":          []gin.H{},
		"rooms":               []gin.H{},
		"created_at":          "",
		"updated_at":          "",
	}
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

	data := make([]gin.H, len(resp.Semesters))
	for i, s := range resp.Semesters {
		data[i] = semesterToJSON(s)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     resp.Pagination.GetTotal(),
		"page":      page,
		"page_size": pageSize,
	})
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
	c.JSON(http.StatusCreated, semesterToJSON(resp.Semester))
}

func (h *TimetableHandler) GetSemester(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.semesters.GetSemester(c.Request.Context(), &timetablev1.GetSemesterRequest{Id: id})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Enrich with time slots (best-effort)
	slotsResp, _ := h.semesters.ListTimeSlots(c.Request.Context(), &timetablev1.ListTimeSlotsRequest{SemesterId: id})

	// Rooms: if semester has specific room_ids use only those, otherwise return all rooms
	var enrichedRooms []gin.H
	if len(resp.Semester.GetRoomIds()) > 0 {
		allRoomsResp, _ := h.timetable.ListRooms(c.Request.Context(), &timetablev1.ListRoomsRequest{})
		selectedIDs := make(map[string]bool, len(resp.Semester.RoomIds))
		for _, rid := range resp.Semester.RoomIds {
			selectedIDs[rid] = true
		}
		for _, r := range allRoomsResp.GetRooms() {
			if selectedIDs[r.Id] {
				enrichedRooms = append(enrichedRooms, gin.H{
					"id":        r.Id,
					"name":      r.Name,
					"capacity":  r.Capacity,
					"room_type": r.RoomType,
				})
			}
		}
	} else {
		allRoomsResp, _ := h.timetable.ListRooms(c.Request.Context(), &timetablev1.ListRoomsRequest{})
		enrichedRooms = roomsToJSON(allRoomsResp.GetRooms())
	}
	if enrichedRooms == nil {
		enrichedRooms = []gin.H{}
	}

	result := semesterToJSON(resp.Semester)
	result["time_slots"] = timeSlotsToJSON(slotsResp.GetTimeSlots())
	result["rooms"] = enrichedRooms

	// Enrich offered_subject_ids with subject names for human-readable AI responses (best-effort)
	if len(resp.Semester.GetOfferedSubjectIds()) > 0 {
		subjectMap := h.buildSubjectMap(c.Request.Context())
		offeredSubjects := make([]gin.H, 0, len(resp.Semester.OfferedSubjectIds))
		for _, subID := range resp.Semester.OfferedSubjectIds {
			if s, ok := subjectMap[subID]; ok {
				offeredSubjects = append(offeredSubjects, gin.H{"id": s.Id, "name": s.Name, "code": s.Code})
			} else {
				offeredSubjects = append(offeredSubjects, gin.H{"id": subID})
			}
		}
		result["offered_subjects"] = offeredSubjects
	}

	c.JSON(http.StatusOK, result)
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
	c.JSON(http.StatusOK, semesterToJSON(resp.Semester))
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
