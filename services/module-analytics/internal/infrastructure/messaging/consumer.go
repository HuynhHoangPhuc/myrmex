package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// Consumer subscribes to NATS JetStream streams and updates dimension/fact tables.
type Consumer struct {
	js   nats.JetStreamContext
	repo *persistence.AnalyticsRepository
	log  *zap.Logger
}

// NewConsumer connects to NATS and returns a Consumer ready to start.
func NewConsumer(url string, repo *persistence.AnalyticsRepository, log *zap.Logger) (*Consumer, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return &Consumer{js: js, repo: repo, log: log}, nil
}

// Start subscribes to all three streams. It is non-blocking; each handler runs in its own goroutine.
func (c *Consumer) Start(ctx context.Context) {
	c.subscribeHR(ctx)
	c.subscribeSubject(ctx)
	c.subscribeTimetable(ctx)
}

// --- HR events ---

type hrTeacherEvent struct {
	TeacherID       string   `json:"teacher_id"`
	FullName        string   `json:"full_name"`
	DepartmentID    string   `json:"department_id"`
	DepartmentName  string   `json:"department_name"`
	Specializations []string `json:"specializations"`
}

type hrDepartmentEvent struct {
	DepartmentID string `json:"department_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
}

func (c *Consumer) subscribeHR(ctx context.Context) {
	sub, err := c.js.PullSubscribe("hr.>", "analytics-hr",
		nats.BindStream("HR_EVENTS"),
		nats.MaxDeliver(5),
	)
	if err != nil {
		c.log.Warn("subscribe HR_EVENTS failed", zap.Error(err))
		return
	}
	go c.pullLoop(ctx, sub, c.handleHRMessage)
}

func (c *Consumer) handleHRMessage(ctx context.Context, msg *nats.Msg) {
	switch msg.Subject {
	case "hr.teacher.created", "hr.teacher.updated":
		var ev hrTeacherEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr teacher event", zap.Error(err))
			return
		}
		tid, _ := uuid.Parse(ev.TeacherID)
		did, _ := uuid.Parse(ev.DepartmentID)
		t := entity.DimTeacher{
			TeacherID:       tid,
			FullName:        ev.FullName,
			DepartmentID:    did,
			DepartmentName:  ev.DepartmentName,
			Specializations: ev.Specializations,
			UpdatedAt:       time.Now(),
		}
		if err := c.repo.UpsertTeacher(ctx, t); err != nil {
			c.log.Error("upsert teacher", zap.Error(err))
		}

	case "hr.teacher.deleted":
		var ev struct {
			TeacherID string `json:"teacher_id"`
		}
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr teacher deleted event", zap.Error(err))
			return
		}
		tid, _ := uuid.Parse(ev.TeacherID)
		if err := c.repo.DeleteTeacher(ctx, tid); err != nil {
			c.log.Error("delete teacher", zap.Error(err))
		}

	case "hr.department.created", "hr.department.updated":
		var ev hrDepartmentEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr department event", zap.Error(err))
			return
		}
		did, _ := uuid.Parse(ev.DepartmentID)
		d := entity.DimDepartment{
			DepartmentID: did,
			Name:         ev.Name,
			Code:         ev.Code,
			UpdatedAt:    time.Now(),
		}
		if err := c.repo.UpsertDepartment(ctx, d); err != nil {
			c.log.Error("upsert department", zap.Error(err))
		}
	}
}

// --- Subject events ---

type subjectEvent struct {
	SubjectID    string `json:"subject_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Credits      int    `json:"credits"`
	DepartmentID string `json:"department_id"`
}

func (c *Consumer) subscribeSubject(ctx context.Context) {
	sub, err := c.js.PullSubscribe("subject.>", "analytics-subject",
		nats.BindStream("SUBJECT_EVENTS"),
		nats.MaxDeliver(5),
	)
	if err != nil {
		c.log.Warn("subscribe SUBJECT_EVENTS failed", zap.Error(err))
		return
	}
	go c.pullLoop(ctx, sub, c.handleSubjectMessage)
}

