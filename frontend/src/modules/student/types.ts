// Student module domain types

export interface Student {
  id: string
  student_code: string
  user_id?: string
  full_name: string
  email: string
  department_id: string
  enrollment_year: number
  status: 'active' | 'graduated' | 'suspended'
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateStudentInput {
  student_code: string
  full_name: string
  email: string
  department_id: string
  enrollment_year: number
}

export interface UpdateStudentInput {
  full_name?: string
  department_id?: string
  status?: Student['status']
}

export interface CreateStudentResponse {
  student: Student
  user_id: string
  temp_password: string
}

export type EnrollmentStatus = 'pending' | 'approved' | 'rejected' | 'completed'

export interface EnrollmentRequest {
  id: string
  student_id: string
  semester_id: string
  offered_subject_id: string
  subject_id: string
  status: EnrollmentStatus
  request_note?: string
  admin_note?: string
  requested_at: string
  reviewed_at?: string
  reviewed_by?: string
}

export interface ReviewEnrollmentInput {
  approve: boolean
  admin_note?: string
}

export interface AssignGradeInput {
  enrollment_id: string
  grade_numeric: number
  graded_by: string
  notes?: string
}

export interface UpdateGradeInput {
  grade_numeric: number
  graded_by: string
  notes?: string
}

export interface Grade {
  id: string
  enrollment_id: string
  grade_numeric: number
  grade_letter: string
  graded_by: string
  graded_at: string
  notes?: string
}

export interface TranscriptEntry {
  enrollment_id: string
  semester_id: string
  subject_id: string
  subject_code: string
  subject_name: string
  credits: number
  status: EnrollmentStatus
  grade_numeric?: number
  grade_letter?: string
  graded_at?: string
}

export interface Transcript {
  student: Student
  entries: TranscriptEntry[]
  gpa: number
  total_credits: number
  passed_credits: number
}

export interface MissingPrerequisite {
  subject_id: string
  subject_code: string
  subject_name: string
  type: 'strict' | 'recommended'
}

export interface PrerequisiteCheckResult {
  can_enroll: boolean
  missing: MissingPrerequisite[]
}

export interface RequestEnrollmentInput {
  student_id: string
  semester_id: string
  offered_subject_id: string
  subject_id: string
  request_note?: string
}
