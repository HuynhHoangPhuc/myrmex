import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSubjects } from '@/modules/subject/hooks/use-subjects'
import { useSemesters } from '@/modules/timetable/hooks/use-semesters'
import { useMyStudentProfile } from './-hooks/use-my-profile'
import { useMyEnrollments, useRequestEnrollment } from './-hooks/use-my-enrollments'
import { toast } from '@/lib/hooks/use-toast'
import type { Subject } from '@/modules/subject/types'
import type { Semester } from '@/modules/timetable/types'

export const Route = createFileRoute('/_student/subjects')({
  component: StudentSubjectsPage,
})

interface EnrollDialogState {
  subject: Subject
  semesterId: string
}

function StudentSubjectsPage() {
  const [search, setSearch] = React.useState('')
  const [selectedSemesterId, setSelectedSemesterId] = React.useState('')
  const [enrollTarget, setEnrollTarget] = React.useState<EnrollDialogState | null>(null)
  const [note, setNote] = React.useState('')

  const { data: profile } = useMyStudentProfile()
  const { data: subjects, isLoading } = useSubjects({ page: 1, pageSize: 100 })
  const { data: semesters } = useSemesters({ page: 1, pageSize: 20 })
  const { data: myEnrollments } = useMyEnrollments()
  const requestMutation = useRequestEnrollment()

  // Set first semester as default when loaded
  React.useEffect(() => {
    if (!selectedSemesterId && semesters?.data[0]) {
      setSelectedSemesterId(semesters.data[0].id)
    }
  }, [semesters, selectedSemesterId])

  // Build set of already-enrolled subject IDs for this semester
  const enrolledSubjectIds = React.useMemo(() => {
    const ids = new Set<string>()
    myEnrollments?.forEach((e) => {
      if (e.semester_id === selectedSemesterId) ids.add(e.subject_id)
    })
    return ids
  }, [myEnrollments, selectedSemesterId])

  const filtered = (subjects?.data ?? []).filter(
    (s) =>
      !search ||
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.code.toLowerCase().includes(search.toLowerCase()),
  )

  function handleEnrollSubmit() {
    if (!enrollTarget || !profile) return
    requestMutation.mutate(
      {
        student_id: profile.id,
        semester_id: enrollTarget.semesterId,
        offered_subject_id: enrollTarget.subject.id, // using subject id as offered_subject_id for MVP
        subject_id: enrollTarget.subject.id,
        request_note: note || undefined,
      },
      {
        onSuccess: () => {
          toast({ title: 'Enrollment requested', description: enrollTarget.subject.name })
          setEnrollTarget(null)
          setNote('')
        },
        onError: () => toast({ title: 'Failed to request enrollment', variant: 'destructive' }),
      },
    )
  }

  if (isLoading) return <LoadingSpinner />

  const activeSemester = semesters?.data.find((s: Semester) => s.id === selectedSemesterId)

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">My Subjects</h1>

      {/* Semester selector + search */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <select
          className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
          value={selectedSemesterId}
          onChange={(e) => setSelectedSemesterId(e.target.value)}
        >
          {semesters?.data.map((s: Semester) => (
            <option key={s.id} value={s.id}>{s.name ?? s.id}</option>
          ))}
        </select>
        <Input
          placeholder="Search subjects…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full sm:w-64"
        />
      </div>

      {activeSemester && (
        <p className="text-xs text-muted-foreground">
          Active semester: <strong>{activeSemester.name ?? activeSemester.id}</strong>
        </p>
      )}

      {/* Subject list */}
      {filtered.length === 0 ? (
        <p className="text-sm text-muted-foreground">No subjects found.</p>
      ) : (
        <div className="space-y-2">
          {filtered.map((subject) => {
            const alreadyEnrolled = enrolledSubjectIds.has(subject.id)
            return (
              <div
                key={subject.id}
                className="flex items-center justify-between rounded-md border px-4 py-3 text-sm"
              >
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-bold">{subject.code}</span>
                    <span className="font-medium">{subject.name}</span>
                    {subject.credits != null && (
                      <Badge variant="outline" className="text-xs">{subject.credits} cr</Badge>
                    )}
                  </div>
                </div>
                {alreadyEnrolled ? (
                  <Badge variant="secondary">Enrolled</Badge>
                ) : (
                  <Button
                    size="sm"
                    variant="outline"
                    className="h-7 text-xs"
                    disabled={!selectedSemesterId}
                    onClick={() => {
                      setEnrollTarget({ subject, semesterId: selectedSemesterId })
                      setNote('')
                    }}
                  >
                    Enroll →
                  </Button>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* Enrollment request dialog */}
      <Dialog open={Boolean(enrollTarget)} onOpenChange={(o) => !o && setEnrollTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Request Enrollment</DialogTitle>
          </DialogHeader>
          {enrollTarget && (
            <div className="space-y-4">
              <div className="rounded-md bg-muted px-4 py-2 text-sm">
                <p className="font-medium">{enrollTarget.subject.name}</p>
                <p className="text-xs text-muted-foreground">{enrollTarget.subject.code}</p>
              </div>
              <div className="space-y-1">
                <Label>Note (optional)</Label>
                <Input
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Any special request…"
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setEnrollTarget(null)}>Cancel</Button>
                <Button onClick={handleEnrollSubmit} disabled={requestMutation.isPending}>
                  Submit Request
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
