import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import type { Schedule, ScheduleEntry } from '../types'
import { periodToTimeLabel } from '../utils/period-to-time'

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

// Pastel color pool for department colour-coding
const DEPT_COLORS = [
  'bg-blue-100 border-blue-300 text-blue-900',
  'bg-green-100 border-green-300 text-green-900',
  'bg-purple-100 border-purple-300 text-purple-900',
  'bg-orange-100 border-orange-300 text-orange-900',
  'bg-pink-100 border-pink-300 text-pink-900',
  'bg-cyan-100 border-cyan-300 text-cyan-900',
]

function deptColor(deptId: string): string {
  let hash = 0
  for (let i = 0; i < deptId.length; i++) hash = (hash * 31 + deptId.charCodeAt(i)) | 0
  return DEPT_COLORS[Math.abs(hash) % DEPT_COLORS.length]
}

interface EntryCardProps {
  entry: ScheduleEntry
  onClick: (entry: ScheduleEntry) => void
}

function EntryCard({ entry, onClick }: EntryCardProps) {
  const color = deptColor(entry.department_id)
  return (
    <button
      type="button"
      onClick={() => onClick(entry)}
      className={`w-full rounded border p-1.5 text-left text-xs space-y-0.5 ${color} ${entry.is_manual_override ? 'ring-2 ring-yellow-400' : ''} hover:opacity-80 transition-opacity`}
    >
      <p className="font-bold truncate">{entry.subject_code}</p>
      <p className="truncate opacity-75">{entry.teacher_name}</p>
      <p className="truncate opacity-60">{entry.room_name}</p>
      {entry.is_manual_override && (
        <Badge variant="outline" className="text-[10px] px-1 py-0 border-yellow-500 text-yellow-700">override</Badge>
      )}
    </button>
  )
}

interface ScheduleCalendarProps {
  schedule: Schedule
  onEntryClick?: (entry: ScheduleEntry) => void
}

// CSS-grid based weekly calendar — days as rows, period slots as columns
export function ScheduleCalendar({ schedule, onEntryClick }: ScheduleCalendarProps) {
  // Collect unique start_period values and derive time labels
  const periodCols = React.useMemo(() => {
    const periods = new Set<number>()
    schedule.entries.forEach((e) => periods.add(e.start_period))
    return Array.from(periods).sort((a, b) => a - b)
  }, [schedule.entries])

  // Index entries by day + start_period for O(1) lookup
  const entryMap = React.useMemo(() => {
    const map = new Map<string, ScheduleEntry[]>()
    schedule.entries.forEach((e) => {
      const key = `${e.day_of_week}-${e.start_period}`
      const list = map.get(key) ?? []
      list.push(e)
      map.set(key, list)
    })
    return map
  }, [schedule.entries])

  if (schedule.entries.length === 0) {
    return <p className="text-sm text-muted-foreground py-8 text-center">No schedule entries yet.</p>
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-sm min-w-[700px]">
        <thead>
          <tr>
            <th className="w-14 py-2 text-left text-xs text-muted-foreground font-normal border-b">Day</th>
            {periodCols.map((period) => (
              <th key={period} className="px-1 py-2 text-center text-xs text-muted-foreground font-normal border-b">
                P{period}<br />
                <span className="font-normal opacity-70">{periodToTimeLabel(period, period)}</span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {DAY_NAMES.map((day, dayIdx) => {
            const dayOfWeek = dayIdx + 1
            return (
              <tr key={day} className="border-b last:border-0">
                <td className="py-2 pr-2 text-xs font-semibold text-muted-foreground align-top">{day}</td>
                {periodCols.map((period) => {
                  const cellKey = `${dayOfWeek}-${period}`
                  const entries = entryMap.get(cellKey) ?? []
                  return (
                    <td key={period} className="px-1 py-1 align-top min-w-[100px]">
                      <div className="space-y-1">
                        {entries.map((entry) => (
                          <EntryCard
                            key={entry.id}
                            entry={entry}
                            onClick={onEntryClick ?? (() => {})}
                          />
                        ))}
                      </div>
                    </td>
                  )
                })}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
