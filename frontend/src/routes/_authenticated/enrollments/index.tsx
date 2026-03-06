import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { useEnrollments, useReviewEnrollment } from '@/modules/student/hooks/use-enrollments'
import { useStudents } from '@/modules/student/hooks/use-students'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import { toast } from '@/lib/hooks/use-toast'
import { buildEnrollmentColumns } from './components/enrollment-columns'
import { RejectEnrollmentDialog } from './components/reject-enrollment-dialog'
import { useEnrollmentFilters } from './hooks/use-enrollment-filters'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  status: z.string().optional().catch(undefined),
  semesterId: z.string().optional().catch(undefined),
  subjectId: z.string().optional().catch(undefined),
  search: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/enrollments/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: EnrollmentQueuePage,
})

const SELECT_CLS =
  'h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring'

function EnrollmentQueuePage() {
  const { page, pageSize, status, semesterId, subjectId, search } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [rejectId, setRejectId] = React.useState<string | null>(null)
  const [adminNote, setAdminNote] = React.useState('')

  const { data: semesters } = useAllSemesters()
  const { data: allSubjects } = useAllSubjects()
  const { data: allStudents } = useStudents({ page: 1, pageSize: 1000 })

  const { data, isLoading } = useEnrollments({
    page,
    pageSize,
    semesterId: semesterId === 'all' ? undefined : semesterId,
    status: status ?? 'pending',
  })
  const reviewMutation = useReviewEnrollment()

  const { latestSemester, studentMap, subjectMap, semesterMap, filteredData } = useEnrollmentFilters({
    semesters: semesters,
    allSubjects: allSubjects,
    allStudents: allStudents,
    data,
    semesterId,
    subjectId,
    search,
  })

  // Auto-select latest (active) semester only on first visit (no semesterId in URL)
  React.useEffect(() => {
    if (semesterId === undefined && latestSemester) {
      void navigate({ search: (prev) => ({ ...prev, semesterId: latestSemester.id }), replace: true })
    }
  }, [latestSemester]) // intentionally omit semesterId — only run when semesters first load

  function approve(id: string) {
    reviewMutation.mutate(
      { id, input: { approve: true } },
      {
        onSuccess: () => toast({ title: 'Enrollment approved' }),
        onError: () => toast({ title: 'Failed to approve', variant: 'destructive' }),
      },
    )
  }

  function reject() {
    if (!rejectId) return
    reviewMutation.mutate(
      { id: rejectId, input: { approve: false, admin_note: adminNote } },
      {
        onSuccess: () => {
          toast({ title: 'Enrollment rejected' })
          setRejectId(null)
          setAdminNote('')
        },
        onError: () => toast({ title: 'Failed to reject', variant: 'destructive' }),
      },
    )
  }

  const columns = React.useMemo(
    () => buildEnrollmentColumns({
      studentMap, subjectMap, semesterMap,
      isPending: reviewMutation.isPending,
      onApprove: approve,
      onRejectOpen: (id) => { setRejectId(id); setAdminNote('') },
    }),
    [studentMap, subjectMap, semesterMap, reviewMutation.isPending],
  )

  return (
    <div>
      <PageHeader title="Enrollment Requests" description="Review and process student enrollment requests." />

      {/* Status filter */}
      <div className="mb-4 flex gap-2">
        {(['pending', 'approved', 'rejected', 'completed'] as const).map((s) => (
          <Button
            key={s}
            size="sm"
            variant={status === s || (!status && s === 'pending') ? 'default' : 'outline'}
            onClick={() => void navigate({ search: (prev) => ({ ...prev, page: 1, status: s }) })}
            className="capitalize"
          >
            {s}
          </Button>
        ))}
      </div>

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

      <RejectEnrollmentDialog
        open={Boolean(rejectId)}
        adminNote={adminNote}
        isPending={reviewMutation.isPending}
        onNoteChange={setAdminNote}
        onConfirm={reject}
        onClose={() => setRejectId(null)}
      />
    </div>
  )
}
