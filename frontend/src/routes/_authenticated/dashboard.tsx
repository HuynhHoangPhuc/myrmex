import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { Users, BookOpen, Calendar, Building2, Sparkles, Clock3 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/shared/empty-state'
import { PageHeader } from '@/components/shared/page-header'
import { useCurrentUser } from '@/lib/hooks/use-current-user'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { DashboardStats, ListResponse } from '@/lib/api/types'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import type { Schedule } from '@/modules/timetable/types'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data: user } = useCurrentUser()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: semesters = [], isLoading: semestersLoading } = useAllSemesters()
  const { data: pendingSchedules = [], isLoading: schedulesLoading } = usePendingSchedules()

  const recentSemesters = React.useMemo(
    () =>
      semesters
        .slice()
        .sort((left, right) => Date.parse(right.created_at) - Date.parse(left.created_at))
        .slice(0, 3),
    [semesters],
  )

  const statCards = [
    {
      label: 'Total Teachers',
      value: stats?.total_teachers ?? 0,
      icon: Users,
      color: 'text-blue-600 dark:text-blue-300',
      bg: 'bg-blue-50 dark:bg-blue-950/40',
    },
    {
      label: 'Departments',
      value: stats?.total_departments ?? 0,
      icon: Building2,
      color: 'text-purple-600 dark:text-purple-300',
      bg: 'bg-purple-50 dark:bg-purple-950/40',
    },
    {
      label: 'Total Subjects',
      value: stats?.total_subjects ?? 0,
      icon: BookOpen,
      color: 'text-green-600 dark:text-green-300',
      bg: 'bg-green-50 dark:bg-green-950/40',
    },
    {
      label: 'Active Semesters',
      value: stats?.active_semesters ?? 0,
      icon: Calendar,
      color: 'text-orange-600 dark:text-orange-300',
      bg: 'bg-orange-50 dark:bg-orange-950/40',
    },
  ]

  return (
    <div className="space-y-8">
      <PageHeader
        title={`Welcome back, ${user?.full_name?.split(' ')[0] ?? 'User'}`}
        description="Here’s what needs attention across semesters, scheduling, and planning."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.label}>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.label}
                </CardTitle>
                <div className={`rounded-lg p-2 ${stat.bg}`}>
                  <Icon className={`h-4 w-4 ${stat.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                {statsLoading ? (
                  <div className="h-8 w-16 animate-pulse rounded bg-muted" />
                ) : (
                  <p className="text-2xl font-bold">{stat.value.toLocaleString()}</p>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Sparkles className="h-4 w-4 text-primary" />
              Recent semesters
            </CardTitle>
          </CardHeader>
          <CardContent>
            {semestersLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 3 }).map((_, index) => (
                  <div key={index} className="h-12 animate-pulse rounded-md bg-muted" />
                ))}
              </div>
            ) : recentSemesters.length === 0 ? (
              <EmptyState
                title="No semesters yet"
                description="Create your first semester to start planning schedules."
                action={
                  <Button asChild size="sm">
                    <Link to="/timetable/semesters/new" search={{ step: 1 }}>
                      New semester
                    </Link>
                  </Button>
                }
                className="py-6"
              />
            ) : (
              <div className="space-y-3">
                {recentSemesters.map((semester) => (
                  <div
                    key={semester.id}
                    className="flex flex-col gap-2 rounded-lg border p-3 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div className="min-w-0">
                      <Link
                        to="/timetable/semesters/$id"
                        params={{ id: semester.id }}
                        className="font-medium text-primary hover:underline"
                      >
                        {semester.name}
                      </Link>
                      <p className="text-sm text-muted-foreground">{semester.academic_year}</p>
                    </div>
                    <Badge variant={semester.is_active ? 'secondary' : 'outline'}>
                      {semester.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Clock3 className="h-4 w-4 text-primary" />
              Pending schedules
            </CardTitle>
          </CardHeader>
          <CardContent>
            {schedulesLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 3 }).map((_, index) => (
                  <div key={index} className="h-12 animate-pulse rounded-md bg-muted" />
                ))}
              </div>
            ) : pendingSchedules.length === 0 ? (
              <EmptyState
                title="No pending schedules"
                description="All current schedule jobs are complete."
                className="py-6"
              />
            ) : (
              <div className="space-y-3">
                {pendingSchedules.map((schedule) => (
                  <div
                    key={schedule.id}
                    className="flex flex-col gap-2 rounded-lg border p-3 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div className="min-w-0">
                      <Link
                        to="/timetable/schedules/$id"
                        params={{ id: schedule.id }}
                        className="font-medium text-primary hover:underline"
                      >
                        Schedule {schedule.id.slice(0, 8)}
                      </Link>
                      <p className="text-sm text-muted-foreground">
                        Created {new Date(schedule.created_at).toLocaleString()}
                      </p>
                    </div>
                    <Badge variant={schedule.status === 'generating' ? 'secondary' : 'outline'}>
                      {schedule.status}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Quick actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button asChild>
              <Link to="/timetable/semesters/new" search={{ step: 1 }}>
                New semester wizard
              </Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/subjects/offerings">Semester offerings</Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/timetable/generate">Generate schedule</Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/timetable/assign">Assign teachers</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard-stats'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<DashboardStats>(ENDPOINTS.dashboard.stats)
      return data
    },
  })
}

function usePendingSchedules() {
  return useQuery({
    queryKey: ['dashboard-pending-schedules'] as const,
    queryFn: async () => {
      const pageSize = 100
      const pendingSchedules: Schedule[] = []
      let page = 1
      let total = 0

      while (page === 1 || (page - 1) * pageSize < total) {
        const { data } = await apiClient.get<ListResponse<Schedule>>(ENDPOINTS.timetable.schedules, {
          params: { page, page_size: pageSize },
        })

        total = data.total
        pendingSchedules.push(
          ...data.data.filter(
            (schedule) => schedule.status === 'pending' || schedule.status === 'generating',
          ),
        )
        page += 1
      }

      return pendingSchedules
        .sort((left, right) => Date.parse(right.created_at) - Date.parse(left.created_at))
        .slice(0, 4)
    },
  })
}
