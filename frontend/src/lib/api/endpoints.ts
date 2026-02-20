// Centralized API route constants — keeps all paths in one place
export const ENDPOINTS = {
  auth: {
    login: '/auth/login',
    register: '/auth/register',
    me: '/auth/me',
    logout: '/auth/logout',
  },
  hr: {
    teachers: '/hr/teachers',
    teacher: (id: string) => `/hr/teachers/${id}`,
    departments: '/hr/departments',
    department: (id: string) => `/hr/departments/${id}`,
  },
  subjects: {
    list: '/subjects',
    detail: (id: string) => `/subjects/${id}`,
    prerequisites: (id: string) => `/subjects/${id}/prerequisites`,
  },
  timetable: {
    semesters: '/timetable/semesters',
    semester: (id: string) => `/timetable/semesters/${id}`,
    generate: (semesterId: string) => `/timetable/semesters/${semesterId}/generate`,
    slots: (semesterId: string) => `/timetable/semesters/${semesterId}/slots`,
    schedules: '/timetable/schedules',
    schedule: (id: string) => `/timetable/schedules/${id}`,
  },
  dashboard: {
    stats: '/dashboard/stats',
  },
} as const
