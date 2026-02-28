import { MoreHorizontal, UserCog } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { ScheduleEntry } from '../types'

interface ScheduleEntryPopoverProps {
  entry: ScheduleEntry
  colorClassName: string
  onChangeTeacher?: (entry: ScheduleEntry) => void
}

export function ScheduleEntryPopover({
  entry,
  colorClassName,
  onChangeTeacher,
}: ScheduleEntryPopoverProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className={`w-full space-y-0.5 rounded border p-1.5 text-left text-xs transition-opacity hover:opacity-85 ${colorClassName} ${entry.is_manual_override ? 'ring-2 ring-yellow-400' : ''}`}
        >
          <div className="flex items-start gap-2">
            <div className="min-w-0 flex-1">
              <p className="truncate font-bold">{entry.subject_code}</p>
              <p className="truncate opacity-75">{entry.teacher_name}</p>
              <p className="truncate opacity-60">{entry.room_name}</p>
            </div>
            <MoreHorizontal className="mt-0.5 h-3.5 w-3.5 shrink-0 opacity-70" />
          </div>
          {entry.is_manual_override && (
            <Badge variant="outline" className="border-yellow-500 px-1 py-0 text-[10px] text-yellow-700">
              override
            </Badge>
          )}
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="start" className="w-72">
        <DropdownMenuLabel className="space-y-1">
          <p className="text-sm font-semibold">{entry.subject_code}</p>
          <p className="text-xs font-normal text-muted-foreground">{entry.subject_name}</p>
        </DropdownMenuLabel>
        <div className="space-y-1 px-2 py-1 text-xs text-muted-foreground">
          <p>Teacher: {entry.teacher_name}</p>
          <p>Room: {entry.room_name}</p>
          <p>
            Periods: P{entry.start_period}–P{entry.end_period}
          </p>
          <p>Department: {entry.department_id}</p>
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => onChangeTeacher?.(entry)}>
          <UserCog className="h-4 w-4" />
          Change teacher
        </DropdownMenuItem>
        <DropdownMenuItem disabled>
          {entry.is_manual_override ? 'Manual override active' : 'No remove action available'}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
