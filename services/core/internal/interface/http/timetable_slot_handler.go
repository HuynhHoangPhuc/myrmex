package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// periodStartTime maps period numbers to their standard university start times.
// Matches the frontend's period-to-time.ts mapping.
var periodStartTimes = map[int]string{
	1: "08:00", 2: "09:45", 3: "11:30", 4: "13:15",
	5: "15:00", 6: "16:45", 7: "18:30", 8: "20:15",
}

// periodEndTimes maps period numbers to their standard university end times.
var periodEndTimes = map[int]string{
	1: "09:30", 2: "11:15", 3: "13:00", 4: "14:45",
	5: "16:30", 6: "18:15", 7: "20:00", 8: "21:45",
}

func periodToStartTime(period int) string {
	if t, ok := periodStartTimes[period]; ok {
		return t
	}
	return fmt.Sprintf("P%d", period)
}

func periodToEndTime(period int) string {
	if t, ok := periodEndTimes[period]; ok {
		return t
	}
	return fmt.Sprintf("P%d", period)
}

// timeSlotsToJSON converts a slice of TimeSlot protos to frontend-compatible JSON maps.
func timeSlotsToJSON(slots []*timetablev1.TimeSlot) []gin.H {
	result := make([]gin.H, len(slots))
	for i, sl := range slots {
		result[i] = gin.H{
			"id":           sl.Id,
			"semester_id":  sl.SemesterId,
			"day_of_week":  sl.DayOfWeek,
			"start_period": sl.StartPeriod,
			"end_period":   sl.EndPeriod,
			"start_time":   periodToStartTime(int(sl.StartPeriod)),
			"end_time":     periodToEndTime(int(sl.EndPeriod)),
			"slot_index":   i,
		}
	}
	return result
}

func (h *TimetableHandler) CreateTimeSlot(c *gin.Context) {
	var body struct {
		DayOfWeek   int32 `json:"day_of_week" binding:"min=0,max=5"`
		StartPeriod int32 `json:"start_period" binding:"required,min=1,max=8"`
		EndPeriod   int32 `json:"end_period" binding:"required,min=1,max=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.EndPeriod <= body.StartPeriod {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_period must be greater than start_period"})
		return
	}
	resp, err := h.semesters.CreateTimeSlot(c.Request.Context(), &timetablev1.CreateTimeSlotRequest{
		SemesterId:  c.Param("id"),
		DayOfWeek:   body.DayOfWeek,
		StartPeriod: body.StartPeriod,
		EndPeriod:   body.EndPeriod,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	sl := resp.TimeSlot
	slotsResp, _ := h.semesters.ListTimeSlots(c.Request.Context(), &timetablev1.ListTimeSlotsRequest{SemesterId: c.Param("id")})
	slotIndex := len(slotsResp.GetTimeSlots()) - 1
	c.JSON(http.StatusCreated, gin.H{
		"id":           sl.Id,
		"semester_id":  sl.SemesterId,
		"day_of_week":  sl.DayOfWeek,
		"start_period": sl.StartPeriod,
		"end_period":   sl.EndPeriod,
		"start_time":   periodToStartTime(int(sl.StartPeriod)),
		"end_time":     periodToEndTime(int(sl.EndPeriod)),
		"slot_index":   slotIndex,
	})
}

func (h *TimetableHandler) DeleteTimeSlot(c *gin.Context) {
	_, err := h.semesters.DeleteTimeSlot(c.Request.Context(), &timetablev1.DeleteTimeSlotRequest{
		Id: c.Param("slotId"),
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TimetableHandler) ApplyTimeSlotPreset(c *gin.Context) {
	var body struct {
		Preset string `json:"preset" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.semesters.ApplyTimeSlotPreset(c.Request.Context(), &timetablev1.ApplyTimeSlotPresetRequest{
		SemesterId: c.Param("id"),
		Preset:     body.Preset,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"time_slots": timeSlotsToJSON(resp.TimeSlots)})
}
