import { createFileRoute, redirect, Outlet } from '@tanstack/react-router'
import { AppLayout } from '@/components/layouts/app-layout'
import { authStore } from '@/lib/stores/auth-store'

// Auth guard: all routes under _authenticated/ require a valid JWT
// On failure, redirect to /login preserving intended destination
export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ location }) => {
    if (!authStore.isAuthenticated()) {
      throw redirect({
        to: '/login',
        search: { redirect: location.href },
      })
    }
    const user = authStore.getUser()
    if (user?.role === 'student') {
      throw redirect({ to: '/student/dashboard' })
    }
  },
  component: AuthenticatedLayout,
})

function AuthenticatedLayout() {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}
