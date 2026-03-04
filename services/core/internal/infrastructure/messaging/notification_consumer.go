package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	notifinfra "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/notification"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
)

// notifEvent is the common NATS event payload shape for all 5 notification types.
// Publishers include the fields relevant to their event type.
type notifEvent struct {
	// Recipients
	RecipientUserIDs []string `json:"recipient_user_ids,omitempty"` // schedule_published
	StudentUserID    string   `json:"student_user_id,omitempty"`    // enrollment/grade events
	DepartmentID     string   `json:"department_id,omitempty"`      // enrollment_requested

	// Content context
	SemesterName string `json:"semester_name,omitempty"`
	SubjectName  string `json:"subject_name,omitempty"`
	GradeValue   string `json:"grade_value,omitempty"`

	// Navigation
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	Link         string `json:"link,omitempty"`
}

// notifSpec defines how each NATS subject maps to notification content.
type notifSpec struct {
	subject   string
	notifType string
	title     func(e notifEvent) string
	body      func(e notifEvent) string
	sendEmail bool // "enrollment_requested" is in-app only
}

var notifSpecs = []notifSpec{
	{
		subject:   "timetable.schedule.generation_completed",
		notifType: "schedule_published",
		title:     func(e notifEvent) string { return "Schedule published: " + e.SemesterName },
		body:      func(e notifEvent) string { return "Your teaching schedule for " + e.SemesterName + " is ready." },
		sendEmail: true,
	},
	{
		subject:   "student.enrollment_approved",
		notifType: "enrollment_approved",
		title:     func(e notifEvent) string { return "Enrollment approved: " + e.SubjectName },
		body:      func(e notifEvent) string { return "Your enrollment in " + e.SubjectName + " was approved." },
		sendEmail: true,
	},
	{
		subject:   "student.enrollment_rejected",
		notifType: "enrollment_rejected",
		title:     func(e notifEvent) string { return "Enrollment rejected: " + e.SubjectName },
		body:      func(e notifEvent) string { return "Your enrollment in " + e.SubjectName + " was rejected." },
		sendEmail: true,
	},
	{
		subject:   "student.grade_assigned",
		notifType: "grade_assigned",
		title:     func(e notifEvent) string { return "Grade posted: " + e.SubjectName },
		body:      func(e notifEvent) string { return "You received " + e.GradeValue + " in " + e.SubjectName + "." },
		sendEmail: true,
	},
	{
		subject:   "student.enrollment_requested",
		notifType: "enrollment_requested",
		title:     func(e notifEvent) string { return "New enrollment request: " + e.SubjectName },
		body:      func(e notifEvent) string { return "A student requested enrollment in " + e.SubjectName + "." },
		sendEmail: false, // in-app only per plan
	},
}

// NotificationConsumer subscribes to domain events and dispatches notifications.
type NotificationConsumer struct {
	js       jetstream.JetStream
	repo     *persistence.NotificationRepository
	broker   *notifinfra.WSBroker
	emailSvc *notifinfra.EmailService // nil when Resend not configured
	log      *zap.Logger
	cancel   context.CancelFunc
}

func NewNotificationConsumer(
	js jetstream.JetStream,
	repo *persistence.NotificationRepository,
	broker *notifinfra.WSBroker,
	emailSvc *notifinfra.EmailService,
	log *zap.Logger,
) *NotificationConsumer {
	return &NotificationConsumer{js: js, repo: repo, broker: broker, emailSvc: emailSvc, log: log}
}

