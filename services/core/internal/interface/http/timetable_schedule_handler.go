package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"

	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// --- Schedule handlers (methods on TimetableHandler) ---

// GenerateSchedule triggers async CSP schedule generation; returns 202 Accepted.
func (h *TimetableHandler) GenerateSchedule(c *gin.Context) {
	var body struct {
		TimeoutSeconds int32 `json:"timeout_seconds"`
	}
	// body is optional — ignore bind error
	_ = c.ShouldBindJSON(&body)
	if body.TimeoutSeconds == 0 {
		body.TimeoutSeconds = 30
	}

	resp, err := h.timetable.GenerateSchedule(c.Request.Context(), &timetablev1.GenerateScheduleRequest{
		SemesterId:     c.Param("id"),
		TimeoutSeconds: body.TimeoutSeconds,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"schedule_id":      resp.Schedule.GetId(),
		"status":           resp.Schedule.GetStatus(),
		"is_partial":       resp.IsPartial,
		"unassigned_count": resp.UnassignedCount,
		"message":          "Schedule generation completed",
	})
}

func (h *TimetableHandler) GetSchedule(c *gin.Context) {
	resp, err := h.timetable.GetSchedule(c.Request.Context(), &timetablev1.GetScheduleRequest{
		Id: c.Param("id"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedule": resp.Schedule})
}

// ListSchedules is a stub — proto TimetableService has no ListSchedules RPC.
// TODO: add ListSchedules RPC to timetable proto and implement here.
func (h *TimetableHandler) ListSchedules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"schedules":  []any{},
		"pagination": nil,
	})
}

// ManualAssign assigns a teacher to a schedule entry via PUT /schedules/:id/entries/:entryId.
func (h *TimetableHandler) ManualAssign(c *gin.Context) {
	var body struct {
		TeacherID string `json:"teacher_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.timetable.ManualAssign(c.Request.Context(), &timetablev1.ManualAssignRequest{
		ScheduleId: c.Param("id"),
		EntryId:    c.Param("entryId"),
		TeacherId:  body.TeacherID,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": resp.Entry})
}

// SuggestTeachers returns teacher suggestions for a given slot.
// GET /suggest-teachers?subject_id=&day_of_week=&start_period=&end_period=
func (h *TimetableHandler) SuggestTeachers(c *gin.Context) {
	subjectID := c.Query("subject_id")
	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject_id is required"})
		return
	}

	dayOfWeek, _ := strconv.Atoi(c.Query("day_of_week"))
	startPeriod, _ := strconv.Atoi(c.Query("start_period"))
	endPeriod, _ := strconv.Atoi(c.Query("end_period"))

	resp, err := h.timetable.SuggestTeachers(c.Request.Context(), &timetablev1.SuggestTeachersRequest{
		SubjectId:   subjectID,
		DayOfWeek:   int32(dayOfWeek),
		StartPeriod: int32(startPeriod),
		EndPeriod:   int32(endPeriod),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestions": resp.Suggestions})
}

// StreamScheduleStatus streams schedule generation events via SSE.
// GET /timetable/schedules/:id/stream
func (h *TimetableHandler) StreamScheduleStatus(c *gin.Context) {
	if h.natsJS == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "streaming unavailable"})
		return
	}

	scheduleID := c.Param("id")
	sse := NewSSEWriter(c)
	subject := fmt.Sprintf("timetable.schedule.%s.>", scheduleID)

	cons, err := h.natsJS.OrderedConsumer(c.Request.Context(), "TIMETABLE", jetstream.OrderedConsumerConfig{
		FilterSubjects: []string{subject},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscribe failed"})
		return
	}

	msgCh := make(chan jetstream.Msg, 10)
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		msgCh <- msg
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "consume failed"})
		return
	}
	defer cc.Stop()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			sse.SendHeartbeat()
		case msg := <-msgCh:
			parts := strings.Split(msg.Subject(), ".")
			eventType := parts[len(parts)-1]
			sse.SendEvent(eventType, msg.Data())
			if eventType == "completed" || eventType == "failed" {
				return
			}
		}
	}
}
