import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, UserCog } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { ScheduleCalendar } from '@/modules/timetable/components/schedule-calendar'
import { TeacherAssignmentDialog } from '@/modules/timetable/components/teacher-assignment'
import { useSchedule } from '@/modules/timetable/hooks/use-schedules'
import type { ScheduleEntry } from '@/modules/timetable/types'

export const Route = createFileRoute('/_authenticated/timetable/schedules/$id/')({
  component: ScheduleDetailPage,
})

function ScheduleDetailPage() {
  const { id } = Route.useParams()
  const { data: schedule, isLoading } = useSchedule(id)
  const [selectedEntry, setSelectedEntry] = React.useState<ScheduleEntry | null>(null)
  const [assignOpen, setAssignOpen] = React.useState(false)

  function handleEntryClick(entry: ScheduleEntry) {
    setSelectedEntry(entry)
    setAssignOpen(true)
  }

  if (isLoading) return <LoadingSpinner />
  if (!schedule) return <p className="text-muted-foreground">Schedule not found.</p>

  const overrideCount = schedule.entries.filter((e) => e.is_manual_override).length

  return (
    <div className="space-y-6">
      <PageHeader
        title="Schedule View"
        description={`Generated ${new Date(schedule.created_at).toLocaleString()}`}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" asChild>
              <Link to="/timetable/schedules" search={{ page: 1, pageSize: 25 }}>
                <ArrowLeft className="mr-2 h-4 w-4" />Back
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/timetable/assign" search={{ scheduleId: id }}>
                <UserCog className="mr-2 h-4 w-4" />Assign Teachers
              </Link>
            </Button>
          </div>
        }
      />

      {/* Stats bar */}
      <div className="flex flex-wrap gap-3">
        <Badge variant="secondary">Status: {schedule.status}</Badge>
        {schedule.status === 'completed' && (
          <>
            <Badge variant="secondary">Score: {schedule.score.toFixed(2)}</Badge>
            <Badge variant={schedule.hard_violations === 0 ? 'secondary' : 'destructive'}>
              {schedule.hard_violations} hard violations
            </Badge>
            <Badge variant="outline">{schedule.soft_violations} soft violations</Badge>
            <Badge variant="outline">{schedule.entries.length} entries</Badge>
            {overrideCount > 0 && (
              <Badge variant="outline" className="border-yellow-500 text-yellow-700">
                {overrideCount} manual overrides
              </Badge>
            )}
          </>
        )}
      </div>

      {/* Calendar grid — click entry to open assignment dialog */}
      <ScheduleCalendar schedule={schedule} onEntryClick={handleEntryClick} />

      <TeacherAssignmentDialog
        scheduleId={id}
        entry={selectedEntry}
        open={assignOpen}
        onOpenChange={(o) => { setAssignOpen(o); if (!o) setSelectedEntry(null) }}
      />
    </div>
  )
}
