import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Input } from '@/components/ui/input'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useSubjects } from '@/modules/subject/hooks/use-subjects'
import { useSemesters } from '@/modules/timetable/hooks/use-semesters'
import { useMyStudentProfile } from './-hooks/use-my-profile'
import { useMyEnrollments, useRequestEnrollment } from './-hooks/use-my-enrollments'
import { toast } from '@/lib/hooks/use-toast'
import { SubjectList } from './components/subject-list'
import { EnrollmentHistoryTable } from './components/enrollment-history-table'
import { EnrollDialog } from './components/enroll-dialog'
import type { Subject } from '@/modules/subject/types'
import type { Semester } from '@/modules/timetable/types'

export const Route = createFileRoute('/_student/student/subjects')({
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

      <SubjectList
        subjects={filtered}
        enrolledSubjectIds={enrolledSubjectIds}
        selectedSemesterId={selectedSemesterId}
        onEnroll={(subject) => { setEnrollTarget({ subject, semesterId: selectedSemesterId }); setNote('') }}
      />

      <EnrollmentHistoryTable
        enrollments={myEnrollments ?? []}
        subjectMap={subjectMap}
        semesterMap={semesterMap}
      />

      <EnrollDialog
        target={enrollTarget}
        note={note}
        isPending={requestMutation.isPending}
        onNoteChange={setNote}
        onConfirm={handleEnrollSubmit}
        onClose={() => setEnrollTarget(null)}
      />
    </div>
  )
}
