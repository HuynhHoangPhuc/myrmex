package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
)

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

// GetTeacherAvailability GET /teachers/:id/availability
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

// UpdateTeacherAvailability PUT /teachers/:id/availability
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