// Start creates the NOTIFICATIONS stream and begins consuming all 5 event subjects.
func (c *NotificationConsumer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	subjects := make([]string, 0, len(notifSpecs))
	for _, s := range notifSpecs {
		subjects = append(subjects, s.subject)
	}

	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "NOTIFICATIONS",
		Subjects: subjects,
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil {
		return err
	}

	cons, err := c.js.CreateOrUpdateConsumer(ctx, "NOTIFICATIONS", jetstream.ConsumerConfig{
		Durable:   "notification-dispatcher",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	go c.loop(ctx, cons)
	c.log.Info("notification consumer started")
	return nil
}

func (c *NotificationConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *NotificationConsumer) loop(ctx context.Context, cons jetstream.Consumer) {
	iter, err := cons.Messages()
	if err != nil {
		c.log.Error("notification consumer subscribe failed", zap.Error(err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			iter.Stop()
			return
		default:
			msg, err := iter.Next()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.log.Warn("notification consumer next error", zap.Error(err))
				continue
			}
			c.handle(ctx, msg)
		}
	}
}

func (c *NotificationConsumer) handle(ctx context.Context, msg jetstream.Msg) {
	subject := msg.Subject()
	var spec *notifSpec
	for i := range notifSpecs {
		if notifSpecs[i].subject == subject {
			spec = &notifSpecs[i]
			break
		}
	}
	if spec == nil {
		_ = msg.Ack()
		return
	}

	var event notifEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		c.log.Warn("notification: malformed event", zap.String("subject", subject), zap.Error(err))
		_ = msg.Ack()
		return
	}

	recipients := c.resolveRecipients(ctx, spec.notifType, event)
	for _, userID := range recipients {
		c.dispatch(ctx, spec, event, userID)
	}
	_ = msg.Ack()
}

// resolveRecipients returns the user IDs who should receive this notification.
func (c *NotificationConsumer) resolveRecipients(ctx context.Context, notifType string, e notifEvent) []string {
	switch notifType {
	case "schedule_published":
		return e.RecipientUserIDs
	case "enrollment_approved", "enrollment_rejected", "grade_assigned":
		if e.StudentUserID != "" {
			return []string{e.StudentUserID}
		}
	case "enrollment_requested":
		if e.DepartmentID == "" {
			return nil
		}
		userID, err := c.repo.FindDeptHeadUserID(ctx, e.DepartmentID)
		if err != nil {
			c.log.Warn("notification: dept_head lookup failed", zap.String("dept_id", e.DepartmentID), zap.Error(err))
			return nil
		}
		return []string{userID}
	}
	return nil
}

// dispatch stores the notification, pushes via WS, and optionally sends email.
func (c *NotificationConsumer) dispatch(ctx context.Context, spec *notifSpec, e notifEvent, userID string) {
	prefs, _ := c.repo.GetPreferences(ctx, userID)
	if !prefs.InAppEnabled {
		return
	}
	for _, t := range prefs.DisabledTypes {
		if t == spec.notifType {
			return
		}
	}

	notifData := map[string]string{
		"resource_type": e.ResourceType,
		"resource_id":   e.ResourceID,
		"link":          e.Link,
	}
	id, err := c.repo.Insert(ctx, userID, spec.notifType, spec.title(e), spec.body(e), notifData)
	if err != nil {
		c.log.Error("notification: DB insert failed", zap.String("user_id", userID), zap.Error(err))
		return
	}

	row := persistence.NotificationRow{ID: id, UserID: userID, Type: spec.notifType,
		Title: spec.title(e), Body: spec.body(e), CreatedAt: time.Now()}
	unread, _ := c.repo.CountUnread(ctx, userID)
	c.broker.Push(userID, row, unread)

	if spec.sendEmail && c.emailSvc != nil && prefs.EmailEnabled {
		go c.sendEmail(spec.notifType, userID, e)
	}
}

func (c *NotificationConsumer) sendEmail(notifType, userID string, e notifEvent) {
	email, err := notifinfra.GetUserEmail(userID)
	if err != nil || email == "" {
		return
	}
	var data any
	switch notifType {
	case "schedule_published":
		data = notifinfra.SchedulePublishedData{SemesterName: e.SemesterName, Link: e.Link}
	case "enrollment_approved", "enrollment_rejected":
		data = notifinfra.EnrollmentData{SubjectName: e.SubjectName, Link: e.Link}
	case "grade_assigned":
		data = notifinfra.GradeAssignedData{SubjectName: e.SubjectName, GradeValue: e.GradeValue, Link: e.Link}
	}
	c.emailSvc.Send(email, notifType, data)
}
