import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { KpiCards } from '@/modules/analytics/components/kpi-cards'
import { WorkloadBarChart } from '@/modules/analytics/components/workload-bar-chart'
import { UtilizationPieChart } from '@/modules/analytics/components/utilization-pie-chart'
import { ScheduleHeatmap } from '@/modules/analytics/components/schedule-heatmap'
import { SemesterFilter } from '@/modules/analytics/components/semester-filter'
import { ExportButton } from '@/modules/analytics/components/export-button'
import { useDashboardSummary } from '@/modules/analytics/hooks/use-dashboard-summary'
import { useWorkloadStats } from '@/modules/analytics/hooks/use-workload-stats'
import { useUtilization } from '@/modules/analytics/hooks/use-utilization'
import { useScheduleHeatmap } from '@/modules/analytics/hooks/use-schedule-heatmap'

export const Route = createFileRoute('/_authenticated/analytics/')({
  component: AnalyticsDashboard,
})

function AnalyticsDashboard() {
  const [semesterId, setSemesterId] = React.useState<string>('')

  const activeSemester = semesterId || undefined

  const { data: summary, isLoading: summaryLoading } = useDashboardSummary()
  const { data: workload, isLoading: workloadLoading } = useWorkloadStats(activeSemester)
  const { data: utilization, isLoading: utilizationLoading } = useUtilization(activeSemester)
  const { data: heatmapData, isLoading: scheduleLoading } = useScheduleHeatmap(activeSemester)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <PageHeader
          title="Analytics"
          description="Faculty workload, slot utilization, and schedule density overview."
        />
        <div className="flex flex-wrap items-center gap-3">
          <SemesterFilter value={semesterId} onChange={setSemesterId} />
          <ExportButton type="workload" semesterId={semesterId || undefined} />
        </div>
      </div>

      {/* KPI summary cards */}
      <KpiCards data={summary} isLoading={summaryLoading} />

      {/* Charts — 2-col on large screens, stacked on mobile */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <WorkloadBarChart data={workload} isLoading={workloadLoading} />
        <UtilizationPieChart data={utilization} isLoading={utilizationLoading} />
      </div>

      {/* Full-width heatmap */}
      <ScheduleHeatmap data={heatmapData} isLoading={scheduleLoading} />
    </div>
  )
}
