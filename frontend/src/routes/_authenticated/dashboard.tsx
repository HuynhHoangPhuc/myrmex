import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Users, BookOpen, Calendar, Building2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useCurrentUser } from '@/lib/hooks/use-current-user'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { DashboardStats } from '@/lib/api/types'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data: user } = useCurrentUser()
  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const { data } = await apiClient.get<DashboardStats>(ENDPOINTS.dashboard.stats)
      return data
    },
  })

  const statCards = [
    {
      label: 'Total Teachers',
      value: stats?.total_teachers ?? 0,
      icon: Users,
      href: '/hr/teachers',
      color: 'text-blue-600',
      bg: 'bg-blue-50',
    },
    {
      label: 'Departments',
      value: stats?.total_departments ?? 0,
      icon: Building2,
      href: '/hr/departments',
      color: 'text-purple-600',
      bg: 'bg-purple-50',
    },
    {
      label: 'Total Subjects',
      value: stats?.total_subjects ?? 0,
      icon: BookOpen,
      href: '/subjects',
      color: 'text-green-600',
      bg: 'bg-green-50',
    },
    {
      label: 'Active Semesters',
      value: stats?.active_semesters ?? 0,
      icon: Calendar,
      href: '/timetable',
      color: 'text-orange-600',
      bg: 'bg-orange-50',
    },
  ]

  return (
    <div>
      <PageHeader
        title={`Welcome back, ${user?.full_name?.split(' ')[0] ?? 'User'}`}
        description="Here's an overview of your faculty management system."
      />

      {/* Stats grid */}
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
                {isLoading ? (
                  <LoadingSpinner size="sm" />
                ) : (
                  <p className="text-2xl font-bold">{stat.value.toLocaleString()}</p>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Quick actions */}
      <div className="mt-8">
        <h2 className="mb-4 text-lg font-semibold">Quick Actions</h2>
        <div className="flex flex-wrap gap-3">
          <Button asChild variant="outline">
            <a href="/hr/teachers">Manage Teachers</a>
          </Button>
          <Button asChild variant="outline">
            <a href="/subjects">Manage Subjects</a>
          </Button>
          <Button asChild>
            <a href="/timetable">Generate Schedule</a>
          </Button>
        </div>
      </div>
    </div>
  )
}
