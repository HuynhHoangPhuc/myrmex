import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus, Calendar, Eye, MoreHorizontal, Wand2, Trash2 } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useSemesters, useDeleteSemester } from '@/modules/timetable/hooks/use-semesters'
import type { Semester } from '@/modules/timetable/types'
import { formatDate } from '@/lib/utils/format-date'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
})

export const Route = createFileRoute('/_authenticated/timetable/semesters/')({
  validateSearch: (search) => searchSchema.parse(search),
  component: SemesterListPage,
})

function SemesterListPage() {
  const { page, pageSize } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [deleteId, setDeleteId] = React.useState<string | null>(null)

  const { data, isLoading } = useSemesters({ page, pageSize })
  const deleteMutation = useDeleteSemester()

  const columns = React.useMemo<ColumnDef<Semester>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Semester',
        cell: ({ row }) => (
          <Link
            to="/timetable/semesters/$id"
            params={{ id: row.original.id }}
            className="font-medium text-primary hover:underline"
          >
            {row.getValue('name')}
          </Link>
        ),
      },
      { accessorKey: 'academic_year', header: 'Academic Year' },
      {
        accessorKey: 'start_date',
        header: 'Start',
        cell: ({ row }) => formatDate(row.getValue('start_date')),
      },
      {
        accessorKey: 'end_date',
        header: 'End',
        cell: ({ row }) => formatDate(row.getValue('end_date')),
      },
      {
        accessorKey: 'is_active',
        header: 'Status',
        cell: ({ row }) =>
          row.getValue('is_active') ? (
            <Badge variant="secondary">Active</Badge>
          ) : (
            <Badge variant="outline">Inactive</Badge>
          ),
      },
      {
        id: 'actions',
        cell: ({ row }) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem asChild>
                <Link to="/timetable/semesters/$id" params={{ id: row.original.id }}>
                  <Eye className="h-4 w-4" />
                  View details
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link
                  to="/timetable/semesters/new"
                  search={{ step: 3, semesterId: row.original.id }}
                >
                  <Wand2 className="h-4 w-4" />
                  Add offerings
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link to="/timetable/generate" search={{ semesterId: row.original.id }}>
                  <Calendar className="h-4 w-4" />
                  Generate schedule
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onSelect={() => setDeleteId(row.original.id)}
              >
                <Trash2 className="h-4 w-4" />
                Delete semester
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [],
  )

  return (
    <div>
      <PageHeader
        title="Semesters"
        description="Manage academic semesters, time slots, and rooms."
        actions={
          <Button asChild>
            <Link to="/timetable/semesters/new" search={{ step: 1 }}>
              <Plus className="mr-2 h-4 w-4" /> New Semester
            </Link>
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(nextPage) => void navigate({ search: { page: nextPage, pageSize } })}
      />

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Delete Semester"
        description="This will permanently remove the semester and all its schedules."
        confirmLabel="Delete"
        variant="destructive"
        isLoading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteId) return
          deleteMutation.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />
    </div>
  )
}
