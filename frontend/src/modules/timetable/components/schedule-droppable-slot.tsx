import { cn } from '@/lib/utils/cn'

interface ScheduleDroppableSlotProps {
  isOver: boolean
  onDragOver: React.DragEventHandler<HTMLDivElement>
  onDragLeave: React.DragEventHandler<HTMLDivElement>
  onDrop: React.DragEventHandler<HTMLDivElement>
  children: React.ReactNode
}

export function ScheduleDroppableSlot({
  isOver,
  onDragOver,
  onDragLeave,
  onDrop,
  children,
}: ScheduleDroppableSlotProps) {
  return (
    <div
      className={cn(
        'min-h-16 rounded-md p-1.5 transition-colors',
        isOver && 'bg-primary/5 ring-2 ring-primary/40',
      )}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      {children}
    </div>
  )
}
