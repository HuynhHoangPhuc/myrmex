// Timetable module domain types

export interface TimeSlot {
  id: string
  semester_id: string
  day_of_week: number // 1=Monday..6=Saturday
  start_time: string  // "08:00"
  end_time: string    // "09:30"
  slot_index: number
}

export interface Room {
  id: string
  name: string
  capacity: number
  room_type: 'lecture' | 'lab' | 'seminar'
}

export interface Semester {
  id: string
  name: string
  year: number
  term: number
  academic_year: string
  start_date: string
  end_date: string
  is_active: boolean
  offered_subject_ids: string[]
  time_slots: TimeSlot[]
  rooms: Room[]
  created_at: string
  updated_at: string
}

export interface CreateSemesterInput {
  name: string
  year: number
  term: number
  start_date: string  // RFC3339 format
  end_date: string    // RFC3339 format
}

export interface CreateTimeSlotInput {
  day_of_week: number  // 0=Mon..5=Sat
  start_period: number // 1-8
  end_period: number   // 1-8, must be > start_period
}

export type TimeSlotPreset = 'standard' | 'mwf' | 'tuth'

export type ScheduleStatus = 'pending' | 'generating' | 'completed' | 'failed'

export interface ScheduleEntry {
  id: string
  schedule_id?: string
  subject_id: string
  subject_code: string
  subject_name: string
  teacher_id: string
  teacher_name: string
  room_id: string
  room_name: string
  day_of_week: number    // 1=Mon..6=Sat
  start_period: number   // lesson period number (1-8)
  end_period: number
  is_manual_override: boolean
  department_id: string
}

export interface Schedule {
  id: string
  semester_id: string
  status: ScheduleStatus
  score: number
  hard_violations: number
  soft_violations: number
  entries: ScheduleEntry[]
  created_at: string
  updated_at: string
}

export interface GenerateScheduleInput {
  semester_id: string
  timeout_seconds?: number
}

export interface TeacherSuggestion {
  teacher_id: string
  teacher_name: string
  score: number
  reasons: string[]
  is_available: boolean
}

export interface AssignTeacherInput {
  entry_id: string
  teacher_id: string
}
