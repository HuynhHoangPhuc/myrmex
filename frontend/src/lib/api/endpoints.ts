// Centralized API route constants — keeps all paths in one place
export const ENDPOINTS = {
  auth: {
    login: '/auth/login',
    register: '/auth/register',
    registerStudent: '/auth/register-student',
    refresh: '/auth/refresh',
    me: '/auth/me',
    logout: '/auth/logout',
    oauthGoogleLogin: '/auth/oauth/google/login',
    oauthMicrosoftLogin: '/auth/oauth/microsoft/login',
    oauthExchange: '/auth/oauth/exchange',
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
    dag: {
      full: '/subjects/dag/full',
      checkConflicts: '/subjects/dag/check-conflicts',
    },
  },
  timetable: {
    rooms: '/timetable/rooms',
    semesterRooms: (semesterId: string) => `/timetable/semesters/${semesterId}/rooms`,
    semesters: '/timetable/semesters',
    semester: (id: string) => `/timetable/semesters/${id}`,
    generate: (semesterId: string) => `/timetable/semesters/${semesterId}/generate`,
    slots: (semesterId: string) => `/timetable/semesters/${semesterId}/slots`,
    slot: (semesterId: string, slotId: string) => `/timetable/semesters/${semesterId}/slots/${slotId}`,
    slotsPreset: (semesterId: string) => `/timetable/semesters/${semesterId}/slots/preset`,
    schedules: '/timetable/schedules',
    schedule: (id: string) => `/timetable/schedules/${id}`,
    scheduleStream: (id: string) => `/timetable/schedules/${id}/stream`,
    suggestTeachers: '/timetable/suggest-teachers',
    manualAssign: (scheduleId: string, entryId: string) =>
      `/timetable/schedules/${scheduleId}/entries/${entryId}`,
    offeredSubjects: (semesterId: string) =>
      `/timetable/semesters/${semesterId}/offered-subjects`,
  },
  users: {
    list: '/users',
    detail: (id: string) => `/users/${id}`,
    updateRole: (id: string) => `/users/${id}/role`,
  },
  students: {
    list: '/students',
    detail: (id: string) => `/students/${id}`,
    inviteCode: (id: string) => `/students/${id}/invite-code`,
    transcript: (id: string) => `/students/${id}/transcript`,
  },
  enrollments: {
    list: '/enrollments',
    review: (id: string) => `/enrollments/${id}/review`,
  },
  grades: {
    assign: '/grades',
    update: (id: string) => `/grades/${id}`,
  },
  studentPortal: {
    me: '/student/me',
    myEnrollments: '/student/enrollments',
    requestEnrollment: '/student/enrollments',
    checkPrerequisites: '/student/enrollments/check-prerequisites',
    myTranscript: '/student/transcript',
    exportTranscript: '/student/transcript/export',
  },
  auditLogs: {
    list: '/audit-logs',
  },
  notifications: {
    list: '/notifications',
    unreadCount: '/notifications/unread-count',
    markRead: (id: string) => `/notifications/${id}/read`,
    markAllRead: '/notifications/mark-all-read',
    preferences: '/notifications/preferences',
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
