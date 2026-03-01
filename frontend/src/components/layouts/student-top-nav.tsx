import { Link, useRouterState } from '@tanstack/react-router'
import { GraduationCap, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils/cn'
import { useLogout } from '@/lib/hooks/use-auth'
import { authStore } from '@/lib/stores/auth-store'

const NAV_TABS = [
  { label: 'Dashboard', to: '/student/dashboard' },
  { label: 'My Subjects', to: '/student/subjects' },
  { label: 'Transcript', to: '/student/transcript' },
  { label: 'Profile', to: '/student/profile' },
] as const

// Minimal top navigation for the student self-service portal
export function StudentTopNav() {
  const router = useRouterState()
  const pathname = router.location.pathname
  const logout = useLogout()
  const user = authStore.getUser()

  return (
    <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur">
      <div className="container mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
        {/* Logo */}
        <div className="flex items-center gap-2">
          <GraduationCap className="h-5 w-5 text-primary" />
          <span className="text-sm font-bold">Student Portal</span>
        </div>

        {/* Tab navigation */}
        <nav className="flex items-center gap-1">
          {NAV_TABS.map(({ label, to }) => (
            <Link
              key={to}
              to={to}
              className={cn(
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                pathname.startsWith(to)
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {label}
            </Link>
          ))}
        </nav>

        {/* User + logout */}
        <div className="flex items-center gap-2">
          {user && (
            <span className="hidden text-xs text-muted-foreground sm:block">{user.full_name}</span>
          )}
          <Button variant="ghost" size="icon" onClick={logout} title="Sign out">
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
