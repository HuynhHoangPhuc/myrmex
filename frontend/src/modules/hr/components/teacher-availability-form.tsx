import * as React from 'react'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useUpdateAvailability } from '../hooks/use-teacher-mutations'
import type { TeacherAvailability } from '../types'

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const TIME_SLOTS = [
  { label: '07–09', start: '07:00', end: '09:00' },
  { label: '09–11', start: '09:00', end: '11:00' },
  { label: '11–13', start: '11:00', end: '13:00' },
  { label: '13–15', start: '13:00', end: '15:00' },
  { label: '15–17', start: '15:00', end: '17:00' },
  { label: '17–19', start: '17:00', end: '19:00' },
]

function cellKey(dayIndex: number, slotIndex: number) {
  return `${dayIndex}-${slotIndex}`
}

function slotsToSet(slots: TeacherAvailability[]): Set<string> {
  const set = new Set<string>()
  slots.forEach((slot) => {
    const dayIdx = slot.day_of_week - 1 // 1-indexed → 0-indexed
    const slotIdx = TIME_SLOTS.findIndex(
      (t) => t.start === slot.start_time && t.end === slot.end_time,
    )
    if (slotIdx >= 0) set.add(cellKey(dayIdx, slotIdx))
  })
  return set
}

function setToSlots(set: Set<string>): TeacherAvailability[] {
  const result: TeacherAvailability[] = []
  set.forEach((key) => {
    const [dayStr, slotStr] = key.split('-')
    const dayIdx = parseInt(dayStr, 10)
    const slotIdx = parseInt(slotStr, 10)
    result.push({
      day_of_week: dayIdx + 1,
      start_time: TIME_SLOTS[slotIdx].start,
      end_time: TIME_SLOTS[slotIdx].end,
    })
  })
  return result
}

interface TeacherAvailabilityFormProps {
  teacherId: string
  initialAvailability: TeacherAvailability[]
}

// Weekly grid toggle — click cells to mark teacher availability
export function TeacherAvailabilityForm({ teacherId, initialAvailability }: TeacherAvailabilityFormProps) {
  const [selected, setSelected] = React.useState<Set<string>>(
    () => slotsToSet(initialAvailability),
  )
  const mutation = useUpdateAvailability(teacherId)

  function toggle(dayIdx: number, slotIdx: number) {
    const key = cellKey(dayIdx, slotIdx)
    setSelected((prev) => {
      const next = new Set(prev)
      next.has(key) ? next.delete(key) : next.add(key)
      return next
    })
  }

  function handleSave() {
    mutation.mutate(setToSlots(selected))
  }

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr>
              <th className="w-16 py-2 text-left text-muted-foreground font-normal">Day</th>
              {TIME_SLOTS.map((t) => (
                <th key={t.label} className="px-1 py-2 text-center text-xs text-muted-foreground font-normal">
                  {t.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {DAYS.map((day, dayIdx) => (
              <tr key={day}>
                <td className="py-1 pr-3 text-sm font-medium">{day}</td>
                {TIME_SLOTS.map((_, slotIdx) => {
                  const active = selected.has(cellKey(dayIdx, slotIdx))
                  return (
                    <td key={slotIdx} className="px-1 py-1 text-center">
                      <button
                        type="button"
                        onClick={() => toggle(dayIdx, slotIdx)}
                        className={`h-8 w-full rounded transition-colors ${
                          active
                            ? 'bg-primary text-primary-foreground'
                            : 'bg-muted hover:bg-muted/80'
                        }`}
                        aria-label={`${day} ${TIME_SLOTS[slotIdx].label} ${active ? 'available' : 'unavailable'}`}
                      />
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center gap-3">
        <span className="text-xs text-muted-foreground">{selected.size} slot(s) selected</span>
        <Button size="sm" onClick={handleSave} disabled={mutation.isPending}>
          {mutation.isPending && <LoadingSpinner size="sm" className="mr-2" />}
          Save Availability
        </Button>
      </div>
    </div>
  )
}
