import * as React from 'react'
import { EmptyState } from '@/components/shared/empty-state'
import { useToast } from '@/lib/hooks/use-toast'
import { useAssignTeacher } from '../hooks/use-schedules'
import { ScheduleDraggableEntry } from './schedule-draggable-entry'
import { ScheduleDroppableSlot } from './schedule-droppable-slot'
import { ScheduleEntryPopover } from './schedule-entry-popover'
import type { Schedule, ScheduleEntry } from '../types'
import type { ScheduleFilters } from './schedule-filter-bar'
import { periodToTimeLabel } from '../utils/period-to-time'

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const DEPT_COLORS = [
  'bg-blue-100 border-blue-300 text-blue-900 dark:bg-blue-900/30 dark:border-blue-700 dark:text-blue-200',
  'bg-green-100 border-green-300 text-green-900 dark:bg-green-900/30 dark:border-green-700 dark:text-green-200',
  'bg-purple-100 border-purple-300 text-purple-900 dark:bg-purple-900/30 dark:border-purple-700 dark:text-purple-200',
  'bg-orange-100 border-orange-300 text-orange-900 dark:bg-orange-900/30 dark:border-orange-700 dark:text-orange-200',
  'bg-pink-100 border-pink-300 text-pink-900 dark:bg-pink-900/30 dark:border-pink-700 dark:text-pink-200',
  'bg-cyan-100 border-cyan-300 text-cyan-900 dark:bg-cyan-900/30 dark:border-cyan-700 dark:text-cyan-200',
]

function deptColor(deptId: string): string {
  let hash = 0
  for (let i = 0; i < deptId.length; i++) hash = (hash * 31 + deptId.charCodeAt(i)) | 0
  return DEPT_COLORS[Math.abs(hash) % DEPT_COLORS.length]
}

interface ScheduleCalendarProps {
  schedule: Schedule
  filters?: ScheduleFilters
  onChangeTeacher?: (entry: ScheduleEntry) => void
}

