import * as React from 'react'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'

interface SemesterFilterProps {
  value: string
  onChange: (semesterId: string) => void
}

// Dropdown for selecting a semester — feeds semesterId into all chart queries
export function SemesterFilter({ value, onChange }: SemesterFilterProps) {
  const { data: semesters, isLoading } = useAllSemesters()

  return (
    <div className="flex items-center gap-2">
      <label htmlFor="semester-filter" className="text-sm font-medium text-muted-foreground">
        Semester
      </label>
      <select
        id="semester-filter"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={isLoading}
        className="rounded-md border border-input bg-background px-3 py-1.5 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50"
      >
        <option value="">All Semesters</option>
        {semesters?.map((s) => (
          <option key={s.id} value={s.id}>
            {s.name}
          </option>
        ))}
      </select>
    </div>
  )
}
