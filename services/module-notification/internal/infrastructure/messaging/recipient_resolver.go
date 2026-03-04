package messaging

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RecipientInfo holds the resolved recipient identity.
type RecipientInfo struct {
	UserID string
	Email  string
}

// RecipientResolver resolves notification recipients via cross-schema Postgres queries.
// All services share the same Postgres instance, so schema-qualified queries work.
type RecipientResolver struct {
	pool *pgxpool.Pool
}

func NewRecipientResolver(pool *pgxpool.Pool) *RecipientResolver {
	return &RecipientResolver{pool: pool}
}

// ByUserID returns identity for a known core user_id.
func (r *RecipientResolver) ByUserID(ctx context.Context, userID string) (RecipientInfo, error) {
	var info RecipientInfo
	info.UserID = userID
	err := r.pool.QueryRow(ctx,
		`SELECT email FROM core.users WHERE id = $1::uuid AND is_active = true`,
		userID,
	).Scan(&info.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("user %s not found", userID)
	}
	return info, err
}

// ByStudentID resolves user_id and email from the student entity ID.
func (r *RecipientResolver) ByStudentID(ctx context.Context, studentID string) (RecipientInfo, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`SELECT user_id::text FROM student.students WHERE id = $1::uuid AND is_active = true`,
		studentID,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("student %s not found or no linked user", studentID)
	}
	if err != nil {
		return RecipientInfo{}, err
	}
	return r.ByUserID(ctx, userID)
}

// ByEnrollmentID resolves the student's user_id from an enrollment_request ID.
func (r *RecipientResolver) ByEnrollmentID(ctx context.Context, enrollmentID string) (RecipientInfo, error) {
	var studentID string
	err := r.pool.QueryRow(ctx,
		`SELECT student_id::text FROM student.enrollment_requests WHERE id = $1::uuid`,
		enrollmentID,
	).Scan(&studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("enrollment %s not found", enrollmentID)
	}
	if err != nil {
		return RecipientInfo{}, err
	}
	return r.ByStudentID(ctx, studentID)
}

// ByDeptHead returns the dept_head user for the given department.
func (r *RecipientResolver) ByDeptHead(ctx context.Context, deptID string) (RecipientInfo, error) {
	var info RecipientInfo
	err := r.pool.QueryRow(ctx,
		`SELECT id::text, email FROM core.users
		 WHERE department_id = $1::uuid AND role = 'dept_head' AND is_active = true
		 LIMIT 1`,
		deptID,
	).Scan(&info.UserID, &info.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("no dept_head for department %s", deptID)
	}
	return info, err
}

// ByHRTeacherID resolves the core user_id for an HR teacher entity (via shared email).
func (r *RecipientResolver) ByHRTeacherID(ctx context.Context, teacherID string) (RecipientInfo, error) {
	var email string
	err := r.pool.QueryRow(ctx,
		`SELECT email FROM hr.teachers WHERE id = $1::uuid AND is_active = true`,
		teacherID,
	).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("hr teacher %s not found", teacherID)
	}
	if err != nil {
		return RecipientInfo{}, err
	}
	// Resolve core user via matching email
	var info RecipientInfo
	info.Email = email
	err = r.pool.QueryRow(ctx,
		`SELECT id::text FROM core.users WHERE email = $1 AND is_active = true`,
		email,
	).Scan(&info.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecipientInfo{}, fmt.Errorf("no core user with email %s for teacher %s", email, teacherID)
	}
	return info, err
}

// BySchedule returns all teachers assigned in a schedule.
func (r *RecipientResolver) BySchedule(ctx context.Context, scheduleID string) ([]RecipientInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT t.email FROM timetable.schedule_entries se
		 JOIN hr.teachers t ON t.id = se.teacher_id
		 WHERE se.schedule_id = $1::uuid AND t.is_active = true`,
		scheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecipientInfo
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			continue
		}
		var userID string
		if err := r.pool.QueryRow(ctx,
			`SELECT id::text FROM core.users WHERE email = $1 AND is_active = true`, email,
		).Scan(&userID); err == nil {
			results = append(results, RecipientInfo{UserID: userID, Email: email})
		}
	}
	return results, rows.Err()
}

// AllTeachers returns all active teacher-role core users.
func (r *RecipientResolver) AllTeachers(ctx context.Context) ([]RecipientInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, email FROM core.users WHERE role = 'teacher' AND is_active = true`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecipients(rows)
}

// AllUsers returns all active core users (for system announcements).
func (r *RecipientResolver) AllUsers(ctx context.Context) ([]RecipientInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, email FROM core.users WHERE is_active = true`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecipients(rows)
}

func scanRecipients(rows pgx.Rows) ([]RecipientInfo, error) {
	var results []RecipientInfo
	for rows.Next() {
		var info RecipientInfo
		if err := rows.Scan(&info.UserID, &info.Email); err != nil {
			continue
		}
		results = append(results, info)
	}
	return results, rows.Err()
}
