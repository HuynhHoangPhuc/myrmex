import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus, Trash2, Calendar } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useSemesters, useDeleteSemester } from '@/modules/timetable/hooks/use-semesters'
import type { Semester } from '@/modules/timetable/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
})

export const Route = createFileRoute('/_authenticated/timetable/semesters/')({
  validateSearch: (s) => searchSchema.parse(s),
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
      { accessorKey: 'start_date', header: 'Start' },
      { accessorKey: 'end_date', header: 'End' },
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
          <div className="flex gap-1">
            <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
              <Link to="/timetable/generate" search={{ semesterId: row.original.id }}>
                <Calendar className="h-4 w-4" />
              </Link>
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setDeleteId(row.original.id)}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
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
            <Link to="/timetable/semesters/new">
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
        onPageChange={(p) => void navigate({ search: { page: p, pageSize } })}
      />

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(o) => !o && setDeleteId(null)}
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
