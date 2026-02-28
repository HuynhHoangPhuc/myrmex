import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, UserCog } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/shared/empty-state'
import { PageHeader } from '@/components/shared/page-header'
import { PageSkeleton } from '@/components/shared/page-skeleton'
import { ScheduleCalendar } from '@/modules/timetable/components/schedule-calendar'
import {
  ScheduleFilterBar,
  type ScheduleFilters,
} from '@/modules/timetable/components/schedule-filter-bar'
import { TeacherAssignmentDialog } from '@/modules/timetable/components/teacher-assignment'
import { useSchedule } from '@/modules/timetable/hooks/use-schedules'
import type { ScheduleEntry } from '@/modules/timetable/types'

const searchSchema = z.object({
  departmentId: z.string().optional().catch(undefined),
  teacherName: z.string().optional().catch(undefined),
  roomId: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/timetable/schedules/$id/')({
  validateSearch: (search) => searchSchema.parse(search),
  component: ScheduleDetailPage,
})

function ScheduleDetailPage() {
  const { id } = Route.useParams()
  const filters = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data: schedule, isLoading } = useSchedule(id)
  const [selectedEntry, setSelectedEntry] = React.useState<ScheduleEntry | null>(null)
  const [assignOpen, setAssignOpen] = React.useState(false)

  function handleEntryAction(entry: ScheduleEntry) {
    setSelectedEntry(entry)
    setAssignOpen(true)
  }

  function handleFilterChange(nextFilters: ScheduleFilters) {
    void navigate({
      search: {
        departmentId: nextFilters.departmentId || undefined,
        teacherName: nextFilters.teacherName || undefined,
        roomId: nextFilters.roomId || undefined,
      },
      replace: true,
    })
  }

  if (isLoading) return <PageSkeleton variant="detail" />

  if (!schedule) {
    return (
      <EmptyState
        title="Schedule not found"
        description="This schedule may have been removed or is no longer available."
      />
    )
  }

  const overrideCount = schedule.entries.filter((entry) => entry.is_manual_override).length

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
              <Badge variant="outline" className="border-yellow-500 text-yellow-700 dark:text-yellow-300">
                {overrideCount} manual overrides
              </Badge>
            )}
          </>
        )}
      </div>

      <ScheduleFilterBar
        entries={schedule.entries}
        filters={filters}
        onFilterChange={handleFilterChange}
      />

      <ScheduleCalendar
        schedule={schedule}
        filters={filters}
        onChangeTeacher={handleEntryAction}
      />

      <TeacherAssignmentDialog
        scheduleId={id}
        entry={selectedEntry}
        open={assignOpen}
        onOpenChange={(open) => {
          setAssignOpen(open)
          if (!open) setSelectedEntry(null)
        }}
      />
    </div>
  )
}
