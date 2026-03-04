import { Link, useRouterState } from '@tanstack/react-router'
import {
  LayoutDashboard,
  Users,
  BookOpen,
  Calendar,
  Building2,
  BarChart3,
  GraduationCap,
  ShieldCheck,
  Bell,
} from 'lucide-react'
import { cn } from '@/lib/utils/cn'
import { usePermissions } from '@/lib/hooks/use-permissions'

export interface NavItem {
  label: string
  to: string
  icon: React.ElementType
  children?: { label: string; to: string }[]
}

interface SidebarNavProps {
  onNavigate?: () => void
}

// Returns true if `pathname` matches or is nested under `path`
function matchesPath(pathname: string, path: string): boolean {
  return pathname === path || pathname === path + '/' || pathname.startsWith(path + '/')
}

// Sidebar navigation with role-based visibility and active-state highlighting
export function SidebarNav({ onNavigate }: SidebarNavProps) {
  const router = useRouterState()
  const pathname = router.location.pathname
  const { isAdmin, isSuperAdmin, isDeptHead, canGrade } = usePermissions()

  // Determine which nav sections this role can see
  const canAccessHR = isAdmin || isSuperAdmin || isDeptHead
  const canAccessAdmin = isAdmin || isSuperAdmin

  const navItems: NavItem[] = [
    { label: 'Dashboard', to: '/dashboard', icon: LayoutDashboard },
    ...(canAccessHR
      ? [
          {
            label: 'HR',
            to: '/hr',
            icon: Users,
            children: [
              { label: 'Teachers', to: '/hr/teachers' },
              { label: 'Departments', to: '/hr/departments' },
            ],
          },
        ]
      : []),
    {
      label: 'Subjects',
      to: '/subjects',
      icon: BookOpen,
      children: [
        { label: 'All Subjects', to: '/subjects' },
        { label: 'Prerequisites', to: '/subjects/prerequisites' },
        { label: 'Offerings', to: '/subjects/offerings' },
      ],
    },
    {
      label: 'Timetable',
      to: '/timetable',
      icon: Calendar,
      children: [
        { label: 'Semesters', to: '/timetable/semesters' },
        { label: 'Schedules', to: '/timetable/schedules' },
        ...(isAdmin || isSuperAdmin
          ? [
              { label: 'Generate', to: '/timetable/generate' },
              { label: 'Assign Teachers', to: '/timetable/assign' },
            ]
          : []),
      ],
    },
    ...(isAdmin || isSuperAdmin || canGrade
      ? [
          {
            label: 'Students',
            to: '/students',
            icon: GraduationCap,
            children: [
              ...(isAdmin || isSuperAdmin
                ? [
                    { label: 'Students', to: '/students' },
                    { label: 'Enrollments', to: '/enrollments' },
                  ]
                : []),
              ...(canGrade ? [{ label: 'Grades', to: '/grades' }] : []),
            ],
          },
        ]
      : []),
    { label: 'Analytics', to: '/analytics', icon: BarChart3 },
    { label: 'Notifications', to: '/notifications', icon: Bell },
    ...(canAccessAdmin
      ? [
          {
            label: 'Admin',
            to: '/admin',
            icon: ShieldCheck,
            children: [
              { label: 'Role Management', to: '/admin/roles' },
              { label: 'Audit Logs', to: '/admin/audit-logs' },
            ],
          },
        ]
      : []),
  ]

  return (
    <nav className="flex flex-col gap-1 px-3 py-4">
      <div className="mb-4 flex items-center gap-2 px-2">
        <Building2 className="h-6 w-6 text-primary" />
        <span className="text-lg font-bold text-sidebar-foreground">Myrmex ERP</span>
      </div>

      {navItems.map((item) => {
        const Icon = item.icon
        const isActive =
          matchesPath(pathname, item.to) ||
          !!item.children?.some((c) => matchesPath(pathname, c.to))

        return (
          <div key={item.to}>
            <Link
              to={item.to}
              onClick={onNavigate}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                  : 'text-sidebar-foreground/85 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {item.label}
            </Link>

            {item.children && isActive && (
              <div className="ml-7 mt-1 flex flex-col gap-1">
                {item.children.map((child) => (
                  <Link
                    key={child.to}
                    to={child.to}
                    onClick={onNavigate}
                    className={cn(
                      'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                      pathname === child.to || pathname === child.to + '/'
                        ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                        : 'text-sidebar-foreground/75 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                    )}
                  >
                    {child.label}
                  </Link>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </nav>
  )
}
