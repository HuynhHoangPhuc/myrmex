import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { TeacherAssignmentDialog } from '@/modules/timetable/components/teacher-assignment'
import { useSchedule, useSchedules } from '@/modules/timetable/hooks/use-schedules'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import type { ScheduleEntry } from '@/modules/timetable/types'

const searchSchema = z.object({
  scheduleId: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/timetable/assign/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: AssignPage,
})

function AssignPage() {
  const { scheduleId: initialScheduleId } = Route.useSearch()
  const navigate = Route.useNavigate()

  const { data: semesters = [] } = useAllSemesters()
  const [semesterId, setSemesterId] = React.useState('')
  const { data: schedulesData } = useSchedules({ page: 1, pageSize: 50, semesterId })
  const [scheduleId, setScheduleId] = React.useState(initialScheduleId ?? '')

  const { data: schedule, isLoading } = useSchedule(scheduleId)
  const [selectedEntry, setSelectedEntry] = React.useState<ScheduleEntry | null>(null)
  const [assignOpen, setAssignOpen] = React.useState(false)

  function handleEntryClick(entry: ScheduleEntry) {
    setSelectedEntry(entry)
    setAssignOpen(true)
  }

  function handleScheduleChange(id: string) {
    setScheduleId(id)
    void navigate({ search: { scheduleId: id || undefined } })
  }

  const overrideEntries = schedule?.entries.filter((e) => e.is_manual_override) ?? []

  return (
    <div className="max-w-4xl space-y-6">
      <PageHeader
        title="Manual Teacher Assignment"
        description="Override auto-assigned teachers for specific schedule entries."
        actions={
          scheduleId ? (
            <Button variant="outline" asChild>
              <Link to="/timetable/schedules/$id" params={{ id: scheduleId }}>
                <ArrowLeft className="mr-2 h-4 w-4" />View Calendar
              </Link>
            </Button>
          ) : undefined
        }
      />

      {/* Selectors */}
      <div className="flex flex-wrap gap-3">
        <select
          value={semesterId}
          onChange={(e) => { setSemesterId(e.target.value); setScheduleId('') }}
          className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
        >
          <option value="">All semesters</option>
          {semesters.map((s) => (
            <option key={s.id} value={s.id}>{s.name}</option>
          ))}
        </select>

        <select
          value={scheduleId}
          onChange={(e) => handleScheduleChange(e.target.value)}
          className="h-9 flex-1 rounded-md border border-input bg-transparent px-3 text-sm"
        >
          <option value="">Select schedule…</option>
          {schedulesData?.data.filter((s) => s.status === 'completed').map((s) => (
            <option key={s.id} value={s.id}>
              {s.id.slice(0, 8)} — Score {s.score.toFixed(2)}
            </option>
          ))}
        </select>
      </div>

      {!scheduleId && (
        <p className="text-sm text-muted-foreground">Select a completed schedule to manage assignments.</p>
      )}

      {scheduleId && isLoading && <LoadingSpinner />}

      {schedule && (
        <>
          {/* Override summary */}
          {overrideEntries.length > 0 && (
            <div className="rounded-lg border p-4 space-y-2">
              <h3 className="text-sm font-semibold">
                Manual Overrides ({overrideEntries.length})
              </h3>
              <div className="space-y-1">
                {overrideEntries.map((e) => (
                  <div key={e.id} className="flex items-center gap-3 text-sm">
                    <span className="font-mono text-primary w-20">{e.subject_code}</span>
                    <span className="text-muted-foreground">{e.teacher_name}</span>
                    <span className="text-xs text-muted-foreground">{e.start_time}–{e.end_time}</span>
                    <Badge variant="outline" className="ml-auto border-yellow-500 text-yellow-700 text-xs">override</Badge>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => handleEntryClick(e)}
                    >
                      Reassign
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* All entries list */}
          <div>
            <h3 className="mb-2 text-sm font-semibold">All Entries ({schedule.entries.length})</h3>
            <div className="divide-y rounded-md border max-h-[500px] overflow-y-auto">
              {schedule.entries.map((entry) => (
                <div key={entry.id} className="flex items-center gap-3 px-4 py-2.5 text-sm hover:bg-muted/30">
                  <span className="font-mono text-primary w-20 shrink-0">{entry.subject_code}</span>
                  <span className="flex-1 truncate">{entry.teacher_name}</span>
                  <span className="text-xs text-muted-foreground w-28 shrink-0">
                    {['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][entry.day_of_week]} {entry.start_time}
                  </span>
                  <span className="text-xs text-muted-foreground w-20 shrink-0">{entry.room_name}</span>
                  {entry.is_manual_override && (
                    <Badge variant="outline" className="border-yellow-500 text-yellow-700 text-xs shrink-0">override</Badge>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 text-xs shrink-0"
                    onClick={() => handleEntryClick(entry)}
                  >
                    Assign
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </>
      )}

      {scheduleId && schedule && (
        <TeacherAssignmentDialog
          scheduleId={scheduleId}
          entry={selectedEntry}
          open={assignOpen}
          onOpenChange={(o) => { setAssignOpen(o); if (!o) setSelectedEntry(null) }}
        />
      )}
    </div>
  )
}
