import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSemester } from '@/modules/timetable/hooks/use-semesters'
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

  return (
    <div className="space-y-6">
      <p className="text-sm text-muted-foreground">
        Add the time slots available for scheduling. Use a preset to quickly populate a standard layout,
        or add slots individually.
      </p>
      <TimeSlotManager semesterId={semesterId} slots={semester.time_slots ?? []} />
    </div>
  )
}
