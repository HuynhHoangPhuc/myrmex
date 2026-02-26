import * as React from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { TeacherSuggestionList } from './teacher-suggestion-list'
import { useAssignTeacher } from '../hooks/use-schedules'
import { periodToTimeLabel } from '../utils/period-to-time'
import type { ScheduleEntry, TeacherSuggestion } from '../types'

interface TeacherAssignmentDialogProps {
  scheduleId: string
  entry: ScheduleEntry | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Dialog: shows entry info + ranked teacher suggestions + confirm-on-assign
export function TeacherAssignmentDialog({
  scheduleId,
  entry,
  open,
  onOpenChange,
}: TeacherAssignmentDialogProps) {
  const [pending, setPending] = React.useState<TeacherSuggestion | null>(null)
  const assignMutation = useAssignTeacher(scheduleId)

  function handleSelect(suggestion: TeacherSuggestion) {
    setPending(suggestion)
  }

  function handleConfirm() {
    if (!entry || !pending) return
    assignMutation.mutate(
      { entry_id: entry.id, teacher_id: pending.teacher_id },
      {
        onSuccess: () => {
          setPending(null)
          onOpenChange(false)
        },
      },
    )
  }

  if (!entry) return null

  return (
    <>
      <Dialog open={open && !pending} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[520px]">
          <DialogHeader>
            <DialogTitle>Assign Teacher</DialogTitle>
            <DialogDescription>
              {entry.subject_code} — {periodToTimeLabel(entry.start_period, entry.end_period)} · {entry.room_name}
            </DialogDescription>
          </DialogHeader>

          <TeacherSuggestionList
            scheduleId={scheduleId}
            entry={entry}
            currentTeacherId={entry.teacher_id}
            onSelect={handleSelect}
          />
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={Boolean(pending)}
        onOpenChange={(o) => !o && setPending(null)}
        title="Confirm Assignment"
        description={`Assign ${pending?.teacher_name} to ${entry.subject_code}? This will mark the entry as a manual override.`}
        confirmLabel="Assign"
        isLoading={assignMutation.isPending}
        onConfirm={handleConfirm}
      />
    </>
  )
}
