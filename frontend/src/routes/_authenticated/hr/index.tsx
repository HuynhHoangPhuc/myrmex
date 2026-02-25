import { createFileRoute, redirect } from '@tanstack/react-router'

// /hr → redirect to /hr/teachers
export const Route = createFileRoute('/_authenticated/hr/')({
  beforeLoad: () => {
    throw redirect({ to: '/hr/teachers', search: { page: 1, pageSize: 25 } })
  },
  component: () => null,
})
