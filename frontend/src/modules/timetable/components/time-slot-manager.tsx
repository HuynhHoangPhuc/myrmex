// Manages time slots for a semester: add, delete, apply presets.
import { Badge } from '@/components/ui/badge'
import { AddSlotDialog, PresetDialog, DeleteConfirmDialog } from './time-slot-dialogs'
import { DAY_NAMES } from './time-slot-constants'
import type { TimeSlot } from '../types'

interface TimeSlotManagerProps {
  semesterId: string
  slots: TimeSlot[]
}

export function TimeSlotManager({ semesterId, slots }: TimeSlotManagerProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold">Time slots ({slots.length})</h4>
        <div className="flex gap-2">
          <PresetDialog semesterId={semesterId} hasSlots={slots.length > 0} />
          <AddSlotDialog semesterId={semesterId} />
        </div>
      </div>

      {!slots.length ? (
        <p className="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
          No time slots defined. Add one manually or use a preset.
        </p>
      ) : (
        <div className="rounded-lg border divide-y">
          {slots.map((slot, i) => (
            <div key={slot.id} className="flex items-center gap-3 px-4 py-2.5 text-sm">
              <span className="w-12 font-medium">{DAY_NAMES[slot.day_of_week]}</span>
              <span className="text-muted-foreground">
                {slot.start_time} – {slot.end_time}
              </span>
              <Badge variant="outline" className="ml-auto text-xs">Slot {i + 1}</Badge>
              <DeleteConfirmDialog slot={slot} semesterId={semesterId} />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
