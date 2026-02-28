import { useRouterState, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { LogOut, User, ChevronRight, Bot, Menu } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/shared/theme-toggle'
import { useCurrentUser } from '@/lib/hooks/use-current-user'
import { useLogout } from '@/lib/hooks/use-auth'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Subject } from '@/modules/subject/types'
import type { Teacher } from '@/modules/hr/types'
import type { Semester } from '@/modules/timetable/types'

// Human-readable labels for known static path segments
const SEGMENT_LABELS: Record<string, string> = {
  dashboard: 'Dashboard',
  subjects: 'Subjects',
  new: 'New',
  edit: 'Edit',
  prerequisites: 'Prerequisites',
  offerings: 'Offerings',
  hr: 'HR',
  teachers: 'Teachers',
  departments: 'Departments',
  timetable: 'Timetable',
  semesters: 'Semesters',
  schedules: 'Schedules',
  generate: 'Generate Schedule',
  assign: 'Assign Teachers',
  analytics: 'Analytics',
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
const isUUID = (s: string) => UUID_RE.test(s)

function parseDynamicSegment(segments: string[]): { id: string; type: string } | null {
  for (let i = 0; i < segments.length; i++) {
    if (isUUID(segments[i])) {
      return { id: segments[i], type: segments[i - 1] ?? '' }
    }
  }

  return null
}

function useDynamicEntityName(segments: string[]) {
  const dynamic = parseDynamicSegment(segments)
  const id = dynamic?.id ?? ''
  const type = dynamic?.type ?? ''

  const { data: subject } = useQuery({
    queryKey: ['subjects', id],
    queryFn: () => apiClient.get<Subject>(ENDPOINTS.subjects.detail(id)).then((response) => response.data),
    enabled: type === 'subjects' && Boolean(id),
    staleTime: Infinity,
    retry: false,
  })

  const { data: teacher } = useQuery({
    queryKey: ['teachers', id],
    queryFn: () => apiClient.get<Teacher>(ENDPOINTS.hr.teacher(id)).then((response) => response.data),
    enabled: type === 'teachers' && Boolean(id),
    staleTime: Infinity,
    retry: false,
  })

  const { data: semester } = useQuery({
    queryKey: ['semesters', id],
    queryFn: () => apiClient.get<Semester>(ENDPOINTS.timetable.semester(id)).then((response) => response.data),
    enabled: type === 'semesters' && Boolean(id),
    staleTime: Infinity,
    retry: false,
  })

  return subject?.name ?? teacher?.full_name ?? semester?.name ?? null
}

interface Crumb {
  label: string
  path: string
  isLast: boolean
}

function useBreadcrumbs(): Crumb[] {
  const router = useRouterState()
  const segments = router.location.pathname.split('/').filter(Boolean)
  const entityName = useDynamicEntityName(segments)

  return segments.map((segment, index) => {
    const label = isUUID(segment)
      ? entityName ?? '…'
      : SEGMENT_LABELS[segment] ?? (segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' '))

    return {
      label,
      path: '/' + segments.slice(0, index + 1).join('/'),
      isLast: index === segments.length - 1,
    }
  })
}

interface TopBarProps {
  chatOpen?: boolean
  onToggleChat?: () => void
  onOpenMobileNav?: () => void
}

function BreadcrumbLink({ crumb }: { crumb: Crumb }) {
  if (crumb.isLast) {
    return <span className="font-medium text-foreground">{crumb.label}</span>
  }

  return (
    <Link to={crumb.path} className="transition-colors duration-150 hover:text-foreground">
      {crumb.label}
    </Link>
  )
}

// Top bar: breadcrumbs on left, theme + AI toggle + user dropdown on right
export function TopBar({ chatOpen, onToggleChat, onOpenMobileNav }: TopBarProps) {
  const breadcrumbs = useBreadcrumbs()
  const { data: user } = useCurrentUser()
  const logout = useLogout()

  const firstCrumb = breadcrumbs[0]
  const lastCrumb = breadcrumbs[breadcrumbs.length - 1]
  const hasCollapsedMiddleCrumbs = breadcrumbs.length > 2

  return (
    <header className="flex h-14 items-center justify-between gap-3 border-b bg-background px-4 md:px-6">
      <div className="flex min-w-0 flex-1 items-center gap-2">
        {onOpenMobileNav && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-10 w-10 md:hidden"
            onClick={onOpenMobileNav}
            aria-label="Open navigation"
          >
            <Menu className="h-5 w-5" />
          </Button>
        )}

        <div className="min-w-0 flex-1">
          <nav className="hidden items-center gap-1 overflow-hidden text-sm text-muted-foreground sm:flex">
            {breadcrumbs.map((crumb, index) => (
              <span key={crumb.path} className="flex min-w-0 items-center gap-1">
                {index > 0 && <ChevronRight className="h-3 w-3 shrink-0" />}
                <span className="truncate">
                  <BreadcrumbLink crumb={crumb} />
                </span>
              </span>
            ))}
          </nav>

          <nav className="flex items-center gap-1 overflow-hidden text-sm text-muted-foreground sm:hidden">
            {firstCrumb && (
              <span className="min-w-0 truncate">
                <BreadcrumbLink crumb={{ ...firstCrumb, isLast: false }} />
              </span>
            )}
            {hasCollapsedMiddleCrumbs && (
              <>
                <ChevronRight className="h-3 w-3 shrink-0" />
                <span className="shrink-0">…</span>
              </>
            )}
            {lastCrumb && lastCrumb.path !== firstCrumb?.path && (
              <>
                <ChevronRight className="h-3 w-3 shrink-0" />
                <span className="min-w-0 truncate font-medium text-foreground">{lastCrumb.label}</span>
              </>
            )}
          </nav>
        </div>
      </div>

      <div className="flex items-center gap-1 sm:gap-2">
        <ThemeToggle />

        {onToggleChat && (
          <Button
            variant={chatOpen ? 'secondary' : 'ghost'}
            size="sm"
            onClick={onToggleChat}
            className="h-9 gap-1.5 px-3"
            title={chatOpen ? 'Close AI assistant' : 'Open AI assistant'}
          >
            <Bot className="h-4 w-4" />
            <span className="hidden text-xs sm:inline">AI</span>
          </Button>
        )}

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-9 gap-2 px-2 sm:px-3">
              <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                {user?.full_name?.charAt(0).toUpperCase() ?? '?'}
              </div>
              <span className="hidden max-w-32 truncate sm:inline">{user?.full_name ?? 'User'}</span>
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
      </div>
    </header>
  )
}
