import { useRouterState } from '@tanstack/react-router'
import { LogOut, User, ChevronRight } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { useCurrentUser } from '@/lib/hooks/use-current-user'
import { useLogout } from '@/lib/hooks/use-auth'

// Derive breadcrumbs from current pathname
function useBreadcrumbs() {
  const router = useRouterState()
  const segments = router.location.pathname.split('/').filter(Boolean)
  return segments.map((seg, i) => ({
    label: seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, ' '),
    path: '/' + segments.slice(0, i + 1).join('/'),
    isLast: i === segments.length - 1,
  }))
}

// Top bar: breadcrumbs on left, user dropdown on right
export function TopBar() {
  const breadcrumbs = useBreadcrumbs()
  const { data: user } = useCurrentUser()
  const logout = useLogout()

  return (
    <header className="flex h-14 items-center justify-between border-b bg-background px-6">
      {/* Breadcrumbs */}
      <nav className="flex items-center gap-1 text-sm text-muted-foreground">
        {breadcrumbs.map((crumb) => (
          <span key={crumb.path} className="flex items-center gap-1">
            {!crumb.isLast && <ChevronRight className="h-3 w-3" />}
            <span className={crumb.isLast ? 'font-medium text-foreground' : ''}>
              {crumb.label}
            </span>
          </span>
        ))}
      </nav>

      {/* User dropdown */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
              {user?.full_name?.charAt(0).toUpperCase() ?? '?'}
            </div>
            <span className="hidden sm:inline">{user?.full_name ?? 'User'}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuLabel className="font-normal">
            <div className="flex flex-col">
              <span className="font-medium">{user?.full_name}</span>
              <span className="text-xs text-muted-foreground">{user?.email}</span>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>
            <User className="mr-2 h-4 w-4" />
            Profile
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-destructive focus:text-destructive"
            onClick={logout}
          >
            <LogOut className="mr-2 h-4 w-4" />
            Log out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
