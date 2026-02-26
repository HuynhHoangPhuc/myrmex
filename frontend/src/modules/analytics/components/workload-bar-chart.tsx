import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingPage } from '@/components/shared/loading-spinner'
import type { WorkloadStat } from '../types'

interface WorkloadBarChartProps {
  data?: WorkloadStat[]
  isLoading: boolean
}

// Deterministic color per department ID
function deptColor(name: string): string {
  const palette = [
    '#3b82f6', '#8b5cf6', '#10b981', '#f59e0b',
    '#ef4444', '#06b6d4', '#ec4899', '#84cc16',
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) & 0xffffffff
  return palette[Math.abs(hash) % palette.length]
}

// Horizontal bar chart — Y axis = teacher names, X axis = total hours
export function WorkloadBarChart({ data, isLoading }: WorkloadBarChartProps) {
  if (isLoading) return <LoadingPage />

  const chartData = (data ?? [])
    .slice()
    .sort((a, b) => b.total_hours - a.total_hours)
    .slice(0, 20) // cap at 20 teachers to keep chart readable

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Teacher Workload (Hours)</CardTitle>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No workload data available.</p>
        ) : (
          <ResponsiveContainer width="100%" height={Math.max(chartData.length * 36, 200)}>
            <BarChart data={chartData} layout="vertical" margin={{ left: 8, right: 24, top: 4, bottom: 4 }}>
              <CartesianGrid strokeDasharray="3 3" horizontal={false} />
              <XAxis type="number" unit="h" tick={{ fontSize: 12 }} />
              <YAxis
                type="category"
                dataKey="teacher_name"
                width={140}
                tick={{ fontSize: 11 }}
                tickLine={false}
              />
              <Tooltip
                formatter={(v: number | undefined) => [`${v ?? 0}h`, 'Total Hours']}
              />
              <Bar dataKey="total_hours" radius={[0, 4, 4, 0]}>
                {chartData.map((entry) => (
                  <Cell key={entry.teacher_id} fill={deptColor(entry.department_id)} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
