import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useRooms } from '../hooks/use-rooms'
import { useAssignRoom } from '../hooks/use-schedules'
import { periodToTimeLabel } from '../utils/period-to-time'
import type { Room, ScheduleEntry } from '../types'

interface RoomAssignmentDialogProps {
  scheduleId: string
  entry: ScheduleEntry | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

const ROOM_TYPE_LABELS: Record<Room['room_type'], string> = {
  lecture: 'Lecture',
  lab: 'Lab',
  seminar: 'Seminar',
}

// Dialog: shows all available rooms + confirm-on-assign
export function RoomAssignmentDialog({
  scheduleId,
  entry,
  open,
  onOpenChange,
}: RoomAssignmentDialogProps) {
  const [pendingRoom, setPendingRoom] = React.useState<Room | null>(null)
  const { data: rooms, isLoading } = useRooms()
  const assignMutation = useAssignRoom(scheduleId)

  function handleConfirm() {
    if (!entry || !pendingRoom) return
    assignMutation.mutate(
      { entry_id: entry.id, room_id: pendingRoom.id },
      {
        onSuccess: () => {
          setPendingRoom(null)
          onOpenChange(false)
        },
      },
    )
  }

  if (!entry) return null

  return (
    <>
      <Dialog open={open && !pendingRoom} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[480px]">
          <DialogHeader>
            <DialogTitle>Assign Room</DialogTitle>
            <DialogDescription>
              {entry.subject_code} — {periodToTimeLabel(entry.start_period, entry.end_period)} · currently {entry.room_name}
            </DialogDescription>
          </DialogHeader>

          {isLoading ? (
            <div className="flex min-h-32 items-center justify-center">
              <LoadingSpinner />
            </div>
          ) : !rooms?.length ? (
            <p className="py-4 text-center text-sm text-muted-foreground">No rooms available.</p>
          ) : (
            <div className="max-h-96 divide-y overflow-y-auto rounded-md border">
              {rooms.map((room) => {
                const isCurrent = room.id === entry.room_id
                return (
                  <div
                    key={room.id}
                    className="flex items-center gap-3 px-3 py-2.5"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{room.name}</p>
                      <p className="text-xs text-muted-foreground">{room.capacity} seats</p>
                    </div>
                    <Badge variant="secondary" className="capitalize shrink-0">
                      {ROOM_TYPE_LABELS[room.room_type] ?? room.room_type}
                    </Badge>
                    <Button
                      size="sm"
                      variant={isCurrent ? 'secondary' : 'outline'}
                      disabled={isCurrent}
                      onClick={() => setPendingRoom(room)}
                    >
                      {isCurrent ? 'Current' : 'Assign'}
                    </Button>
                  </div>
                )
              })}
            </div>
          )}
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={Boolean(pendingRoom)}
        onOpenChange={(o) => !o && setPendingRoom(null)}
        title="Confirm Room Assignment"
        description={`Assign ${pendingRoom?.name} to ${entry.subject_code}? This will mark the entry as a manual override.`}
        confirmLabel="Assign"
        isLoading={assignMutation.isPending}
        onConfirm={handleConfirm}
      />
    </>
  )
}
