import type { ColumnDef } from '@tanstack/react-table'
import { Link } from '@tanstack/react-router'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import type { Teacher } from '../types'

interface ActionsProps {
  teacher: Teacher
  onDelete: (id: string) => void
}

function TeacherActions({ teacher, onDelete }: ActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8">
          <MoreHorizontal className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem asChild>
          <Link to="/hr/teachers/$id/edit" params={{ id: teacher.id }}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onClick={() => onDelete(teacher.id)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// Column definitions for teacher DataTable — passed onDelete via closure
export function buildTeacherColumns(onDelete: (id: string) => void): ColumnDef<Teacher>[] {
  return [
    {
      accessorKey: 'employee_code',
      header: 'Code',
      cell: ({ row }) => (
        <Link
          to="/hr/teachers/$id"
          params={{ id: row.original.id }}
          className="font-mono text-sm font-medium text-primary hover:underline"
        >
          {row.getValue('employee_code')}
        </Link>
      ),
    },
    {
      accessorKey: 'full_name',
      header: 'Full Name',
    },
    {
      accessorKey: 'email',
      header: 'Email',
    },
    {
      accessorKey: 'department',
      header: 'Department',
      cell: ({ row }) => row.original.department?.name ?? '—',
    },
    {
      accessorKey: 'max_hours_per_week',
      header: 'Max hrs/wk',
      cell: ({ row }) => `${row.getValue('max_hours_per_week')}h`,
    },
    {
      accessorKey: 'specializations',
      header: 'Specializations',
      cell: ({ row }) => {
        const specs = row.original.specializations
        if (!specs?.length) return <span className="text-muted-foreground">—</span>
        return (
          <div className="flex flex-wrap gap-1">
            {specs.slice(0, 2).map((s) => (
              <Badge key={s} variant="secondary" className="text-xs">
                {s}
              </Badge>
            ))}
            {specs.length > 2 && (
              <Badge variant="outline" className="text-xs">
                +{specs.length - 2}
              </Badge>
            )}
          </div>
        )
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => <TeacherActions teacher={row.original} onDelete={onDelete} />,
    },
  ]
}
