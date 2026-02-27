import { useRouterState, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
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

// Parse the first UUID segment from URL and its preceding context segment
function parseDynamicSegment(segments: string[]): { id: string; type: string } | null {
  for (let i = 0; i < segments.length; i++) {
    if (isUUID(segments[i])) {
      return { id: segments[i], type: segments[i - 1] ?? '' }
    }
  }
  return null
}

// Resolve entity name from React Query cache (reactive via useQuery)
// Always calls all 3 hooks with enabled flags — safe with Rules of Hooks
function useDynamicEntityName(segments: string[]) {
  const dynamic = parseDynamicSegment(segments)
  const id = dynamic?.id ?? ''
  const type = dynamic?.type ?? ''

  const { data: subject } = useQuery({
    queryKey: ['subjects', id],
    queryFn: () => apiClient.get<Subject>(ENDPOINTS.subjects.detail(id)).then((r) => r.data),
    enabled: type === 'subjects' && Boolean(id),
    staleTime: Infinity,
    retry: false,
  })

  const { data: teacher } = useQuery({
    queryKey: ['teachers', id],
    queryFn: () => apiClient.get<Teacher>(ENDPOINTS.hr.teacher(id)).then((r) => r.data),
    enabled: type === 'teachers' && Boolean(id),
    staleTime: Infinity,
    retry: false,
  })

  const { data: semester } = useQuery({
    queryKey: ['semesters', id],
    queryFn: () => apiClient.get<Semester>(ENDPOINTS.timetable.semester(id)).then((r) => r.data),
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

  return segments.map((seg, i) => {
    let label: string
    if (isUUID(seg)) {
      // Show entity name once resolved, ellipsis while loading
      label = entityName ?? '…'
    } else {
      label = SEGMENT_LABELS[seg] ?? (seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, ' '))
    }
    return {
      label,
      path: '/' + segments.slice(0, i + 1).join('/'),
      isLast: i === segments.length - 1,
    }
  })
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
        {breadcrumbs.map((crumb, idx) => (
          <span key={crumb.path} className="flex items-center gap-1">
            {idx > 0 && <ChevronRight className="h-3 w-3 shrink-0" />}
            {crumb.isLast ? (
              <span className="font-medium text-foreground">{crumb.label}</span>
            ) : (
              <Link
                to={crumb.path}
                className="hover:text-foreground transition-colors duration-150"
              >
                {crumb.label}
              </Link>
            )}
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
