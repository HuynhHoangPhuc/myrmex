package entity

import (
	"time"

	"github.com/google/uuid"
)

// DimTeacher is the teacher dimension table record.
type DimTeacher struct {
	TeacherID       uuid.UUID
	FullName        string
	DepartmentID    uuid.UUID
	DepartmentName  string
	Specializations []string
	UpdatedAt       time.Time
}

// DimDepartment is the department dimension table record.
type DimDepartment struct {
	DepartmentID uuid.UUID
	Name         string
	Code         string
	UpdatedAt    time.Time
}

// DimSubject is the subject dimension table record.
type DimSubject struct {
	SubjectID    uuid.UUID
	Name         string
	Code         string
	Credits      int
	DepartmentID uuid.UUID
	UpdatedAt    time.Time
}

// DimSemester is the semester dimension table record.
type DimSemester struct {
	SemesterID uuid.UUID
	Name       string
	Year       int
	Term       string
	StartDate  time.Time
	EndDate    time.Time
	UpdatedAt  time.Time
}

// FactWorkload is a denormalized workload record per teacher per semester.
type FactWorkload struct {
	ID           uuid.UUID
	TeacherID    uuid.UUID
	SemesterID   uuid.UUID
	SubjectID    uuid.UUID
	HoursPerWeek float64
	TotalHours   float64
	CreatedAt    time.Time
}

// FactScheduleEntry is a denormalized schedule entry record.
type FactScheduleEntry struct {
	ID         uuid.UUID
	ScheduleID uuid.UUID
	SemesterID uuid.UUID
	TeacherID  uuid.UUID
	SubjectID  uuid.UUID
	RoomID     uuid.UUID
	DayOfWeek  int
	Period     int
	IsAssigned bool
	CreatedAt  time.Time
}

// WorkloadStat is the query result for teacher workload.
type WorkloadStat struct {
	TeacherID      uuid.UUID `json:"teacher_id"`
	TeacherName    string    `json:"teacher_name"`
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	SemesterID     uuid.UUID `json:"semester_id"`
	SubjectID      uuid.UUID `json:"subject_id"`
	SubjectCode    string    `json:"subject_code"`
	HoursPerWeek   float64   `json:"hours_per_week"`
	TotalHours     float64   `json:"total_hours"`
}

// UtilizationStat is the query result for room/department utilization.
type UtilizationStat struct {
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	SemesterID     uuid.UUID `json:"semester_id"`
	AssignedSlots  int       `json:"assigned_slots"`
	TotalSlots     int       `json:"total_slots"`
	UtilizationPct float64   `json:"utilization_pct"`
}

// DashboardSummary aggregates counts for the dashboard view.
type DashboardSummary struct {
	TotalTeachers    int `json:"total_teachers"`
	TotalDepartments int `json:"total_departments"`
	TotalSubjects    int `json:"total_subjects"`
	TotalSemesters   int `json:"total_semesters"`
}

// DepartmentMetric aggregates per-department stats.
type DepartmentMetric struct {
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	TeacherCount   int       `json:"teacher_count"`
	SubjectCount   int       `json:"subject_count"`
}

// ScheduleMetric aggregates per-semester schedule stats.
type ScheduleMetric struct {
	SemesterID    uuid.UUID `json:"semester_id"`
	SemesterName  string    `json:"semester_name"`
	AssignedSlots int       `json:"assigned_slots"`
	TotalSlots    int       `json:"total_slots"`
}

// ScheduleHeatmapCell aggregates entry counts by day-of-week and period for heatmap.
type ScheduleHeatmapCell struct {
	DayOfWeek  int `json:"day_of_week"`
	Period     int `json:"period"`
	EntryCount int `json:"entry_count"`
}
