import { Link, useRouterState } from '@tanstack/react-router'
import {
  LayoutDashboard,
  Users,
  BookOpen,
  Calendar,
  Building2,
} from 'lucide-react'
import { cn } from '@/lib/utils/cn'

interface NavItem {
  label: string
  to: string
  icon: React.ElementType
  children?: { label: string; to: string }[]
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', to: '/dashboard', icon: LayoutDashboard },
  {
    label: 'HR',
    to: '/hr',
    icon: Users,
    children: [
      { label: 'Teachers', to: '/hr/teachers' },
      { label: 'Departments', to: '/hr/departments' },
    ],
  },
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
      { label: 'Generate', to: '/timetable/generate' },
      { label: 'Assign Teachers', to: '/timetable/assign' },
    ],
  },
]

// Sidebar navigation with active-state highlighting and nested items
export function SidebarNav() {
  const router = useRouterState()
  const pathname = router.location.pathname

  return (
    <nav className="flex flex-col gap-1 px-3 py-4">
      <div className="mb-4 flex items-center gap-2 px-2">
        <Building2 className="h-6 w-6 text-primary" />
        <span className="text-lg font-bold text-sidebar-foreground">Myrmex ERP</span>
      </div>

      {NAV_ITEMS.map((item) => {
        const Icon = item.icon
        const isActive = pathname === item.to || pathname.startsWith(item.to + '/')

        return (
          <div key={item.to}>
            <Link
              to={item.to}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {item.label}
            </Link>

            {/* Nested items — shown when parent is active */}
            {item.children && isActive && (
              <div className="ml-7 mt-1 flex flex-col gap-1">
                {item.children.map((child) => (
                  <Link
                    key={child.to}
                    to={child.to}
                    className={cn(
                      'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                      pathname === child.to
                        ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                        : 'text-sidebar-foreground/60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
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
