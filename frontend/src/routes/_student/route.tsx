import { createFileRoute, redirect, Outlet } from '@tanstack/react-router'
import { StudentTopNav } from '@/components/layouts/student-top-nav'
import { authStore } from '@/lib/stores/auth-store'

// Student portal guard: requires authenticated user with role=student
// Admin/manager users are redirected to the main dashboard
export const Route = createFileRoute('/_student')({
  beforeLoad: ({ location }) => {
    if (!authStore.isAuthenticated()) {
      throw redirect({ to: '/login', search: { redirect: location.href } })
    }
    const user = authStore.getUser()
    if (user?.role !== 'student') {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: StudentPortalLayout,
})

function StudentPortalLayout() {
  return (
    <div className="min-h-screen bg-background">
      <StudentTopNav />
      <main className="container mx-auto max-w-5xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
