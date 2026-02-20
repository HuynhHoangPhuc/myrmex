// HR module domain types

export interface Department {
  id: string
  name: string
  code: string
  description?: string
  created_at: string
  updated_at: string
}

export interface TeacherAvailability {
  day_of_week: number // 1=Monday..6=Saturday
  start_time: string  // "08:00"
  end_time: string    // "17:00"
}

export interface Teacher {
  id: string
  employee_code: string
  full_name: string
  email: string
  phone?: string
  department_id: string
  department?: Department
  max_hours_per_week: number
  specializations: string[]
  availability: TeacherAvailability[]
  created_at: string
  updated_at: string
}

export interface CreateTeacherInput {
  employee_code: string
  full_name: string
  email: string
  phone?: string
  department_id: string
  max_hours_per_week: number
  specializations: string[]
}

export interface UpdateTeacherInput extends Partial<CreateTeacherInput> {}

export interface CreateDepartmentInput {
  name: string
  code: string
  description?: string
}

export interface UpdateDepartmentInput extends Partial<CreateDepartmentInput> {}
