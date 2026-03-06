import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { useEnrollments } from '@/modules/student/hooks/use-enrollments'
import { useStudents } from '@/modules/student/hooks/use-students'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import { buildGradeEntryColumns } from './components/grade-entry-columns'
import { GradeDialog } from './components/grade-dialog'
import { useGradeAssignment } from './hooks/use-grade-assignment'
import { useEnrollmentFilters } from '../enrollments/hooks/use-enrollment-filters'
import type { EnrollmentRequest } from '@/modules/student/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  semesterId: z.string().optional().catch(undefined),
  subjectId: z.string().optional().catch(undefined),
  search: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/grades/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: GradeEntryPage,
})

const SELECT_CLS =
  'h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring'

function GradeEntryPage() {
  const { page, pageSize, semesterId, subjectId, search } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [targetEnrollment, setTargetEnrollment] = React.useState<EnrollmentRequest | null>(null)

  const { data: semesters } = useAllSemesters()
  const { data: allSubjects } = useAllSubjects()
  const { data: allStudents } = useStudents({ page: 1, pageSize: 1000 })

  const { data, isLoading } = useEnrollments({
    page, pageSize,
    semesterId: semesterId === 'all' ? undefined : semesterId,
    status: 'approved',
  })

  const { latestSemester, studentMap, subjectMap, filteredData } = useEnrollmentFilters({
    semesters: semesters,
    allSubjects: allSubjects,
    allStudents: allStudents,
    data,
    semesterId,
    subjectId,
    search,
  })

  const gradeAssignment = useGradeAssignment()

  // Auto-select latest (active) semester only on first visit
  React.useEffect(() => {
    if (semesterId === undefined && latestSemester) {
      void navigate({ search: (prev) => ({ ...prev, semesterId: latestSemester.id }), replace: true })
    }
  }, [latestSemester]) // intentionally omit semesterId — only run when semesters first load

  const columns = React.useMemo(
    () => buildGradeEntryColumns({
      studentMap, subjectMap,
      onAssign: (enrollment) => {
        setTargetEnrollment(enrollment)
        gradeAssignment.resetForm()
      },
    }),
    [studentMap, subjectMap],
  )

  return (
    <div>
      <PageHeader title="Grade Entry" description="Assign grades to approved enrollments." />

      {/* Search / Semester / Subject filters */}
      <div className="mb-4 flex flex-wrap gap-3">
        <div className="relative min-w-48 flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            className="pl-8"
            placeholder="Search student name…"
            value={search ?? ''}
            onChange={(e) =>
              void navigate({ search: (prev) => ({ ...prev, page: 1, search: e.target.value || undefined }) })
            }
          />
        </div>
        <select
          className={SELECT_CLS + ' w-48'}
          value={semesterId ?? 'all'}
          onChange={(e) =>
            void navigate({ search: (prev) => ({ ...prev, page: 1, semesterId: e.target.value || 'all' }) })
          }
        >
          <option value="all">All semesters</option>
          {(semesters ?? []).map((s) => (
            <option key={s.id} value={s.id}>{s.name}</option>
          ))}
        </select>
        <select
          className={SELECT_CLS + ' w-64'}
          value={subjectId ?? ''}
          onChange={(e) =>
            void navigate({ search: (prev) => ({ ...prev, page: 1, subjectId: e.target.value || undefined }) })
          }
        >
          <option value="">All subjects</option>
          {(allSubjects ?? []).map((s) => (
            <option key={s.id} value={s.id}>{s.code} — {s.name}</option>
          ))}
        </select>
      </div>

      <DataTable
        columns={columns}
        data={filteredData}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: (prev) => ({ ...prev, page: p }) })}
      />

      <GradeDialog
        enrollment={targetEnrollment}
        gradeValue={gradeAssignment.gradeValue}
        notes={gradeAssignment.notes}
        preview={gradeAssignment.preview}
        isPending={gradeAssignment.isPending}
        subjectMap={subjectMap}
        studentMap={studentMap}
        onGradeChange={gradeAssignment.setGradeValue}
        onNotesChange={gradeAssignment.setNotes}
        onConfirm={() => gradeAssignment.handleAssign(targetEnrollment, () => setTargetEnrollment(null))}
        onClose={() => setTargetEnrollment(null)}
      />
    </div>
  )
}
