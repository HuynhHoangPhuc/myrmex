import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { Pencil, ArrowLeft, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { PrerequisiteGraph } from '@/modules/subject/components/prerequisite-graph'
import { useSubject, useAllSubjects, useAddPrerequisite, useRemovePrerequisite } from '@/modules/subject/hooks/use-subjects'
import type { PrerequisiteType } from '@/modules/subject/types'

export const Route = createFileRoute('/_authenticated/subjects/$id/')({
  component: SubjectDetailPage,
})

function SubjectDetailPage() {
  const { id } = Route.useParams()
  const { data: subject, isLoading } = useSubject(id)
  const { data: allSubjects = [] } = useAllSubjects()
  const addPrereq = useAddPrerequisite(id)
  const removePrereq = useRemovePrerequisite(id)

  const [removeId, setRemoveId] = React.useState<string | null>(null)
  const [addForm, setAddForm] = React.useState({ prerequisite_id: '', prerequisite_type: 'hard' as PrerequisiteType })
  const [showAddForm, setShowAddForm] = React.useState(false)

  if (isLoading) return <LoadingSpinner />
  if (!subject) return <p className="text-muted-foreground">Subject not found.</p>

  const existingPrereqIds = new Set(subject.prerequisites?.map((p) => p.prerequisite_id) ?? [])
  const availableSubjects = allSubjects.filter((s) => s.id !== id && !existingPrereqIds.has(s.id))

  function handleAddPrereq(e: React.FormEvent) {
    e.preventDefault()
    if (!addForm.prerequisite_id) return
    addPrereq.mutate(addForm, { onSuccess: () => { setShowAddForm(false); setAddForm({ prerequisite_id: '', prerequisite_type: 'hard' }) } })
  }

  return (
    <div className="max-w-4xl space-y-8">
      <PageHeader
        title={subject.name}
        description={`${subject.code} · ${subject.credits} credits · ${subject.weekly_hours}h/week`}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" asChild>
              <Link to="/subjects" search={{ page: 1, pageSize: 25 }}><ArrowLeft className="mr-2 h-4 w-4" />Back</Link>
            </Button>
            <Button asChild>
              <Link to="/subjects/$id/edit" params={{ id }}>
                <Pencil className="mr-2 h-4 w-4" />Edit
              </Link>
            </Button>
          </div>
        }
      />

      {subject.description && (
        <p className="text-sm text-muted-foreground rounded-lg border p-4">{subject.description}</p>
      )}

      {/* Prerequisites list */}
      <div>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Prerequisites</h2>
          <Button size="sm" variant="outline" onClick={() => setShowAddForm((v) => !v)}>
            <Plus className="mr-1 h-3.5 w-3.5" />Add
          </Button>
        </div>

        {showAddForm && (
          <form onSubmit={handleAddPrereq} className="mb-4 flex items-center gap-2 rounded-lg border p-3">
            <select
              value={addForm.prerequisite_id}
              onChange={(e) => setAddForm((f) => ({ ...f, prerequisite_id: e.target.value }))}
              className="h-9 flex-1 rounded-md border border-input bg-transparent px-3 text-sm"
            >
              <option value="">Select subject…</option>
              {availableSubjects.map((s) => (
                <option key={s.id} value={s.id}>{s.code} — {s.name}</option>
              ))}
            </select>
            <select
              value={addForm.prerequisite_type}
              onChange={(e) => setAddForm((f) => ({ ...f, prerequisite_type: e.target.value as PrerequisiteType }))}
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            >
              <option value="hard">Hard</option>
              <option value="soft">Soft</option>
            </select>
            <Button type="submit" size="sm" disabled={addPrereq.isPending || !addForm.prerequisite_id}>
              Add
            </Button>
          </form>
        )}

        {subject.prerequisites?.length === 0 && (
          <p className="text-sm text-muted-foreground">No prerequisites defined.</p>
        )}
        <div className="space-y-1">
          {subject.prerequisites?.map((p) => {
            const prereqSubject = allSubjects.find((s) => s.id === p.prerequisite_id)
            return (
              <div key={p.prerequisite_id} className="flex items-center justify-between rounded-md border px-3 py-2">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm font-medium text-primary">
                    {prereqSubject?.code ?? p.prerequisite_id}
                  </span>
                  <span className="text-sm text-muted-foreground">{prereqSubject?.name}</span>
                  <Badge variant={p.prerequisite_type === 'hard' ? 'destructive' : 'outline'} className="text-xs">
                    {p.prerequisite_type}
                  </Badge>
                </div>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setRemoveId(p.prerequisite_id)}>
                  <Trash2 className="h-3.5 w-3.5 text-destructive" />
                </Button>
              </div>
            )
          })}
        </div>
      </div>

      {/* DAG visualization */}
      <div>
        <h2 className="mb-3 text-lg font-semibold">Prerequisite Graph</h2>
        <div className="rounded-lg border overflow-hidden">
          <PrerequisiteGraph focusSubjectId={id} />
        </div>
      </div>

      <ConfirmDialog
        open={Boolean(removeId)}
        onOpenChange={(o) => !o && setRemoveId(null)}
        title="Remove Prerequisite"
        description="Remove this prerequisite relationship?"
        confirmLabel="Remove"
        variant="destructive"
        isLoading={removePrereq.isPending}
        onConfirm={() => {
          if (!removeId) return
          removePrereq.mutate(removeId, { onSuccess: () => setRemoveId(null) })
        }}
      />
    </div>
  )
}
