import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingPage } from '@/components/shared/loading-spinner'
import type { UtilizationStat } from '../types'

interface UtilizationPieChartProps {
  data?: UtilizationStat[]
  isLoading: boolean
}

const COLORS = [
  '#3b82f6', '#8b5cf6', '#10b981', '#f59e0b',
  '#ef4444', '#06b6d4', '#ec4899', '#84cc16',
]

// Donut chart showing slot utilization % per department
export function UtilizationPieChart({ data, isLoading }: UtilizationPieChartProps) {
  if (isLoading) return <LoadingPage />

  const chartData = (data ?? []).map((d) => ({
    name: d.department_name,
    value: d.utilization_pct,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Slot Utilization by Department (%)</CardTitle>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No utilization data available.</p>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={chartData}
                cx="50%"
                cy="50%"
                innerRadius={70}
                outerRadius={110}
                paddingAngle={3}
                dataKey="value"
                label={({ name, value }) => `${name} ${(value as number).toFixed(1)}%`}
                labelLine={false}
              >
                {chartData.map((_, index) => (
                  <Cell key={index} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip formatter={(v: number | undefined) => [`${(v ?? 0).toFixed(1)}%`, 'Utilization']} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
