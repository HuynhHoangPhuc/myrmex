import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSemester } from '@/modules/timetable/hooks/use-semesters'
import { RoomManager } from './room-manager'
import { TimeSlotManager } from './time-slot-manager'

interface SemesterWizardStepSlotsProps {
  semesterId: string
}

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

  const selectedRoomIds = semester.rooms?.map((r) => r.id) ?? []

  return (
    <div className="space-y-8">
      {/* Time slots section */}
      <div className="space-y-3">
        <div>
          <h3 className="text-base font-semibold">Time Slots</h3>
          <p className="text-sm text-muted-foreground">
            Add the time slots available for scheduling. Use a preset to quickly populate a standard
            layout, or add slots individually.
          </p>
        </div>
        <TimeSlotManager semesterId={semesterId} slots={semester.time_slots ?? []} />
      </div>

      {/* Rooms section */}
      <div className="space-y-3">
        <div>
          <h3 className="text-base font-semibold">Rooms</h3>
          <p className="text-sm text-muted-foreground">
            Select which rooms are available for scheduling in this semester. Only checked rooms
            will be considered when generating the timetable.
          </p>
        </div>
        <RoomManager semesterId={semesterId} selectedRoomIds={selectedRoomIds} />
      </div>
    </div>
  )
}
