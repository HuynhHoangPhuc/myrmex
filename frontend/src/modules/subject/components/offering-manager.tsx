import * as React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useAllSubjects } from '../hooks/use-subjects'
import { useAllSemesters, useSemester } from '@/modules/timetable/hooks/use-semesters'

// Add a single subject offering to a semester
function useAddOffering(semesterId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (subjectId: string) => {
      await apiClient.post(ENDPOINTS.timetable.offeredSubjects(semesterId), { subject_id: subjectId })
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['semesters', semesterId] })
    },
  })
}

// Remove a single subject offering from a semester
function useRemoveOffering(semesterId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (subjectId: string) => {
      await apiClient.delete(`${ENDPOINTS.timetable.offeredSubjects(semesterId)}/${subjectId}`)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['semesters', semesterId] })
    },
  })
}

// Checkbox grid to toggle which subjects are offered in a semester.
// Uses per-item add/remove since backend has no bulk update endpoint.
export function OfferingManager() {
  const { data: semesters = [] } = useAllSemesters()
  const { data: subjects = [] } = useAllSubjects()
  const [selectedSemesterId, setSelectedSemesterId] = React.useState('')

  const { data: semester, isLoading: loadingSemester } = useSemester(selectedSemesterId)
  const currentOfferings = semester?.offered_subject_ids ?? []

  // Local checkbox state — initialized from server data
  const [checkedIds, setCheckedIds] = React.useState<Set<string>>(new Set())
  React.useEffect(() => {
    setCheckedIds(new Set(currentOfferings))
  }, [currentOfferings.join(',')])  // eslint-disable-line react-hooks/exhaustive-deps

  const addOffering = useAddOffering(selectedSemesterId)
  const removeOffering = useRemoveOffering(selectedSemesterId)
  const isSaving = addOffering.isPending || removeOffering.isPending

  function toggle(subjectId: string) {
    setCheckedIds((prev) => {
      const next = new Set(prev)
      next.has(subjectId) ? next.delete(subjectId) : next.add(subjectId)
      return next
    })
  }

  async function handleSave() {
    const current = new Set(currentOfferings)
    const desired = checkedIds
    const toAdd = [...desired].filter((id) => !current.has(id))
    const toRemove = [...current].filter((id) => !desired.has(id))
    // Sequential to avoid race conditions on the server
    for (const id of toAdd) await addOffering.mutateAsync(id)
    for (const id of toRemove) await removeOffering.mutateAsync(id)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <label className="text-sm font-medium">Semester</label>
        <select
          value={selectedSemesterId}
          onChange={(e) => setSelectedSemesterId(e.target.value)}
          className="flex h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
        >
          <option value="">Select semester…</option>
          {semesters.map((s) => (
            <option key={s.id} value={s.id}>{s.name} ({s.academic_year})</option>
          ))}
        </select>
      </div>

      {!selectedSemesterId && (
        <p className="text-sm text-muted-foreground">Select a semester to manage subject offerings.</p>
      )}

      {selectedSemesterId && loadingSemester && <LoadingSpinner />}

      {selectedSemesterId && !loadingSemester && (
        <>
          <div className="rounded-md border divide-y max-h-[480px] overflow-y-auto">
            {subjects.map((s) => (
              <label key={s.id} className="flex items-center gap-3 px-4 py-2.5 hover:bg-muted/50 cursor-pointer">
                <input
                  type="checkbox"
                  checked={checkedIds.has(s.id)}
                  onChange={() => toggle(s.id)}
                  className="h-4 w-4"
                />
                <span className="font-mono text-xs text-primary w-20 shrink-0">{s.code}</span>
                <span className="text-sm flex-1">{s.name}</span>
                <span className="text-xs text-muted-foreground">{s.credits} cr</span>
              </label>
            ))}
          </div>

          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">{checkedIds.size} subject(s) offered</p>
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              {isSaving && <LoadingSpinner size="sm" className="mr-2" />}
              Save Offerings
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
