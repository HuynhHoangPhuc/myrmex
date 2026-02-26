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
    scheduleStream: (id: string) => `/timetable/schedules/${id}/stream`,
    suggestTeachers: '/timetable/suggest-teachers',
    manualAssign: (scheduleId: string, entryId: string) =>
      `/timetable/schedules/${scheduleId}/entries/${entryId}`,
    offeredSubjects: (semesterId: string) =>
      `/timetable/semesters/${semesterId}/offered-subjects`,
  },
  dashboard: {
    stats: '/dashboard/stats',
  },
  analytics: {
    workload: '/analytics/workload',
    utilization: '/analytics/utilization',
    dashboard: '/analytics/dashboard',
    departmentMetrics: '/analytics/department-metrics',
    scheduleMetrics: '/analytics/schedule-metrics',
    scheduleHeatmap: '/analytics/schedule-heatmap',
    export: '/analytics/export',
  },
} as const
