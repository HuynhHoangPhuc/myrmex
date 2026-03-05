package messaging

import (
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/application/command"
	notif_email "github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/email"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// subscribedSubjects lists all domain event subjects the notification module processes.
var subscribedSubjects = []string{
	"student.enrollment_approved",
	"student.enrollment_rejected",
	"student.enrollment_requested",
	"student.grade_assigned",
	"hr.teacher.created",
	"timetable.semester.created",
	"timetable.entry.assigned",
	"timetable.schedule.generated",
	"core.user.role_updated",
	"notification.system.announcement",
}

// EventConsumer subscribes to domain events via the messaging backend and dispatches notifications.
type EventConsumer struct {
	consumer   messaging.Consumer
	dispatch   *command.DispatchNotificationCommand
	renderer   *notif_email.TemplateRenderer
	emailQueue *persistence.EmailQueueRepository
	resolver   *RecipientResolver
	log        *zap.Logger
}

func NewEventConsumer(
	consumer messaging.Consumer,
	dispatch *command.DispatchNotificationCommand,
	renderer *notif_email.TemplateRenderer,
	emailQueue *persistence.EmailQueueRepository,
	resolver *RecipientResolver,
	log *zap.Logger,
) *EventConsumer {
	return &EventConsumer{
		consumer:   consumer,
		dispatch:   dispatch,
		renderer:   renderer,
		emailQueue: emailQueue,
		resolver:   resolver,
		log:        log,
	}
}

// Start subscribes to each domain event subject independently.
// Each subscription creates a durable consumer on the source stream.
func (c *EventConsumer) Start(ctx context.Context) error {
	for _, subject := range subscribedSubjects {
		// Durable name: "notif-dispatcher-" + sanitized subject
		durable := "notif-dispatcher-" + strings.ReplaceAll(subject, ".", "-")
		sub := subject // capture for closure
		if err := c.consumer.Subscribe(ctx, durable, sub, func(msg *messaging.Message) error {
			return c.handleMessage(ctx, msg)
		}); err != nil {
			return err
		}
	}
	c.log.Info("notification event consumer started", zap.Int("subscriptions", len(subscribedSubjects)))
	return nil
}

func (c *EventConsumer) handleMessage(ctx context.Context, msg *messaging.Message) error {
	subject := msg.Subject
	spec, ok := GetEventSpec(subject)
	if !ok {
		return nil // unknown subject — ack and skip
	}

	var payload map[string]any
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		c.log.Warn("notification event: malformed payload", zap.String("subject", subject), zap.Error(err))
		return nil // bad payload is non-retryable — ack and discard
	}

	title := spec.BuildTitle(payload)
	body := spec.BuildBody(payload)

	recipients, err := c.resolveRecipients(ctx, subject, payload)
	if err != nil {
		c.log.Warn("notification event: recipient resolution failed",
			zap.String("subject", subject), zap.Error(err))
		return err // transient failure — nack to retry
	}
	if len(recipients) == 0 {
		return nil
	}

	for _, r := range recipients {
		c.dispatchToRecipient(ctx, spec, r, title, body)
	}
	return nil
}

func (c *EventConsumer) dispatchToRecipient(
	ctx context.Context,
	spec EventSpec,
	recipient RecipientInfo,
	title, body string,
) {
	for _, ch := range spec.Channels {
		notifID, err := c.dispatch.Execute(ctx, command.DispatchInput{
			UserID:  recipient.UserID,
			Type:    spec.NotifType,
			Channel: ch,
			Title:   title,
			Body:    body,
		})
		if err != nil {
			c.log.Warn("dispatch failed",
				zap.String("user_id", recipient.UserID),
				zap.String("channel", ch),
				zap.Error(err),
			)
			continue
		}

		if ch == "email" && notifID != "" && recipient.Email != "" {
			emailSubject, html, renderErr := c.renderer.Render(spec.NotifType, title, body, "")
			if renderErr != nil {
				c.log.Warn("email render failed", zap.String("notif_type", spec.NotifType), zap.Error(renderErr))
				continue
			}
			if enqErr := c.emailQueue.Enqueue(ctx, notifID, recipient.Email, emailSubject, html); enqErr != nil {
				c.log.Warn("email enqueue failed", zap.String("notif_id", notifID), zap.Error(enqErr))
			}
		}
	}
}

// resolveRecipients returns the list of recipients for a given subject + payload.
func (c *EventConsumer) resolveRecipients(ctx context.Context, subject string, p map[string]any) ([]RecipientInfo, error) {
	switch subject {
	case "student.enrollment_approved", "student.enrollment_rejected", "student.enrollment_requested":
		studentID := strFromPayload(p, "StudentID", "student_id")
		if studentID == "" {
			return nil, nil
		}
		info, err := c.resolver.ByStudentID(ctx, studentID)
		if err != nil {
			return nil, err
		}
		return []RecipientInfo{info}, nil

	case "student.grade_assigned":
		enrollmentID := strFromPayload(p, "EnrollmentID", "enrollment_id")
		if enrollmentID == "" {
			return nil, nil
		}
		info, err := c.resolver.ByEnrollmentID(ctx, enrollmentID)
		if err != nil {
			return nil, err
		}
		return []RecipientInfo{info}, nil

	case "hr.teacher.created":
		deptID := strFromPayload(p, "DepartmentID", "department_id")
		if deptID == "" {
			return nil, nil
		}
		info, err := c.resolver.ByDeptHead(ctx, deptID)
		if err != nil {
			return nil, err
		}
		return []RecipientInfo{info}, nil

	case "timetable.semester.created":
		return c.resolver.AllTeachers(ctx)

	case "timetable.entry.assigned":
		teacherID := strFromPayload(p, "teacher_id", "TeacherID")
		if teacherID == "" {
			return nil, nil
		}
		info, err := c.resolver.ByHRTeacherID(ctx, teacherID)
		if err != nil {
			return nil, err
		}
		return []RecipientInfo{info}, nil

	case "timetable.schedule.generated":
		scheduleID := strFromPayload(p, "schedule_id", "ScheduleID")
		if scheduleID == "" {
			return nil, nil
		}
		return c.resolver.BySchedule(ctx, scheduleID)

	case "core.user.role_updated":
		userID := strFromPayload(p, "user_id", "UserID")
		if userID == "" {
			return nil, nil
		}
		info, err := c.resolver.ByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		return []RecipientInfo{info}, nil

	case "notification.system.announcement":
		return c.resolver.AllUsers(ctx)

	default:
		return nil, nil
	}
}
