package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
)

// AnalyticsRepository handles all analytics reads/writes against Postgres.
type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepository creates a new AnalyticsRepository.
func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

// UpsertTeacher inserts or updates a teacher dimension record.
func (r *AnalyticsRepository) UpsertTeacher(ctx context.Context, t entity.DimTeacher) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics.dim_teacher
			(teacher_id, full_name, department_id, department_name, specializations, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (teacher_id) DO UPDATE SET
			full_name        = EXCLUDED.full_name,
			department_id    = EXCLUDED.department_id,
			department_name  = EXCLUDED.department_name,
			specializations  = EXCLUDED.specializations,
			updated_at       = EXCLUDED.updated_at`,
		t.TeacherID, t.FullName, t.DepartmentID, t.DepartmentName, t.Specializations, t.UpdatedAt,
	)
	return err
}

// DeleteTeacher removes a teacher dimension record.
func (r *AnalyticsRepository) DeleteTeacher(ctx context.Context, teacherID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM analytics.dim_teacher WHERE teacher_id = $1`, teacherID)
	return err
}

// UpsertDepartment inserts or updates a department dimension record.
func (r *AnalyticsRepository) UpsertDepartment(ctx context.Context, d entity.DimDepartment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics.dim_department (department_id, name, code, updated_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (department_id) DO UPDATE SET
			name       = EXCLUDED.name,
			code       = EXCLUDED.code,
			updated_at = EXCLUDED.updated_at`,
		d.DepartmentID, d.Name, d.Code, d.UpdatedAt,
	)
	return err
}

// UpsertSubject inserts or updates a subject dimension record.
func (r *AnalyticsRepository) UpsertSubject(ctx context.Context, s entity.DimSubject) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics.dim_subject
			(subject_id, name, code, credits, department_id, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (subject_id) DO UPDATE SET
			name          = EXCLUDED.name,
			code          = EXCLUDED.code,
			credits       = EXCLUDED.credits,
			department_id = EXCLUDED.department_id,
			updated_at    = EXCLUDED.updated_at`,
		s.SubjectID, s.Name, s.Code, s.Credits, s.DepartmentID, s.UpdatedAt,
	)
	return err
}

// DeleteSubject removes a subject dimension record.
func (r *AnalyticsRepository) DeleteSubject(ctx context.Context, subjectID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM analytics.dim_subject WHERE subject_id = $1`, subjectID)
	return err
}

// UpsertSemester inserts or updates a semester dimension record.
func (r *AnalyticsRepository) UpsertSemester(ctx context.Context, s entity.DimSemester) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics.dim_semester (semester_id, name, year, term, start_date, end_date, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (semester_id) DO UPDATE SET
			name       = EXCLUDED.name,
			year       = EXCLUDED.year,
			term       = EXCLUDED.term,
			start_date = EXCLUDED.start_date,
			end_date   = EXCLUDED.end_date,
			updated_at = EXCLUDED.updated_at`,
		s.SemesterID, s.Name, s.Year, s.Term, s.StartDate, s.EndDate, s.UpdatedAt,
	)
	return err
}

