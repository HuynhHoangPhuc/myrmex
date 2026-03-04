package entity

// Channel constants for notification delivery.
const (
	ChannelInApp = "in_app"
	ChannelEmail = "email"
)

// Event type constants — match notification.type column values.
const (
	EventSchedulePublished   = "schedule.published"
	EventEnrollmentRequested = "enrollment.requested"
	EventEnrollmentApproved  = "enrollment.approved"
	EventEnrollmentRejected  = "enrollment.rejected"
	EventGradePosted         = "grade.posted"
	EventGradeUpdated        = "grade.updated"
	EventAssignmentPosted    = "assignment.posted"
	EventAssignmentDue       = "assignment.due"
	EventAnnouncementPosted  = "announcement.posted"
	EventPaymentDue          = "payment.due"
	EventPaymentReceived     = "payment.received"
	EventAccountUpdated      = "account.updated"
)

// AllEventTypes is the ordered list of all supported event types (12 total).
var AllEventTypes = []string{
	EventSchedulePublished,
	EventEnrollmentRequested,
	EventEnrollmentApproved,
	EventEnrollmentRejected,
	EventGradePosted,
	EventGradeUpdated,
	EventAssignmentPosted,
	EventAssignmentDue,
	EventAnnouncementPosted,
	EventPaymentDue,
	EventPaymentReceived,
	EventAccountUpdated,
}

// AllChannels is the list of supported delivery channels.
var AllChannels = []string{ChannelInApp, ChannelEmail}

// Preference represents one cell in the user's event × channel preference matrix.
type Preference struct {
	EventType string
	Channel   string
	Enabled   bool
}

// DefaultPreferences returns the full 24-item matrix with all enabled.
// Used when a user has no stored preferences yet.
func DefaultPreferences() []Preference {
	prefs := make([]Preference, 0, len(AllEventTypes)*len(AllChannels))
	for _, et := range AllEventTypes {
		for _, ch := range AllChannels {
			prefs = append(prefs, Preference{EventType: et, Channel: ch, Enabled: true})
		}
	}
	return prefs
}
