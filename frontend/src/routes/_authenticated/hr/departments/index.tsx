import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus, Trash2 } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useDepartments, useDeleteDepartment } from '@/modules/hr/hooks/use-departments'
import type { Department } from '@/modules/hr/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
})

export const Route = createFileRoute('/_authenticated/hr/departments/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: DepartmentListPage,
})

function DepartmentListPage() {
  const { page, pageSize } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [deleteId, setDeleteId] = React.useState<string | null>(null)

  const { data, isLoading } = useDepartments({ page, pageSize })
  const deleteMutation = useDeleteDepartment()

  const columns = React.useMemo<ColumnDef<Department>[]>(
    () => [
      {
        accessorKey: 'code',
        header: 'Code',
        cell: ({ row }) => (
          <span className="font-mono text-sm font-bold text-primary">{row.getValue('code')}</span>
        ),
      },
      { accessorKey: 'name', header: 'Name' },
      {
        accessorKey: 'description',
        header: 'Description',
        cell: ({ row }) => row.getValue('description') ?? <span className="text-muted-foreground">—</span>,
      },
      {
        id: 'actions',
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => setDeleteId(row.original.id)}
          >
            <Trash2 className="h-4 w-4 text-destructive" />
          </Button>
        ),
      },
    ],
    [],
  )

  return (
    <div>
      <PageHeader
        title="Departments"
        description="Manage faculty departments."
        actions={
          <Button asChild>
            <Link to="/hr/departments/new">
              <Plus className="mr-2 h-4 w-4" /> Add Department
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
        title="Delete Department"
        description="This will permanently remove the department."
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