// UpsertScheduleEntry inserts or updates a schedule entry fact record.
func (r *AnalyticsRepository) UpsertScheduleEntry(ctx context.Context, e entity.FactScheduleEntry) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics.fact_schedule_entry
			(schedule_id, semester_id, teacher_id, subject_id, room_id, day_of_week, period, is_assigned, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (schedule_id, day_of_week, period, room_id) DO UPDATE SET
			semester_id = EXCLUDED.semester_id,
			teacher_id  = EXCLUDED.teacher_id,
			subject_id  = EXCLUDED.subject_id,
			is_assigned = EXCLUDED.is_assigned`,
		e.ScheduleID, e.SemesterID, e.TeacherID, e.SubjectID, e.RoomID,
		e.DayOfWeek, e.Period, e.IsAssigned, e.CreatedAt,
	)
	return err
}

// GetWorkloadStats returns workload stats filtered by semester.
func (r *AnalyticsRepository) GetWorkloadStats(ctx context.Context, semesterID uuid.UUID) ([]entity.WorkloadStat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			w.teacher_id,
			COALESCE(t.full_name, ''),
			COALESCE(t.department_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(t.department_name, ''),
			w.semester_id,
			w.subject_id,
			COALESCE(s.code, ''),
			w.hours_per_week,
			w.total_hours
		FROM analytics.fact_workload w
		LEFT JOIN analytics.dim_teacher t ON t.teacher_id = w.teacher_id
		LEFT JOIN analytics.dim_subject s ON s.subject_id = w.subject_id
		WHERE ($1::uuid IS NULL OR w.semester_id = $1)
		ORDER BY w.total_hours DESC`,
		semesterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query workload stats: %w", err)
	}
	defer rows.Close()

	var stats []entity.WorkloadStat
	for rows.Next() {
		var s entity.WorkloadStat
		if err := rows.Scan(
			&s.TeacherID, &s.TeacherName, &s.DepartmentID, &s.DepartmentName,
			&s.SemesterID, &s.SubjectID, &s.SubjectCode, &s.HoursPerWeek, &s.TotalHours,
		); err != nil {
			return nil, fmt.Errorf("scan workload row: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// GetUtilizationStats returns utilization stats grouped by department.
func (r *AnalyticsRepository) GetUtilizationStats(ctx context.Context, semesterID uuid.UUID) ([]entity.UtilizationStat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(t.department_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(t.department_name, 'Unknown'),
			f.semester_id,
			COUNT(*) FILTER (WHERE f.is_assigned) AS assigned_slots,
			COUNT(*) AS total_slots
		FROM analytics.fact_schedule_entry f
		LEFT JOIN analytics.dim_teacher t ON t.teacher_id = f.teacher_id
		WHERE ($1::uuid IS NULL OR f.semester_id = $1)
		GROUP BY t.department_id, t.department_name, f.semester_id
		ORDER BY t.department_name`,
		semesterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query utilization stats: %w", err)
	}
	defer rows.Close()

	var stats []entity.UtilizationStat
	for rows.Next() {
		var s entity.UtilizationStat
		if err := rows.Scan(
			&s.DepartmentID, &s.DepartmentName, &s.SemesterID,
			&s.AssignedSlots, &s.TotalSlots,
		); err != nil {
			return nil, fmt.Errorf("scan utilization row: %w", err)
		}
		if s.TotalSlots > 0 {
			s.UtilizationPct = float64(s.AssignedSlots) / float64(s.TotalSlots) * 100
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// GetDashboardSummary returns aggregate counts from dimension tables.
// GetSemesterName returns the display name of a semester by ID.
// Returns empty string if not found (caller should fall back to "All Semesters").
func (r *AnalyticsRepository) GetSemesterName(ctx context.Context, semesterID uuid.UUID) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx,
		`SELECT name FROM analytics.dim_semester WHERE semester_id = $1`,
		semesterID,
	).Scan(&name)
	if err != nil {
		return "", nil // not found → caller uses fallback
	}
	return name, nil
}

func (r *AnalyticsRepository) GetDashboardSummary(ctx context.Context) (entity.DashboardSummary, error) {
	var s entity.DashboardSummary
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM analytics.dim_teacher),
			(SELECT COUNT(*) FROM analytics.dim_department),
			(SELECT COUNT(*) FROM analytics.dim_subject),
			(SELECT COUNT(*) FROM analytics.dim_semester)`,
	).Scan(&s.TotalTeachers, &s.TotalDepartments, &s.TotalSubjects, &s.TotalSemesters)
	if err != nil {
		return s, fmt.Errorf("query dashboard summary: %w", err)
	}
	return s, nil
}

// GetDepartmentMetrics returns per-department teacher/subject counts.
func (r *AnalyticsRepository) GetDepartmentMetrics(ctx context.Context) ([]entity.DepartmentMetric, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			d.department_id,
			d.name,
			COUNT(DISTINCT t.teacher_id) AS teacher_count,
			COUNT(DISTINCT s.subject_id) AS subject_count
		FROM analytics.dim_department d
		LEFT JOIN analytics.dim_teacher  t ON t.department_id = d.department_id
		LEFT JOIN analytics.dim_subject  s ON s.department_id = d.department_id
		GROUP BY d.department_id, d.name
		ORDER BY d.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("query department metrics: %w", err)
	}
	defer rows.Close()

	var metrics []entity.DepartmentMetric
	for rows.Next() {
		var m entity.DepartmentMetric
		if err := rows.Scan(&m.DepartmentID, &m.DepartmentName, &m.TeacherCount, &m.SubjectCount); err != nil {
			return nil, fmt.Errorf("scan department metric row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetScheduleMetrics returns per-semester assigned/total slot counts.
func (r *AnalyticsRepository) GetScheduleMetrics(ctx context.Context, semesterID uuid.UUID) ([]entity.ScheduleMetric, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			f.semester_id,
			COALESCE(sm.name, ''),
			COUNT(*) FILTER (WHERE f.is_assigned) AS assigned_slots,
			COUNT(*) AS total_slots
		FROM analytics.fact_schedule_entry f
		LEFT JOIN analytics.dim_semester sm ON sm.semester_id = f.semester_id
		WHERE ($1::uuid IS NULL OR f.semester_id = $1)
		GROUP BY f.semester_id, sm.name
		ORDER BY sm.name`,
		semesterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query schedule metrics: %w", err)
	}
	defer rows.Close()

	var metrics []entity.ScheduleMetric
	for rows.Next() {
		var m entity.ScheduleMetric
		if err := rows.Scan(&m.SemesterID, &m.SemesterName, &m.AssignedSlots, &m.TotalSlots); err != nil {
			return nil, fmt.Errorf("scan schedule metric row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetScheduleHeatmap returns entry counts grouped by day-of-week and period for heatmap display.
func (r *AnalyticsRepository) GetScheduleHeatmap(ctx context.Context, semesterID uuid.UUID) ([]entity.ScheduleHeatmapCell, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT day_of_week, period, COUNT(*) AS entry_count
		FROM analytics.fact_schedule_entry
		WHERE ($1::uuid IS NULL OR semester_id = $1)
		GROUP BY day_of_week, period
		ORDER BY day_of_week, period`,
		semesterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query schedule heatmap: %w", err)
	}
	defer rows.Close()

	var cells []entity.ScheduleHeatmapCell
	for rows.Next() {
		var c entity.ScheduleHeatmapCell
		if err := rows.Scan(&c.DayOfWeek, &c.Period, &c.EntryCount); err != nil {
			return nil, fmt.Errorf("scan schedule heatmap row: %w", err)
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}
