import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { AlertTriangle } from 'lucide-react'
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
import { useCheckPrerequisites } from './-hooks/use-my-prerequisites'
import { toast } from '@/lib/hooks/use-toast'
import type { Subject } from '@/modules/subject/types'
import type { Semester } from '@/modules/timetable/types'

export const Route = createFileRoute('/_student/student/subjects')({
  component: StudentSubjectsPage,
})

const STATUS_VARIANT = {
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
} as const

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
  const { data: prereqResult, isLoading: prereqLoading } = useCheckPrerequisites(
    enrollTarget?.subject.id ?? null,
  )

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

  // Maps for enrollment history display
  const subjectMap = React.useMemo(
    () => new Map((subjects?.data ?? []).map((s) => [s.id, `${s.code} — ${s.name}`])),
    [subjects],
  )
  const semesterMap = React.useMemo(
    () => new Map((semesters?.data ?? []).map((s: Semester) => [s.id, s.name ?? s.id])),
    [semesters],
  )

  const filtered = (subjects?.data ?? []).filter(
    (s) =>
      !search ||
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.code.toLowerCase().includes(search.toLowerCase()),
  )

  const hasStrictMissing = prereqResult && !prereqResult.can_enroll &&
    prereqResult.missing.some((m) => m.type === 'strict')

  function handleEnrollSubmit() {
    if (!enrollTarget || !profile) return
    requestMutation.mutate(
      {
        student_id: profile.id,
        semester_id: enrollTarget.semesterId,
        offered_subject_id: enrollTarget.subject.id,
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
    <div className="space-y-6">
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

      {/* Enrollment history */}
      {(myEnrollments?.length ?? 0) > 0 && (
        <div className="space-y-2">
          <h2 className="text-base font-semibold">My Enrollment History</h2>
          <div className="rounded-md border">
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/50">
                <tr>
                  <th className="px-4 py-2 text-left font-medium">Subject</th>
                  <th className="px-4 py-2 text-left font-medium">Semester</th>
                  <th className="px-4 py-2 text-left font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {myEnrollments!.map((e) => (
                  <tr key={e.id} className="border-b last:border-0">
                    <td className="px-4 py-2">{subjectMap.get(e.subject_id) ?? e.subject_id.slice(0, 8)}</td>
                    <td className="px-4 py-2 text-muted-foreground">{semesterMap.get(e.semester_id) ?? e.semester_id.slice(0, 8)}</td>
                    <td className="px-4 py-2">
                      <Badge variant={STATUS_VARIANT[e.status] ?? 'outline'}>{e.status}</Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
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

              {/* Prerequisite check */}
              {prereqLoading && (
                <p className="flex items-center gap-2 text-xs text-muted-foreground">
                  <LoadingSpinner size="sm" /> Checking prerequisites…
                </p>
              )}
              {!prereqLoading && prereqResult && prereqResult.missing.length > 0 && (
                <div className={`rounded-md border px-3 py-2 text-xs ${hasStrictMissing ? 'border-destructive/50 bg-destructive/10 text-destructive' : 'border-amber-500/50 bg-amber-500/10 text-amber-700 dark:text-amber-400'}`}>
                  <div className="flex items-center gap-1.5 font-medium mb-1">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    {hasStrictMissing ? 'Missing required prerequisites' : 'Missing recommended prerequisites'}
                  </div>
                  <ul className="space-y-0.5 pl-1">
                    {prereqResult.missing.map((m) => (
                      <li key={m.subject_id}>
                        {m.subject_code} — {m.subject_name}
                        {m.type === 'recommended' && (
                          <span className="ml-1 opacity-70">(recommended)</span>
                        )}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

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
                <Button
                  onClick={handleEnrollSubmit}
                  disabled={requestMutation.isPending || Boolean(hasStrictMissing)}
                >
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
