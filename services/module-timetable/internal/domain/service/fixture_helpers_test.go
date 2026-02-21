package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
)

func mustUUID(n int) uuid.UUID {
	b := [16]byte{}
	b[15] = byte(n)
	return uuid.UUID(b)
}

func makeSlot(id, day, start, end int) *entity.TimeSlot {
	return &entity.TimeSlot{
		ID:          mustUUID(id),
		DayOfWeek:   day,
		StartPeriod: start,
		EndPeriod:   end,
	}
}

func makeTimeSlot(id, day, start, end int) *entity.TimeSlot {
	return makeSlot(id, day, start, end)
}

func slotsMap(slots ...*entity.TimeSlot) map[uuid.UUID]*entity.TimeSlot {
	m := make(map[uuid.UUID]*entity.TimeSlot, len(slots))
	for _, slot := range slots {
		m[slot.ID] = slot
	}
	return m
}

func makeVar(id int, specs ...string) ScheduleVariable {
	return ScheduleVariable{
		SubjectID:               mustUUID(id),
		SubjectCode:             fmt.Sprintf("S%d", id),
		RequiredSpecializations: specs,
	}
}

func makeAssign(teacher, room, slot int) Assignment {
	return Assignment{
		TeacherID: mustUUID(teacher),
		RoomID:    mustUUID(room),
		SlotID:    mustUUID(slot),
	}
}

func openChecker() *ConstraintChecker {
	return NewConstraintChecker(nil, nil, nil, nil)
}
