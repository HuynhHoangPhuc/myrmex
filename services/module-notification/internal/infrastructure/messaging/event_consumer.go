package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

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

// EventConsumer subscribes to domain events via JetStream and dispatches notifications.
type EventConsumer struct {
	js         jetstream.JetStream
	dispatch   *command.DispatchNotificationCommand
	renderer   *notif_email.TemplateRenderer
	emailQueue *persistence.EmailQueueRepository
	resolver   *RecipientResolver
	log        *zap.Logger
}

func NewEventConsumer(
	js jetstream.JetStream,
	dispatch *command.DispatchNotificationCommand,
	renderer *notif_email.TemplateRenderer,
	emailQueue *persistence.EmailQueueRepository,
	resolver *RecipientResolver,
	log *zap.Logger,
) *EventConsumer {
	return &EventConsumer{
		js:         js,
		dispatch:   dispatch,
		renderer:   renderer,
		emailQueue: emailQueue,
		resolver:   resolver,
		log:        log,
	}
}

// Start creates the NOTIFICATION_EVENTS stream and begins consuming.
func (c *EventConsumer) Start(ctx context.Context) error {
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "NOTIFICATION_EVENTS",
		Subjects: subscribedSubjects,
		MaxAge:   7 * 24 * time.Hour, // 7-day retention
	})
	if err != nil {
		return err
	}

	cons, err := c.js.CreateOrUpdateConsumer(ctx, "NOTIFICATION_EVENTS", jetstream.ConsumerConfig{
		Durable:   "notification-dispatcher",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	go func() {
		iter, err := cons.Messages()
		if err != nil {
			c.log.Error("event consumer subscribe failed", zap.Error(err))
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
					c.log.Warn("event consumer next error", zap.Error(err))
					continue
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	c.log.Info("notification event consumer started")
	return nil
}

func (c *EventConsumer) handleMessage(ctx context.Context, msg jetstream.Msg) {
	subject := msg.Subject()
	spec, ok := GetEventSpec(subject)
	if !ok {
		_ = msg.Ack() // unknown subject — skip silently
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		c.log.Warn("notification event: malformed payload", zap.String("subject", subject), zap.Error(err))
		_ = msg.Ack() // bad payload is non-retryable
		return
	}

	title := spec.BuildTitle(payload)
	body := spec.BuildBody(payload)

	recipients, err := c.resolveRecipients(ctx, subject, payload)
	if err != nil {
		c.log.Warn("notification event: recipient resolution failed",
			zap.String("subject", subject), zap.Error(err))
		_ = msg.Nak() // transient failure — retry later
		return
	}
	if len(recipients) == 0 {
		_ = msg.Ack()
		return
	}

	for _, r := range recipients {
		c.dispatchToRecipient(ctx, spec, r, title, body)
	}
	_ = msg.Ack()
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

		// For email channel: render template and enqueue for SMTP delivery
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
