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

export type UserRole = 'admin' | 'manager' | 'viewer'

export interface User {
  id: string
  email: string
  full_name: string
  role: UserRole
  created_at: string
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
