import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useStudent } from '@/modules/student/hooks/use-students'
import { useEnrollments } from '@/modules/student/hooks/use-enrollments'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'
import { useSemesters } from '@/modules/timetable/hooks/use-semesters'
import type { EnrollmentRequest } from '@/modules/student/types'

export const Route = createFileRoute('/_authenticated/students/$studentId/')({
  component: StudentDetailPage,
})

const STATUS_VARIANT = {
  // enrollment statuses
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
  // student statuses
  active: 'default',
  graduated: 'outline',
  suspended: 'destructive',
} as const

function StudentDetailPage() {
  const { studentId } = Route.useParams()
  const [tab, setTab] = React.useState<'enrollments' | 'grades'>('enrollments')
  const { data: student, isLoading } = useStudent(studentId)
  const { data: enrollments } = useEnrollments({ page: 1, pageSize: 100, studentId })
  const { data: allSubjects } = useAllSubjects()
  const { data: allSemesters } = useSemesters({ page: 1, pageSize: 100 })

  const subjectMap = React.useMemo(
    () => new Map((allSubjects ?? []).map((s) => [s.id, `${s.code} — ${s.name}`])),
    [allSubjects],
  )
  const semesterMap = React.useMemo(
    () => new Map((allSemesters?.data ?? []).map((s) => [s.id, s.name])),
    [allSemesters],
  )

  if (isLoading) return <LoadingSpinner />
  if (!student) return <p className="text-muted-foreground">Student not found.</p>

  const graded = enrollments?.data.filter((e) => e.status === 'completed') ?? []

  return (
    <div className="max-w-3xl space-y-6">
      <PageHeader
        title={student.full_name}
        description={`${student.student_code} · Year ${student.enrollment_year}`}
        actions={
          <Button variant="outline" asChild>
            <Link to="/students" search={{ page: 1, pageSize: 25 }}>
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
            </Link>
          </Button>
        }
      />

      {/* Info grid */}
      <div className="grid gap-4 sm:grid-cols-2 rounded-lg border p-5">
        <InfoRow label="Email" value={student.email} />
        <InfoRow label="Enrollment Year" value={String(student.enrollment_year)} />
        <InfoRow label="Status">
          <Badge variant={STATUS_VARIANT[student.status] ?? 'outline'}>{student.status}</Badge>
        </InfoRow>
        <InfoRow label="Account" value={student.user_id ? 'Linked' : 'Not yet activated'} />
      </div>

      {/* Tabs */}
      <div>
        <div className="flex gap-1 border-b mb-4">
          {(['enrollments', 'grades'] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2 text-sm font-medium capitalize transition-colors ${
                tab === t
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {t === 'grades' ? 'Grades Summary' : 'Enrollments'}
            </button>
          ))}
        </div>

        {tab === 'enrollments' && (
          <EnrollmentList enrollments={enrollments?.data ?? []} subjectMap={subjectMap} />
        )}
        {tab === 'grades' && (
          <GradeList enrollments={graded} subjectMap={subjectMap} semesterMap={semesterMap} />
        )}
      </div>
    </div>
  )
}

function EnrollmentList({ enrollments, subjectMap }: {
  enrollments: EnrollmentRequest[]
  subjectMap: Map<string, string>
}) {
  if (!enrollments.length) return <p className="text-sm text-muted-foreground">No enrollment requests.</p>
  return (
    <div className="space-y-2">
      {enrollments.map((e) => (
        <div key={e.id} className="flex items-center justify-between rounded-md border px-4 py-2 text-sm">
          <span>{subjectMap.get(e.subject_id) ?? e.subject_id.slice(0, 8)}</span>
          <Badge variant={STATUS_VARIANT[e.status] ?? 'outline'}>{e.status}</Badge>
        </div>
      ))}
    </div>
  )
}

function GradeList({ enrollments, subjectMap, semesterMap }: {
  enrollments: EnrollmentRequest[]
  subjectMap: Map<string, string>
  semesterMap: Map<string, string>
}) {
  if (!enrollments.length) return <p className="text-sm text-muted-foreground">No grades recorded yet.</p>
  return (
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
          {enrollments.map((e) => (
            <tr key={e.id} className="border-b last:border-0">
              <td className="px-4 py-2">{subjectMap.get(e.subject_id) ?? e.subject_id.slice(0, 8)}</td>
              <td className="px-4 py-2 text-muted-foreground">{semesterMap.get(e.semester_id) ?? e.semester_id.slice(0, 8)}</td>
              <td className="px-4 py-2">
                <Badge variant="outline">{e.status}</Badge>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function InfoRow({ label, value, children }: { label: string; value?: string; children?: React.ReactNode }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-sm font-medium">{children ?? value}</p>
    </div>
  )
}
