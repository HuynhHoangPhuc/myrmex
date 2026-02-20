import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useAllSubjects } from '../hooks/use-subjects'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import type { SemesterOffering } from '../types'

// Fetch current offerings for a semester
function useOfferings(semesterId: string) {
  return useQuery({
    queryKey: ['offerings', semesterId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<SemesterOffering[]>(
        `/timetable/semesters/${semesterId}/offerings`,
      )
      return data
    },
    enabled: Boolean(semesterId),
  })
}

function useSaveOfferings(semesterId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (subjectIds: string[]) => {
      await apiClient.put(`/timetable/semesters/${semesterId}/offerings`, { subject_ids: subjectIds })
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['offerings', semesterId] })
    },
  })
}

// Checkbox grid to toggle which subjects are offered in a semester
export function OfferingManager() {
  const { data: semesters = [] } = useAllSemesters()
  const { data: subjects = [] } = useAllSubjects()
  const [selectedSemesterId, setSelectedSemesterId] = React.useState('')
  const { data: offerings = [], isLoading: loadingOfferings } = useOfferings(selectedSemesterId)
  const saveOfferings = useSaveOfferings(selectedSemesterId)

  const [checkedIds, setCheckedIds] = React.useState<Set<string>>(new Set())

  // Sync checkbox state when offerings load
  React.useEffect(() => {
    setCheckedIds(new Set(offerings.map((o) => o.subject_id)))
  }, [offerings])

  function toggle(subjectId: string) {
    setCheckedIds((prev) => {
      const next = new Set(prev)
      next.has(subjectId) ? next.delete(subjectId) : next.add(subjectId)
      return next
    })
  }

  function handleSave() {
    saveOfferings.mutate(Array.from(checkedIds))
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

      {selectedSemesterId && loadingOfferings && <LoadingSpinner />}

      {selectedSemesterId && !loadingOfferings && (
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
            <Button size="sm" onClick={handleSave} disabled={saveOfferings.isPending}>
              {saveOfferings.isPending && <LoadingSpinner size="sm" className="mr-2" />}
              Save Offerings
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
