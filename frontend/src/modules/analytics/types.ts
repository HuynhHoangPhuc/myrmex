// Analytics module types — mirrors backend analytics service DTOs

export interface WorkloadStat {
  teacher_id: string
  teacher_name: string
  department_id: string
  semester_id: string
  subject_id: string
  hours_per_week: number
  total_hours: number
}

export interface UtilizationStat {
  department_id: string
  department_name: string
  semester_id: string
  assigned_slots: number
  total_slots: number
  utilization_pct: number
}

export interface DashboardSummary {
  total_teachers: number
  total_departments: number
  total_subjects: number
  total_semesters: number
}

export interface DepartmentMetric {
  department_id: string
  department_name: string
  teacher_count: number
  subject_count: number
}

export interface ScheduleMetric {
  semester_id: string
  semester_name: string
  assigned_slots: number
  total_slots: number
}

export interface ScheduleHeatmapCell {
  day_of_week: number
  period: number
  entry_count: number
}
