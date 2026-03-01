import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useToast } from '@/lib/hooks/use-toast'
import { useRooms } from '../hooks/use-rooms'
import { useSetSemesterRooms } from '../hooks/use-semesters'
import type { Room } from '../types'

interface RoomManagerProps {
  semesterId: string
  /** Currently persisted room IDs for this semester (empty = all rooms used by default) */
  selectedRoomIds: string[]
}

const ROOM_TYPE_LABELS: Record<Room['room_type'], string> = {
  lecture: 'Lecture',
  lab: 'Lab',
  seminar: 'Seminar',
}

export function RoomManager({ semesterId, selectedRoomIds }: RoomManagerProps) {
  const { data: allRooms, isLoading } = useRooms()
  const setRoomsMutation = useSetSemesterRooms(semesterId)
  const { toast } = useToast()

  const [selected, setSelected] = React.useState<Set<string>>(() => new Set(selectedRoomIds))

  // Pre-select all rooms when no selection is persisted yet and rooms have loaded
  const initialised = React.useRef(false)
  React.useEffect(() => {
    if (!initialised.current && allRooms) {
      initialised.current = true
      if (selectedRoomIds.length === 0) {
        setSelected(new Set(allRooms.map((r) => r.id)))
      }
    }
  }, [allRooms, selectedRoomIds])

  function toggle(roomId: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(roomId)) next.delete(roomId)
      else next.add(roomId)
      return next
    })
  }

  function toggleAll() {
    if (!allRooms) return
    setSelected(
      selected.size === allRooms.length ? new Set() : new Set(allRooms.map((r) => r.id)),
    )
  }

  async function save() {
    setRoomsMutation.mutate([...selected], {
      onSuccess: () => toast({ title: 'Room selection saved' }),
      onError: () => toast({ title: 'Failed to save rooms', variant: 'destructive' }),
    })
  }

  if (isLoading) {
    return (
      <div className="flex min-h-24 items-center justify-center">
        <LoadingSpinner />
      </div>
    )
  }

  if (!allRooms?.length) {
    return (
      <p className="text-sm text-muted-foreground">
        No rooms found. Add rooms via the system settings first.
      </p>
    )
  }

  const allChecked = selected.size === allRooms.length

  return (
    <div className="space-y-3">
      {/* Select-all header */}
      <label className="flex cursor-pointer items-center gap-3 rounded-md border bg-muted/40 px-4 py-2.5">
        <input
          type="checkbox"
          checked={allChecked}
          ref={(el) => {
            if (el) el.indeterminate = selected.size > 0 && !allChecked
          }}
          onChange={toggleAll}
          className="h-4 w-4 rounded border-input accent-primary"
        />
        <span className="flex-1 text-sm font-medium">
          {allChecked ? 'Deselect all' : 'Select all'} ({selected.size}/{allRooms.length} selected)
        </span>
      </label>

      {/* Room list */}
      <div className="divide-y rounded-md border">
        {allRooms.map((room) => (
          <label
            key={room.id}
            className="flex cursor-pointer items-center gap-3 px-4 py-2.5 hover:bg-muted/30 transition-colors"
          >
            <input
              type="checkbox"
              checked={selected.has(room.id)}
              onChange={() => toggle(room.id)}
              className="h-4 w-4 rounded border-input accent-primary"
            />
            <span className="w-32 truncate text-sm font-medium">{room.name}</span>
            <span className="text-sm text-muted-foreground">{room.capacity} seats</span>
            <Badge variant="secondary" className="ml-auto capitalize">
              {ROOM_TYPE_LABELS[room.room_type] ?? room.room_type}
            </Badge>
          </label>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <Button
          type="button"
          size="sm"
          onClick={save}
          disabled={setRoomsMutation.isPending || selected.size === 0}
        >
          {setRoomsMutation.isPending ? 'Saving…' : 'Save room selection'}
        </Button>
        {selected.size === 0 && (
          <p className="text-xs text-destructive">Select at least one room to continue.</p>
        )}
      </div>
    </div>
  )
}
