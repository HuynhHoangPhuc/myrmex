import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSemester } from '@/modules/timetable/hooks/use-semesters'

interface SemesterWizardStepSlotsProps {
  semesterId: string
}

const DAY_NAMES = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export function SemesterWizardStepSlots({ semesterId }: SemesterWizardStepSlotsProps) {
  const { data: semester, isLoading } = useSemester(semesterId)

  if (isLoading) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <LoadingSpinner />
      </div>
    )
  }

  if (!semester) {
    return <p className="text-sm text-muted-foreground">Semester not found.</p>
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-4">
        <h3 className="font-semibold">Configure time slots and rooms</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Incremental slot and room creation is not available yet. Review the current setup, then use the semester detail page to manage the full configuration.
        </p>
        <Button asChild variant="outline" className="mt-4">
          <Link to="/timetable/semesters/$id" params={{ id: semesterId }}>
            Open semester details
          </Link>
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="space-y-3">
          <h4 className="text-sm font-semibold">Current time slots</h4>
          {!semester.time_slots.length ? (
            <p className="rounded-lg border border-dashed px-4 py-6 text-sm text-muted-foreground">
              No time slots defined yet.
            </p>
          ) : (
            <div className="rounded-lg border divide-y">
              {semester.time_slots.map((slot) => (
                <div key={slot.id} className="flex items-center gap-3 px-4 py-3 text-sm">
                  <span className="w-12 font-medium">{DAY_NAMES[slot.day_of_week]}</span>
                  <span className="text-muted-foreground">
                    {slot.start_time} – {slot.end_time}
                  </span>
                  <Badge variant="outline" className="ml-auto text-xs">
                    Slot {slot.slot_index + 1}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="space-y-3">
          <h4 className="text-sm font-semibold">Current rooms</h4>
          {!semester.rooms.length ? (
            <p className="rounded-lg border border-dashed px-4 py-6 text-sm text-muted-foreground">
              No rooms defined yet.
            </p>
          ) : (
            <div className="rounded-lg border divide-y">
              {semester.rooms.map((room) => (
                <div key={room.id} className="flex items-center gap-3 px-4 py-3 text-sm">
                  <span className="font-medium">{room.name}</span>
                  <span className="text-muted-foreground">{room.capacity} seats</span>
                  <Badge variant="secondary" className="ml-auto capitalize">
                    {room.room_type}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
