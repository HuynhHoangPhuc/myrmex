import type { ColumnDef } from '@tanstack/react-table'
import { Link } from '@tanstack/react-router'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Subject } from '../types'

interface ActionsProps {
  subject: Subject
  onDelete: (id: string) => void
}

function SubjectActions({ subject, onDelete }: ActionsProps) {
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
          <Link to="/subjects/$id/edit" params={{ id: subject.id }}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onClick={() => onDelete(subject.id)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function buildSubjectColumns(onDelete: (id: string) => void): ColumnDef<Subject>[] {
  return [
    {
      accessorKey: 'code',
      header: 'Code',
      cell: ({ row }) => (
        <Link
          to="/subjects/$id"
          params={{ id: row.original.id }}
          className="font-mono text-sm font-medium text-primary hover:underline"
        >
          {row.getValue('code')}
        </Link>
      ),
    },
    {
      accessorKey: 'name',
      header: 'Name',
    },
    {
      accessorKey: 'credits',
      header: 'Credits',
      cell: ({ row }) => (
        <Badge variant="secondary">{row.getValue<number>('credits')} cr</Badge>
      ),
    },
    {
      accessorKey: 'weekly_hours',
      header: 'Hrs/Week',
      cell: ({ row }) => `${row.getValue('weekly_hours')}h`,
    },
    {
      accessorKey: 'prerequisites',
      header: 'Prerequisites',
      cell: ({ row }) => {
        const count = row.original.prerequisites?.length ?? 0
        return count === 0
          ? <span className="text-muted-foreground">None</span>
          : <Badge variant="outline">{count}</Badge>
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => <SubjectActions subject={row.original} onDelete={onDelete} />,
    },
  ]
}
