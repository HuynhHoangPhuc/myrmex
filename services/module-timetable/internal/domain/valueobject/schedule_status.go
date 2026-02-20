package valueobject

// ScheduleStatus represents the lifecycle state of a schedule.
type ScheduleStatus string

const (
	ScheduleStatusDraft     ScheduleStatus = "draft"
	ScheduleStatusPublished ScheduleStatus = "published"
	ScheduleStatusArchived  ScheduleStatus = "archived"
)

func (s ScheduleStatus) IsValid() bool {
	switch s {
	case ScheduleStatusDraft, ScheduleStatusPublished, ScheduleStatusArchived:
		return true
	}
	return false
}

func (s ScheduleStatus) String() string { return string(s) }
