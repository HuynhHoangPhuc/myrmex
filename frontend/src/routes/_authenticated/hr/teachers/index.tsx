import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useTeachers } from '@/modules/hr/hooks/use-teachers'
import { useDeleteTeacher } from '@/modules/hr/hooks/use-teacher-mutations'
import { buildTeacherColumns } from '@/modules/hr/components/teacher-columns'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  search: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/hr/teachers/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: TeacherListPage,
})

function TeacherListPage() {
  const { page, pageSize, search } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [deleteId, setDeleteId] = React.useState<string | null>(null)
  const [searchInput, setSearchInput] = React.useState(search ?? '')

  const { data, isLoading } = useTeachers({ page, pageSize, search })
  const deleteMutation = useDeleteTeacher()

  const columns = React.useMemo(() => buildTeacherColumns(setDeleteId), [])

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    void navigate({ search: { page: 1, pageSize, search: searchInput || undefined } })
  }

  return (
    <div>
      <PageHeader
        title="Teachers"
        description="Manage faculty teachers and their availability."
        actions={
          <Button asChild>
            <Link to="/hr/teachers/new">
              <Plus className="mr-2 h-4 w-4" />
              Add Teacher
            </Link>
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: { page: p, pageSize, search } })}
        toolbar={
          <form onSubmit={handleSearch} className="flex gap-2">
            <Input
              placeholder="Search by name or code…"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="w-64"
            />
            <Button type="submit" variant="outline" size="sm">Search</Button>
          </form>
        }
      />

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(o) => !o && setDeleteId(null)}
        title="Delete Teacher"
        description="This action cannot be undone. The teacher will be permanently removed."
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
