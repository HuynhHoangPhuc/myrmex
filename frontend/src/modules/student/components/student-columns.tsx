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
import type { Student } from '../types'

const STATUS_VARIANT = {
  active: 'default',
  graduated: 'secondary',
  suspended: 'destructive',
} as const

interface ActionsProps {
  student: Student
  onDelete: (id: string) => void
}

function StudentActions({ student, onDelete }: ActionsProps) {
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
          <Link to="/students/$studentId" params={{ studentId: student.id }}>
            <Pencil className="mr-2 h-4 w-4" />
            View / Edit
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onClick={() => onDelete(student.id)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function buildStudentColumns(onDelete: (id: string) => void): ColumnDef<Student>[] {
  return [
    {
      accessorKey: 'student_code',
      header: 'Code',
      cell: ({ row }) => (
        <Link
          to="/students/$studentId"
          params={{ studentId: row.original.id }}
          className="font-mono text-sm font-medium text-primary hover:underline"
        >
          {row.getValue('student_code')}
        </Link>
      ),
    },
    { accessorKey: 'full_name', header: 'Full Name' },
    { accessorKey: 'email', header: 'Email' },
    {
      accessorKey: 'enrollment_year',
      header: 'Year',
      cell: ({ row }) => row.getValue('enrollment_year'),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={STATUS_VARIANT[s] ?? 'outline'}>{s}</Badge>
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => <StudentActions student={row.original} onDelete={onDelete} />,
    },
  ]
}
