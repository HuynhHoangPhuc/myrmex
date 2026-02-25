import { createFileRoute, redirect } from '@tanstack/react-router'

// /timetable → redirect to /timetable/semesters
export const Route = createFileRoute('/_authenticated/timetable/')({
  beforeLoad: () => {
    throw redirect({ to: '/timetable/semesters', search: { page: 1, pageSize: 25 } })
  },
  component: () => null,
})
