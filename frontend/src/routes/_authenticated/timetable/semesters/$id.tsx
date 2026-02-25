import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Calendar } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSemester } from '@/modules/timetable/hooks/use-semesters'

export const Route = createFileRoute('/_authenticated/timetable/semesters/$id')({
  component: SemesterDetailPage,
})

function SemesterDetailPage() {
  const { id } = Route.useParams()
  const { data: semester, isLoading } = useSemester(id)

  if (isLoading) return <LoadingSpinner />
  if (!semester) return <p className="text-muted-foreground">Semester not found.</p>

  return (
    <div className="max-w-4xl space-y-8">
      <PageHeader
        title={semester.name}
        description={`${semester.academic_year} · ${semester.start_date} → ${semester.end_date}`}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" asChild>
              <Link to="/timetable/semesters" search={{ page: 1, pageSize: 25 }}>
                <ArrowLeft className="mr-2 h-4 w-4" />Back
              </Link>
            </Button>
            <Button asChild>
              <Link to="/timetable/generate" search={{ semesterId: id }}>
                <Calendar className="mr-2 h-4 w-4" />Generate Schedule
              </Link>
            </Button>
          </div>
        }
      />

      <div className="flex gap-2">
        {semester.is_active ? (
          <Badge variant="secondary">Active</Badge>
        ) : (
          <Badge variant="outline">Inactive</Badge>
        )}
      </div>

      {/* Time slots */}
      <div>
        <h2 className="mb-3 text-lg font-semibold">Time Slots ({semester.time_slots?.length ?? 0})</h2>
        {!semester.time_slots?.length ? (
          <p className="text-sm text-muted-foreground">No time slots defined.</p>
        ) : (
          <div className="rounded-md border divide-y">
            {semester.time_slots.map((slot) => (
              <div key={slot.id} className="flex items-center gap-4 px-4 py-2.5 text-sm">
                <span className="w-24 font-medium">
                  {['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][slot.day_of_week]}
                </span>
                <span className="text-muted-foreground">
                  {slot.start_time} – {slot.end_time}
                </span>
                <Badge variant="outline" className="ml-auto text-xs">Slot {slot.slot_index + 1}</Badge>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Rooms */}
      <div>
        <h2 className="mb-3 text-lg font-semibold">Rooms ({semester.rooms?.length ?? 0})</h2>
        {!semester.rooms?.length ? (
          <p className="text-sm text-muted-foreground">No rooms defined.</p>
        ) : (
          <div className="rounded-md border divide-y">
            {semester.rooms.map((room) => (
              <div key={room.id} className="flex items-center gap-4 px-4 py-2.5 text-sm">
                <span className="font-medium w-32">{room.name}</span>
                <span className="text-muted-foreground">{room.capacity} seats</span>
                <Badge variant="secondary" className="ml-auto capitalize">{room.room_type}</Badge>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
