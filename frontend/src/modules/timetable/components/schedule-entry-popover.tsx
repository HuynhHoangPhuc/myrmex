import { DoorOpen, MoreHorizontal, UserCog } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAllDepartments } from '@/modules/hr/hooks/use-departments'
import type { ScheduleEntry } from '../types'

interface ScheduleEntryPopoverProps {
  entry: ScheduleEntry
  colorClassName: string
  onChangeTeacher?: (entry: ScheduleEntry) => void
  onChangeRoom?: (entry: ScheduleEntry) => void
}

export function ScheduleEntryPopover({
  entry,
  colorClassName,
  onChangeTeacher,
  onChangeRoom,
}: ScheduleEntryPopoverProps) {
  const { data: departments } = useAllDepartments()
  const departmentName = departments?.find((d) => d.id === entry.department_id)?.name ?? entry.department_id

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className={`w-full space-y-0.5 rounded border p-1.5 text-left text-xs transition-all hover:opacity-85 data-[state=open]:shadow-md data-[state=open]:ring-2 data-[state=open]:ring-primary/60 data-[state=open]:opacity-100 ${colorClassName} ${entry.is_manual_override ? 'ring-2 ring-yellow-400' : ''}`}
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
          <Link
            to="/subjects/$id"
            params={{ id: entry.subject_id }}
            className="text-sm font-semibold hover:underline"
          >
            {entry.subject_code}
          </Link>
          <p className="text-xs font-normal text-muted-foreground">{entry.subject_name}</p>
        </DropdownMenuLabel>
        <div className="space-y-1 px-2 py-1 text-xs text-muted-foreground">
          <p>
            Teacher:{' '}
            <Link
              to="/hr/teachers/$id"
              params={{ id: entry.teacher_id }}
              className="font-medium text-foreground hover:underline"
            >
              {entry.teacher_name}
            </Link>
          </p>
          <p>Room: {entry.room_name}</p>
          <p>Periods: P{entry.start_period}–P{entry.end_period}</p>
          <p>Department: {departmentName}</p>
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => onChangeTeacher?.(entry)}>
          <UserCog className="h-4 w-4" />
          Change teacher
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => onChangeRoom?.(entry)}>
          <DoorOpen className="h-4 w-4" />
          Change room
        </DropdownMenuItem>
        {entry.is_manual_override && (
          <DropdownMenuItem disabled>Manual override active</DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