// Weekly calendar with mobile card layout, quick actions, and desktop teacher-swap drag/drop.
export function ScheduleCalendar({ schedule, filters, onChangeTeacher }: ScheduleCalendarProps) {
  const { toast } = useToast()
  const assignTeacher = useAssignTeacher(schedule.id)
  const [draggedEntry, setDraggedEntry] = React.useState<ScheduleEntry | null>(null)
  const [hoveredSlot, setHoveredSlot] = React.useState<string | null>(null)

  const filteredEntries = React.useMemo(() => {
    return schedule.entries.filter((entry) => {
      if (filters?.departmentId && entry.department_id !== filters.departmentId) return false
      if (filters?.roomId && entry.room_id !== filters.roomId) return false
      if (
        filters?.teacherName &&
        !entry.teacher_name.toLowerCase().includes(filters.teacherName.toLowerCase())
      ) {
        return false
      }
      return true
    })
  }, [filters, schedule.entries])

  const periodCols = React.useMemo(() => {
    const periods = new Set<number>()
    filteredEntries.forEach((entry) => periods.add(entry.start_period))
    return Array.from(periods).sort((a, b) => a - b)
  }, [filteredEntries])

  const entryMap = React.useMemo(() => {
    const map = new Map<string, ScheduleEntry[]>()
    filteredEntries.forEach((entry) => {
      const key = `${entry.day_of_week}-${entry.start_period}`
      const list = map.get(key) ?? []
      list.push(entry)
      map.set(key, list)
    })
    return map
  }, [filteredEntries])

  const entriesByDay = React.useMemo(() => {
    const map = new Map<number, ScheduleEntry[]>()
    filteredEntries.forEach((entry) => {
      const list = map.get(entry.day_of_week) ?? []
      list.push(entry)
      map.set(entry.day_of_week, list)
    })

    map.forEach((entries) => {
      entries.sort((a, b) => a.start_period - b.start_period)
    })

    return map
  }, [filteredEntries])

  async function handleDrop(targetEntries: ScheduleEntry[]) {
    if (!draggedEntry) return

    setHoveredSlot(null)

    if (targetEntries.length !== 1) {
      toast({
        title: 'Drop onto a single entry',
        description: 'Empty slots and multi-entry swaps are not supported yet.',
        variant: 'destructive',
      })
      setDraggedEntry(null)
      return
    }

    const targetEntry = targetEntries[0]
    if (targetEntry.id === draggedEntry.id) {
      setDraggedEntry(null)
      return
    }

    try {
      await assignTeacher.mutateAsync({
        entry_id: targetEntry.id,
        teacher_id: draggedEntry.teacher_id,
      })
      await assignTeacher.mutateAsync({
        entry_id: draggedEntry.id,
        teacher_id: targetEntry.teacher_id,
      })

      toast({
        title: 'Teacher swap saved',
        description:
          schedule.hard_violations > 0
            ? 'Recheck hard violations after the schedule refreshes.'
            : 'Assignments updated successfully.',
        variant: schedule.hard_violations > 0 ? 'destructive' : 'default',
      })
    } catch {
      toast({ title: 'Failed to swap teachers', variant: 'destructive' })
    } finally {
      setDraggedEntry(null)
    }
  }

  if (filteredEntries.length === 0) {
    return (
      <EmptyState
        title={schedule.entries.length === 0 ? 'No schedule entries yet.' : 'No entries match your filters.'}
        description={
          schedule.entries.length === 0
            ? 'Generate a schedule to see timetable entries here.'
            : 'Adjust the filters to see more results.'
        }
      />
    )
  }

  return (
    <div className="space-y-4">
      <div className="space-y-3 md:hidden">
        {DAY_NAMES.map((day, dayIndex) => {
          const entries = entriesByDay.get(dayIndex + 1) ?? []
          if (entries.length === 0) return null

          return (
            <div key={day} className="rounded-lg border bg-card p-4">
              <h3 className="text-sm font-semibold">{day}</h3>
              <div className="mt-3 space-y-2">
                {entries.map((entry) => (
                  <div key={entry.id} className="space-y-1 rounded-md border p-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="font-medium">{entry.subject_code}</p>
                        <p className="text-sm text-muted-foreground">{entry.teacher_name}</p>
                        <p className="text-xs text-muted-foreground">{entry.room_name}</p>
                      </div>
                      <button
                        type="button"
                        className="rounded-md border px-2 py-1 text-xs font-medium"
                        onClick={() => onChangeTeacher?.(entry)}
                      >
                        Change
                      </button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      P{entry.start_period}–P{entry.end_period} · {periodToTimeLabel(entry.start_period, entry.end_period)}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          )
        })}
      </div>

      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-[700px] w-full border-collapse text-sm">
          <thead>
            <tr>
              <th className="w-14 border-b py-2 text-left text-xs font-normal text-muted-foreground">Day</th>
              {periodCols.map((period) => (
                <th
                  key={period}
                  className="border-b px-1 py-2 text-center text-xs font-normal text-muted-foreground"
                >
                  P{period}
                  <br />
                  <span className="font-normal opacity-70">{periodToTimeLabel(period, period)}</span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {DAY_NAMES.map((day, dayIndex) => {
              const dayOfWeek = dayIndex + 1
              return (
                <tr key={day} className="border-b last:border-0">
                  <td className="py-2 pr-2 align-top text-xs font-semibold text-muted-foreground">{day}</td>
                  {periodCols.map((period) => {
                    const cellKey = `${dayOfWeek}-${period}`
                    const entries = entryMap.get(cellKey) ?? []
                    return (
                      <td key={period} className="min-w-[110px] px-1 py-1 align-top">
                        <ScheduleDroppableSlot
                          isOver={hoveredSlot === cellKey}
                          onDragOver={(event) => {
                            event.preventDefault()
                            setHoveredSlot(cellKey)
                          }}
                          onDragLeave={() => setHoveredSlot((current) => (current === cellKey ? null : current))}
                          onDrop={(event) => {
                            event.preventDefault()
                            void handleDrop(entries)
                          }}
                        >
                          <div className="space-y-1">
                            {entries.map((entry) => (
                              <ScheduleDraggableEntry
                                key={entry.id}
                                entry={entry}
                                isDragging={draggedEntry?.id === entry.id}
                                onDragStart={setDraggedEntry}
                                onDragEnd={() => {
                                  setDraggedEntry(null)
                                  setHoveredSlot(null)
                                }}
                              >
                                <ScheduleEntryPopover
                                  entry={entry}
                                  colorClassName={deptColor(entry.department_id)}
                                  onChangeTeacher={onChangeTeacher}
                                />
                              </ScheduleDraggableEntry>
                            ))}
                          </div>
                        </ScheduleDroppableSlot>
                      </td>
                    )
                  })}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
