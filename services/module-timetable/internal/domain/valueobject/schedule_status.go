package valueobject

// ScheduleStatus represents the lifecycle state of a schedule.
type ScheduleStatus string

const (
	// Generation workflow states (used by CSP solver and frontend)
	ScheduleStatusGenerating ScheduleStatus = "generating"
	ScheduleStatusCompleted  ScheduleStatus = "completed"
	ScheduleStatusFailed     ScheduleStatus = "failed"

	// Manual publish workflow states (used by publish/archive operations)
	ScheduleStatusDraft     ScheduleStatus = "draft"
	ScheduleStatusPublished ScheduleStatus = "published"
	ScheduleStatusArchived  ScheduleStatus = "archived"
)

func (s ScheduleStatus) IsValid() bool {
	switch s {
	case ScheduleStatusGenerating, ScheduleStatusCompleted, ScheduleStatusFailed,
		ScheduleStatusDraft, ScheduleStatusPublished, ScheduleStatusArchived:
		return true
	}
	return false
}

func (s ScheduleStatus) String() string { return string(s) }
