import * as React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useToast } from '@/lib/hooks/use-toast'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { useSemester } from '@/modules/timetable/hooks/use-semesters'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'

interface SemesterWizardStepOfferingsProps {
  semesterId: string
  onComplete?: () => void
}

export function SemesterWizardStepOfferings({
  semesterId,
  onComplete,
}: SemesterWizardStepOfferingsProps) {
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const { data: semester, isLoading: semesterLoading } = useSemester(semesterId)
  const { data: subjects = [], isLoading: subjectsLoading } = useAllSubjects()
  const [checkedIds, setCheckedIds] = React.useState<Set<string>>(new Set())

  React.useEffect(() => {
    setCheckedIds(new Set(semester?.offered_subject_ids ?? []))
  }, [semester?.offered_subject_ids])

  const syncOfferings = useMutation({
    mutationFn: async () => {
      const current = new Set(semester?.offered_subject_ids ?? [])
      const desired = checkedIds
      const toAdd = [...desired].filter((id) => !current.has(id))
      const toRemove = [...current].filter((id) => !desired.has(id))

      for (const id of toAdd) {
        await apiClient.post(ENDPOINTS.timetable.offeredSubjects(semesterId), { subject_id: id })
      }

      for (const id of toRemove) {
        await apiClient.delete(`${ENDPOINTS.timetable.offeredSubjects(semesterId)}/${id}`)
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['semesters', semesterId] })
      toast({
        title: 'Offerings saved',
        description: `${checkedIds.size} subject(s) available for scheduling.`,
      })
      onComplete?.()
    },
    onError: () => {
      toast({ title: 'Failed to save offerings', variant: 'destructive' })
    },
  })

  const isLoading = semesterLoading || subjectsLoading

  function toggle(subjectId: string) {
    setCheckedIds((current) => {
      const next = new Set(current)
      if (next.has(subjectId)) next.delete(subjectId)
      else next.add(subjectId)
      return next
    })
  }

  if (isLoading) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <LoadingSpinner />
      </div>
    )
  }

  if (!semester) {
    return <p className="text-sm text-muted-foreground">Semester not found.</p>
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-lg border bg-card p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h3 className="font-semibold">Select offered subjects</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Choose which subjects can be scheduled for {semester.name}.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" size="sm" onClick={() => setCheckedIds(new Set(subjects.map((subject) => subject.id)))}>
            Select all
          </Button>
          <Button type="button" variant="outline" size="sm" onClick={() => setCheckedIds(new Set())}>
            Clear
          </Button>
        </div>
      </div>

      <div className="max-h-[420px] overflow-y-auto rounded-lg border divide-y">
        {subjects.map((subject) => (
          <label
            key={subject.id}
            className="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/40"
          >
            <input
              type="checkbox"
              checked={checkedIds.has(subject.id)}
              onChange={() => toggle(subject.id)}
              className="h-4 w-4"
            />
            <span className="w-20 shrink-0 font-mono text-xs text-primary">{subject.code}</span>
            <span className="min-w-0 flex-1 text-sm">{subject.name}</span>
            <span className="shrink-0 text-xs text-muted-foreground">{subject.credits} cr</span>
          </label>
        ))}
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">{checkedIds.size} subject(s) selected</p>
        <Button type="button" onClick={() => syncOfferings.mutate()} disabled={syncOfferings.isPending}>
          {syncOfferings.isPending && <LoadingSpinner size="sm" className="mr-2" />}
          Save and continue
        </Button>
      </div>
    </div>
  )
}
