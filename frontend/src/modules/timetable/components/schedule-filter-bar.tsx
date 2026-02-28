import * as React from 'react'
import { Input } from '@/components/ui/input'
import type { ScheduleEntry } from '../types'

export interface ScheduleFilters {
  departmentId?: string
  teacherName?: string
  roomId?: string
}

interface ScheduleFilterBarProps {
  entries: ScheduleEntry[]
  filters: ScheduleFilters
  onFilterChange: (filters: ScheduleFilters) => void
}

export function ScheduleFilterBar({ entries, filters, onFilterChange }: ScheduleFilterBarProps) {
  const [teacherName, setTeacherName] = React.useState(filters.teacherName ?? '')

  React.useEffect(() => {
    setTeacherName(filters.teacherName ?? '')
  }, [filters.teacherName])

  React.useEffect(() => {
    const timeout = window.setTimeout(() => {
      if (teacherName !== (filters.teacherName ?? '')) {
        onFilterChange({ ...filters, teacherName: teacherName || undefined })
      }
    }, 300)

    return () => window.clearTimeout(timeout)
  }, [teacherName, filters, onFilterChange])

  const departmentIds = React.useMemo(
    () => Array.from(new Set(entries.map((entry) => entry.department_id))).sort(),
    [entries],
  )
  const roomIds = React.useMemo(
    () => Array.from(new Set(entries.map((entry) => entry.room_id))).sort(),
    [entries],
  )

  const hasFilters = Boolean(filters.departmentId || filters.teacherName || filters.roomId)

  return (
    <div className="grid gap-3 rounded-lg border bg-card p-4 md:grid-cols-[1fr_1fr_1.2fr_auto] md:items-end">
      <div className="space-y-1.5">
        <label className="text-sm font-medium">Department</label>
        <select
          value={filters.departmentId ?? ''}
          onChange={(event) =>
            onFilterChange({
              ...filters,
              departmentId: event.target.value || undefined,
            })
          }
          className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm"
        >
          <option value="">All departments</option>
          {departmentIds.map((departmentId) => (
            <option key={departmentId} value={departmentId}>
              {departmentId}
            </option>
          ))}
        </select>
      </div>

      <div className="space-y-1.5">
        <label className="text-sm font-medium">Room</label>
        <select
          value={filters.roomId ?? ''}
          onChange={(event) =>
            onFilterChange({
              ...filters,
              roomId: event.target.value || undefined,
            })
          }
          className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm"
        >
          <option value="">All rooms</option>
          {roomIds.map((roomId) => (
            <option key={roomId} value={roomId}>
              {roomId}
            </option>
          ))}
        </select>
      </div>

      <div className="space-y-1.5">
        <label className="text-sm font-medium">Teacher</label>
        <Input
          value={teacherName}
          onChange={(event) => setTeacherName(event.target.value)}
          placeholder="Filter by teacher name"
        />
      </div>

      {hasFilters ? (
        <button
          type="button"
          className="h-9 rounded-md border px-3 text-sm font-medium transition-colors hover:bg-accent"
          onClick={() => {
            setTeacherName('')
            onFilterChange({})
          }}
        >
          Clear filters
        </button>
      ) : (
        <div />
      )}
    </div>
  )
}
