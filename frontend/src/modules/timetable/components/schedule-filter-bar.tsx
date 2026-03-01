import * as React from 'react'
import { Input } from '@/components/ui/input'
import { useAllDepartments } from '@/modules/hr/hooks/use-departments'
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
  const { data: allDepartments } = useAllDepartments()

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

  // Build department id→name map from fetched departments
  const departmentNameMap = React.useMemo(() => {
    const map = new Map<string, string>()
    allDepartments?.forEach((d) => map.set(d.id, d.name))
    return map
  }, [allDepartments])

  // Unique department IDs that actually appear in schedule entries
  const departmentIds = React.useMemo(
    () => Array.from(new Set(entries.map((e) => e.department_id).filter(Boolean))).sort(),
    [entries],
  )

  // Build room id→name map directly from entry data (room_name is already available)
  const roomOptions = React.useMemo(() => {
    const map = new Map<string, string>()
    entries.forEach((e) => {
      if (e.room_id) map.set(e.room_id, e.room_name || e.room_id)
    })
    return Array.from(map.entries()).sort((a, b) => a[1].localeCompare(b[1]))
  }, [entries])

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
          {departmentIds.map((id) => (
            <option key={id} value={id}>
              {departmentNameMap.get(id) ?? id}
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
          {roomOptions.map(([id, name]) => (
            <option key={id} value={id}>
              {name}
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
