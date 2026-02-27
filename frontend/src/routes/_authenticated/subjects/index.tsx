import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useSubjects, useDeleteSubject, usePrereqMap } from '@/modules/subject/hooks/use-subjects'
import { buildSubjectColumns } from '@/modules/subject/components/subject-columns'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  search: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/subjects/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: SubjectListPage,
})

function SubjectListPage() {
  const { page, pageSize, search } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [deleteId, setDeleteId] = React.useState<string | null>(null)
  const [searchInput, setSearchInput] = React.useState(search ?? '')

  const { data, isLoading } = useSubjects({ page, pageSize, search })
  const deleteMutation = useDeleteSubject()
  const prereqMap = usePrereqMap()
  const columns = React.useMemo(() => buildSubjectColumns(setDeleteId, prereqMap), [prereqMap])

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    void navigate({ search: { page: 1, pageSize, search: searchInput || undefined } })
  }

  return (
    <div>
      <PageHeader
        title="Subjects"
        description="Manage course subjects and prerequisite relationships."
        actions={
          <Button asChild>
            <Link to="/subjects/new">
              <Plus className="mr-2 h-4 w-4" /> Add Subject
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
              placeholder="Search by code or name…"
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
        title="Delete Subject"
        description="This will permanently remove the subject and its prerequisite relationships."
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
