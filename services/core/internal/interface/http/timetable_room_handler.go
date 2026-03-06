package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
)

// roomsToJSON converts a slice of Room protos to frontend-compatible JSON maps.
func roomsToJSON(rooms []*timetablev1.Room) []gin.H {
	result := make([]gin.H, len(rooms))
	for i, r := range rooms {
		result[i] = gin.H{
			"id":        r.Id,
			"name":      r.Name,
			"capacity":  r.Capacity,
			"room_type": r.RoomType,
		}
	}
	return result
}

// ListRooms returns all active rooms for use in the room selector.
func (h *TimetableHandler) ListRooms(c *gin.Context) {
	resp, err := h.timetable.ListRooms(c.Request.Context(), &timetablev1.ListRoomsRequest{})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roomsToJSON(resp.GetRooms())})
}

// SetSemesterRooms replaces the room selection for a semester.
func (h *TimetableHandler) SetSemesterRooms(c *gin.Context) {
	var body struct {
		RoomIDs []string `json:"room_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.semesters.SetSemesterRooms(c.Request.Context(), &timetablev1.SetSemesterRoomsRequest{
		SemesterId: c.Param("id"),
		RoomIds:    body.RoomIDs,
	})
	if err != nil {
		c.JSON(grpcToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room_ids": resp.RoomIds})
}
