package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// Consumer subscribes to domain event streams and updates dimension/fact tables.
type Consumer struct {
	consumer messaging.Consumer
	repo     *persistence.AnalyticsRepository
	log      *zap.Logger
}

// NewConsumer returns a Consumer ready to start.
func NewConsumer(consumer messaging.Consumer, repo *persistence.AnalyticsRepository, log *zap.Logger) *Consumer {
	return &Consumer{consumer: consumer, repo: repo, log: log}
}

// Start subscribes to all four domain event streams. Non-blocking; each handler
// runs in a background goroutine managed by the messaging consumer.
func (c *Consumer) Start(ctx context.Context) error {
	subs := []struct {
		durable string
		subject string
		handler func(*messaging.Message) error
	}{
		{"analytics-hr", "hr.>", c.handleHRMessage},
		{"analytics-subject", "subject.>", c.handleSubjectMessage},
		{"analytics-timetable", "timetable.>", c.handleTimetableMessage},
		{"analytics-student", "student.>", c.handleStudentMessage},
	}
	for _, s := range subs {
		if err := c.consumer.Subscribe(ctx, s.durable, s.subject, s.handler); err != nil {
			c.log.Warn("subscribe failed", zap.String("subject", s.subject), zap.Error(err))
			// Non-fatal: analytics is best-effort
		}
	}
	return nil
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

func (c *Consumer) handleHRMessage(msg *messaging.Message) error {
	switch msg.Subject {
	case "hr.teacher.created", "hr.teacher.updated":
		var ev hrTeacherEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr teacher event", zap.Error(err))
			return nil
		}
		tid, _ := uuid.Parse(ev.TeacherID)
		did, _ := uuid.Parse(ev.DepartmentID)
		if err := c.repo.UpsertTeacher(context.Background(), entity.DimTeacher{
			TeacherID:       tid,
			FullName:        ev.FullName,
			DepartmentID:    did,
			DepartmentName:  ev.DepartmentName,
			Specializations: ev.Specializations,
			UpdatedAt:       time.Now(),
		}); err != nil {
			c.log.Error("upsert teacher", zap.Error(err))
		}

	case "hr.teacher.deleted":
		var ev struct {
			TeacherID string `json:"teacher_id"`
		}
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr teacher deleted event", zap.Error(err))
			return nil
		}
		tid, _ := uuid.Parse(ev.TeacherID)
		if err := c.repo.DeleteTeacher(context.Background(), tid); err != nil {
			c.log.Error("delete teacher", zap.Error(err))
		}

	case "hr.department.created", "hr.department.updated":
		var ev hrDepartmentEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal hr department event", zap.Error(err))
			return nil
		}
		did, _ := uuid.Parse(ev.DepartmentID)
		if err := c.repo.UpsertDepartment(context.Background(), entity.DimDepartment{
			DepartmentID: did,
			Name:         ev.Name,
			Code:         ev.Code,
			UpdatedAt:    time.Now(),
		}); err != nil {
			c.log.Error("upsert department", zap.Error(err))
		}
	}
	return nil
}

// --- Subject events ---

