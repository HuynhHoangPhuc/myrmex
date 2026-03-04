// Shared API response types used across all modules

export interface ListResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, string[]>
}

export type UserRole =
  | 'super_admin'
  | 'admin'
  | 'dean'
  | 'dept_head'
  | 'manager'
  | 'viewer'
  | 'student'
  | 'teacher'

export interface User {
  id: string
  email: string
  full_name: string
  role: UserRole
  department_id?: string
  is_active?: boolean
  created_at: string
  updated_at?: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface DashboardStats {
  total_teachers: number
  total_departments: number
  total_subjects: number
  active_semesters: number
}
