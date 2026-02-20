import { createFileRoute, redirect } from '@tanstack/react-router'
import { authStore } from '@/lib/stores/auth-store'

// Root index: redirect authenticated users to dashboard, others to login
export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (authStore.isAuthenticated()) {
      throw redirect({ to: '/dashboard' })
    }
    throw redirect({ to: '/login' })
  },
  component: () => null,
})
