import type { ScheduleEntry } from '../types'

interface ScheduleDraggableEntryProps {
  entry: ScheduleEntry
  disabled?: boolean
  isDragging?: boolean
  onDragStart: (entry: ScheduleEntry) => void
  onDragEnd: () => void
  children: React.ReactNode
}

export function ScheduleDraggableEntry({
  entry,
  disabled = false,
  isDragging = false,
  onDragStart,
  onDragEnd,
  children,
}: ScheduleDraggableEntryProps) {
  return (
    <div
      draggable={!disabled}
      onDragStart={(event) => {
        if (disabled) return
        event.dataTransfer.effectAllowed = 'move'
        onDragStart(entry)
      }}
      onDragEnd={onDragEnd}
      className={isDragging ? 'opacity-50' : undefined}
    >
      {children}
    </div>
  )
}