type subjectEvent struct {
	SubjectID    string `json:"subject_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Credits      int    `json:"credits"`
	DepartmentID string `json:"department_id"`
}

func (c *Consumer) handleSubjectMessage(msg *messaging.Message) error {
	switch msg.Subject {
	case "subject.created", "subject.updated":
		var ev subjectEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal subject event", zap.Error(err))
			return nil
		}
		sid, _ := uuid.Parse(ev.SubjectID)
		did, _ := uuid.Parse(ev.DepartmentID)
		if err := c.repo.UpsertSubject(context.Background(), entity.DimSubject{
			SubjectID:    sid,
			Name:         ev.Name,
			Code:         ev.Code,
			Credits:      ev.Credits,
			DepartmentID: did,
			UpdatedAt:    time.Now(),
		}); err != nil {
			c.log.Error("upsert subject", zap.Error(err))
		}

	case "subject.deleted":
		var ev struct {
			SubjectID string `json:"subject_id"`
		}
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal subject deleted event", zap.Error(err))
			return nil
		}
		sid, _ := uuid.Parse(ev.SubjectID)
		if err := c.repo.DeleteSubject(context.Background(), sid); err != nil {
			c.log.Error("delete subject", zap.Error(err))
		}
	}
	return nil
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

func (c *Consumer) handleTimetableMessage(msg *messaging.Message) error {
	switch msg.Subject {
	case "timetable.semester.created":
		var ev semesterEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal timetable semester event", zap.Error(err))
			return nil
		}
		startDate, _ := time.Parse(time.DateOnly, ev.StartDate)
		endDate, _ := time.Parse(time.DateOnly, ev.EndDate)
		if err := c.repo.UpsertSemester(context.Background(), entity.DimSemester{
			SemesterID: mustParseUUID(ev.SemesterID),
			Name:       ev.Name,
			Year:       ev.Year,
			Term:       ev.Term,
			StartDate:  startDate,
			EndDate:    endDate,
			UpdatedAt:  time.Now(),
		}); err != nil {
			c.log.Error("upsert semester", zap.Error(err))
		}

	case "timetable.schedule.generated":
		var ev scheduleGeneratedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal timetable schedule event", zap.Error(err))
			return nil
		}
		for _, e := range ev.Entries {
			if err := c.repo.UpsertScheduleEntry(context.Background(), entity.FactScheduleEntry{
				ScheduleID: mustParseUUID(e.ScheduleID),
				SemesterID: mustParseUUID(e.SemesterID),
				TeacherID:  mustParseUUID(e.TeacherID),
				SubjectID:  mustParseUUID(e.SubjectID),
				RoomID:     mustParseUUID(e.RoomID),
				DayOfWeek:  e.DayOfWeek,
				Period:     e.Period,
				IsAssigned: e.IsAssigned,
				CreatedAt:  time.Now(),
			}); err != nil {
				c.log.Error("upsert schedule entry", zap.Error(err), zap.String("schedule_id", e.ScheduleID))
			}
		}
	}
	return nil
}

// --- Student events ---

type studentEnrollmentApprovedEvent struct {
	EnrollmentID    string `json:"enrollment_id"`
	StudentID       string `json:"student_id"`
	StudentCode     string `json:"student_code"`
	StudentFullName string `json:"student_full_name"`
	DepartmentID    string `json:"department_id"`
	EnrollmentYear  int    `json:"enrollment_year"`
	SubjectID       string `json:"subject_id"`
	SemesterID      string `json:"semester_id"`
	EnrolledAt      string `json:"enrolled_at"`
}

type studentGradeAssignedEvent struct {
	EnrollmentID string  `json:"enrollment_id"`
	StudentID    string  `json:"student_id"`
	SubjectID    string  `json:"subject_id"`
	SemesterID   string  `json:"semester_id"`
	GradeNumeric float64 `json:"grade_numeric"`
	GradeLetter  string  `json:"grade_letter"`
	GradedAt     string  `json:"graded_at"`
}

func (c *Consumer) handleStudentMessage(msg *messaging.Message) error {
	switch msg.Subject {
	case "student.enrollment_approved":
		var ev studentEnrollmentApprovedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal student enrollment_approved event", zap.Error(err))
			return nil
		}
		did, _ := uuid.Parse(ev.DepartmentID)
		sid, _ := uuid.Parse(ev.StudentID)
		_ = c.repo.UpsertStudent(context.Background(), entity.DimStudent{
			StudentID:      sid,
			StudentCode:    ev.StudentCode,
			FullName:       ev.StudentFullName,
			DepartmentID:   did,
			EnrollmentYear: ev.EnrollmentYear,
			UpdatedAt:      time.Now(),
		})
		enrolledAt, _ := time.Parse(time.RFC3339, ev.EnrolledAt)
		if enrolledAt.IsZero() {
			enrolledAt = time.Now()
		}
		eid, _ := uuid.Parse(ev.EnrollmentID)
		subjectID, _ := uuid.Parse(ev.SubjectID)
		semesterID, _ := uuid.Parse(ev.SemesterID)
		if err := c.repo.UpsertEnrollment(context.Background(), entity.FactEnrollment{
			EnrollmentID: eid,
			StudentID:    sid,
			SubjectID:    subjectID,
			SemesterID:   semesterID,
			Status:       "approved",
			EnrolledAt:   enrolledAt,
		}); err != nil {
			c.log.Error("upsert enrollment", zap.Error(err))
		}

	case "student.grade_assigned":
		var ev studentGradeAssignedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			c.log.Error("unmarshal student grade_assigned event", zap.Error(err))
			return nil
		}
		gradedAt, _ := time.Parse(time.RFC3339, ev.GradedAt)
		if gradedAt.IsZero() {
			gradedAt = time.Now()
		}
		eid, _ := uuid.Parse(ev.EnrollmentID)
		sid, _ := uuid.Parse(ev.StudentID)
		subjectID, _ := uuid.Parse(ev.SubjectID)
		semesterID, _ := uuid.Parse(ev.SemesterID)
		if err := c.repo.UpsertEnrollment(context.Background(), entity.FactEnrollment{
			EnrollmentID: eid,
			StudentID:    sid,
			SubjectID:    subjectID,
			SemesterID:   semesterID,
			Status:       "completed",
			GradeNumeric: &ev.GradeNumeric,
			GradeLetter:  &ev.GradeLetter,
			GradedAt:     &gradedAt,
		}); err != nil {
			c.log.Error("update enrollment grade", zap.Error(err))
		}
	}
	return nil
}

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
