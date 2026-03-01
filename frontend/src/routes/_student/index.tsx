import { createFileRoute, redirect } from '@tanstack/react-router'

// Redirect /student → /student/dashboard
export const Route = createFileRoute('/_student/')({
  beforeLoad: () => {
    throw redirect({ to: '/student/dashboard' })
  },
})