func (c *Consumer) handleSubjectMessage(ctx context.Context, msg *nats.Msg) {
	switch msg.Subject {
	case "subject.created", "subject.updated":
		var ev subjectEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal subject event", zap.Error(err))
			return
		}
		sid, _ := uuid.Parse(ev.SubjectID)
		did, _ := uuid.Parse(ev.DepartmentID)
		s := entity.DimSubject{
			SubjectID:    sid,
			Name:         ev.Name,
			Code:         ev.Code,
			Credits:      ev.Credits,
			DepartmentID: did,
			UpdatedAt:    time.Now(),
		}
		if err := c.repo.UpsertSubject(ctx, s); err != nil {
			c.log.Error("upsert subject", zap.Error(err))
		}

	case "subject.deleted":
		var ev struct {
			SubjectID string `json:"subject_id"`
		}
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal subject deleted event", zap.Error(err))
			return
		}
		sid, _ := uuid.Parse(ev.SubjectID)
		if err := c.repo.DeleteSubject(ctx, sid); err != nil {
			c.log.Error("delete subject", zap.Error(err))
		}
	}
}

// --- Timetable events ---

type semesterEvent struct {
	SemesterID string `json:"semester_id"`
	Name       string `json:"name"`
	Year       int    `json:"year"`
	Term       string `json:"term"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type scheduleEntryEvent struct {
	ScheduleID string `json:"schedule_id"`
	SemesterID string `json:"semester_id"`
	TeacherID  string `json:"teacher_id"`
	SubjectID  string `json:"subject_id"`
	RoomID     string `json:"room_id"`
	DayOfWeek  int    `json:"day_of_week"`
	Period     int    `json:"period"`
	IsAssigned bool   `json:"is_assigned"`
}

type scheduleGeneratedEvent struct {
	Entries []scheduleEntryEvent `json:"entries"`
}

func (c *Consumer) subscribeTimetable(ctx context.Context) {
	sub, err := c.js.PullSubscribe("timetable.>", "analytics-timetable",
		nats.BindStream("TIMETABLE"),
		nats.MaxDeliver(5),
	)
	if err != nil {
		c.log.Warn("subscribe TIMETABLE failed", zap.Error(err))
		return
	}
	go c.pullLoop(ctx, sub, c.handleTimetableMessage)
}

func (c *Consumer) handleTimetableMessage(ctx context.Context, msg *nats.Msg) {
	switch msg.Subject {
	case "timetable.semester.created":
		c.handleSemesterCreated(ctx, msg)
		return
	case "timetable.schedule.generated":
		// handled below
	default:
		return
	}
	var ev scheduleGeneratedEvent
	if err := json.Unmarshal(msg.Data, &ev); err != nil {
		c.log.Error("unmarshal timetable schedule event", zap.Error(err))
		return
	}
	for _, e := range ev.Entries {
		entry := entity.FactScheduleEntry{
			ScheduleID: mustParseUUID(e.ScheduleID),
			SemesterID: mustParseUUID(e.SemesterID),
			TeacherID:  mustParseUUID(e.TeacherID),
			SubjectID:  mustParseUUID(e.SubjectID),
			RoomID:     mustParseUUID(e.RoomID),
			DayOfWeek:  e.DayOfWeek,
			Period:     e.Period,
			IsAssigned: e.IsAssigned,
			CreatedAt:  time.Now(),
		}
		if err := c.repo.UpsertScheduleEntry(ctx, entry); err != nil {
			c.log.Error("upsert schedule entry", zap.Error(err), zap.String("schedule_id", e.ScheduleID))
		}
	}
}

func (c *Consumer) handleSemesterCreated(ctx context.Context, msg *nats.Msg) {
	var ev semesterEvent
	if err := json.Unmarshal(msg.Data, &ev); err != nil {
		c.log.Error("unmarshal timetable semester event", zap.Error(err))
		return
	}
	startDate, _ := time.Parse(time.DateOnly, ev.StartDate)
	endDate, _ := time.Parse(time.DateOnly, ev.EndDate)
	s := entity.DimSemester{
		SemesterID: mustParseUUID(ev.SemesterID),
		Name:       ev.Name,
		Year:       ev.Year,
		Term:       ev.Term,
		StartDate:  startDate,
		EndDate:    endDate,
		UpdatedAt:  time.Now(),
	}
	if err := c.repo.UpsertSemester(ctx, s); err != nil {
		c.log.Error("upsert semester", zap.Error(err))
	}
}

// pullLoop fetches messages from a pull subscription in a tight loop until ctx is cancelled.
func (c *Consumer) pullLoop(ctx context.Context, sub *nats.Subscription, handler func(context.Context, *nats.Msg)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil {
				// Timeout is normal; any other error gets logged at debug level.
				continue
			}
			for _, msg := range msgs {
				handler(ctx, msg)
				_ = msg.Ack()
			}
		}
	}
}

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
