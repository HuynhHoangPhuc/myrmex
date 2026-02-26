import { Users, BookOpen, Calendar, Building2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import type { DashboardSummary } from '../types'

interface KpiCardsProps {
  data?: DashboardSummary
  isLoading: boolean
}

const KPI_CONFIG = [
  { key: 'total_teachers' as const, label: 'Total Teachers', icon: Users, color: 'text-blue-600', bg: 'bg-blue-50' },
  { key: 'total_departments' as const, label: 'Total Departments', icon: Building2, color: 'text-indigo-600', bg: 'bg-indigo-50' },
  { key: 'total_subjects' as const, label: 'Total Subjects', icon: BookOpen, color: 'text-green-600', bg: 'bg-green-50' },
  { key: 'total_semesters' as const, label: 'Total Semesters', icon: Calendar, color: 'text-purple-600', bg: 'bg-purple-50' },
]

// Grid of KPI stat cards summarising the analytics dashboard
export function KpiCards({ data, isLoading }: KpiCardsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
      {KPI_CONFIG.map(({ key, label, icon: Icon, color, bg }) => {
        const raw = data?.[key] ?? 0
        const display = raw.toLocaleString()

        return (
          <Card key={key}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
              <div className={`rounded-lg p-2 ${bg}`}>
                <Icon className={`h-4 w-4 ${color}`} />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <LoadingSpinner size="sm" />
              ) : (
                <p className="text-2xl font-bold">{display}</p>
              )}
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
