import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { PageHeader } from '@/components/shared/page-header'
import { GenerationPanel } from '@/modules/timetable/components/generation-panel'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import * as React from 'react'

const searchSchema = z.object({
  semesterId: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/timetable/generate/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: GeneratePage,
})

function GeneratePage() {
  const { semesterId: initialSemesterId } = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data: semesters = [] } = useAllSemesters()
  const [semesterId, setSemesterId] = React.useState(initialSemesterId ?? '')

  // Sync URL search param when semester changes
  function handleSemesterChange(id: string) {
    setSemesterId(id)
    void navigate({ search: { semesterId: id || undefined } })
  }

  return (
    <div className="max-w-2xl space-y-6">
      <PageHeader
        title="Generate Schedule"
        description="Run the CSP solver to generate an optimised timetable for a semester."
      />

      <div className="flex items-center gap-3">
        <label className="text-sm font-medium shrink-0">Semester</label>
        <select
          value={semesterId}
          onChange={(e) => handleSemesterChange(e.target.value)}
          className="flex h-9 flex-1 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
        >
          <option value="">Select semester…</option>
          {semesters.map((s) => (
            <option key={s.id} value={s.id}>{s.name} ({s.academic_year})</option>
          ))}
        </select>
      </div>

      {semesterId ? (
        <GenerationPanel semesterId={semesterId} />
      ) : (
        <p className="text-sm text-muted-foreground">Select a semester to start generation.</p>
      )}
    </div>
  )
}
