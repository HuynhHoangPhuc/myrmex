import { createFileRoute, redirect, Outlet } from '@tanstack/react-router'
import { authStore } from '@/lib/stores/auth-store'

// Admin guard: only admin and super_admin can access /admin/* routes
export const Route = createFileRoute('/_authenticated/admin')({
  beforeLoad: () => {
    const user = authStore.getUser()
    const role = user?.role
    if (role !== 'admin' && role !== 'super_admin') {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: () => <Outlet />,
})
