import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingPage } from '@/components/shared/loading-spinner'
import type { ScheduleHeatmapCell } from '../types'

interface ScheduleHeatmapProps {
  data?: ScheduleHeatmapCell[]
  isLoading: boolean
}

const DAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const PERIODS = [1, 2, 3, 4, 5, 6, 7, 8]

// Maps entry count to a tailwind bg color — higher = darker blue
function intensityClass(count: number, max: number): string {
  if (max === 0 || count === 0) return 'bg-muted'
  const ratio = count / max
  if (ratio < 0.2) return 'bg-blue-100'
  if (ratio < 0.4) return 'bg-blue-200'
  if (ratio < 0.6) return 'bg-blue-400'
  if (ratio < 0.8) return 'bg-blue-600'
  return 'bg-blue-800'
}

function intensityTextClass(count: number, max: number): string {
  if (max === 0 || count === 0) return 'text-muted-foreground'
  const ratio = count / max
  return ratio >= 0.6 ? 'text-white' : 'text-foreground'
}

// CSS grid heatmap — rows = days (1-6), cols = periods (1-8)
export function ScheduleHeatmap({ data, isLoading }: ScheduleHeatmapProps) {
  if (isLoading) return <LoadingPage />

  // Build a lookup map: [day][period] = entry_count
  const grid: Record<number, Record<number, number>> = {}
  let max = 0
  for (const m of data ?? []) {
    grid[m.day_of_week] ??= {}
    grid[m.day_of_week][m.period] = m.entry_count
    if (m.entry_count > max) max = m.entry_count
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Schedule Density (Day × Period)</CardTitle>
      </CardHeader>
      <CardContent className="overflow-x-auto">
        <div
          className="grid gap-1 min-w-[480px]"
          style={{ gridTemplateColumns: `56px repeat(${PERIODS.length}, 1fr)` }}
        >
          {/* Header row */}
          <div />
          {PERIODS.map((p) => (
            <div key={p} className="py-1 text-center text-xs font-medium text-muted-foreground">
              P{p}
            </div>
          ))}

          {/* Data rows — one per day */}
          {DAY_LABELS.map((day, idx) => {
            const dayNum = idx + 1
            return [
              <div key={`label-${dayNum}`} className="flex items-center text-xs font-medium text-muted-foreground">
                {day}
              </div>,
              ...PERIODS.map((p) => {
                const count = grid[dayNum]?.[p] ?? 0
                return (
                  <div
                    key={`${dayNum}-${p}`}
                    title={`${day} P${p}: ${count} entries`}
                    className={`flex h-9 items-center justify-center rounded text-xs font-medium transition-colors ${intensityClass(count, max)} ${intensityTextClass(count, max)}`}
                  >
                    {count > 0 ? count : ''}
                  </div>
                )
              }),
            ]
          })}
        </div>

        {/* Legend */}
        <div className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
          <span>Low</span>
          {['bg-blue-100', 'bg-blue-200', 'bg-blue-400', 'bg-blue-600', 'bg-blue-800'].map((cls) => (
            <div key={cls} className={`h-3 w-5 rounded ${cls}`} />
          ))}
          <span>High</span>
        </div>
      </CardContent>
    </Card>
  )
}
