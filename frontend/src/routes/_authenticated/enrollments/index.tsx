import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Check, Search, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useEnrollments, useReviewEnrollment } from '@/modules/student/hooks/use-enrollments'
import { useStudents } from '@/modules/student/hooks/use-students'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import { toast } from '@/lib/hooks/use-toast'
import type { ColumnDef } from '@tanstack/react-table'
import type { EnrollmentRequest } from '@/modules/student/types'

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

const STATUS_VARIANT = {
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
} as const

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

  // Auto-select latest (active) semester only when URL has no semesterId at all (first visit).
  // Using sentinel 'all' so "All semesters" is an explicit URL value, not undefined.
  const latestSemester = React.useMemo(() => {
    if (!semesters?.length) return undefined
    return (
      semesters.find((s) => s.is_active) ??
      [...semesters].sort((a, b) => b.start_date.localeCompare(a.start_date))[0]
    )
  }, [semesters])

  React.useEffect(() => {
    if (semesterId === undefined && latestSemester) {
      void navigate({
        search: (prev) => ({ ...prev, semesterId: latestSemester.id }),
        replace: true,
      })
    }
  }, [latestSemester]) // intentionally omit semesterId — only run when semesters first load

  // 'all' sentinel means no semester filter; any real UUID is passed through
  const effectiveSemesterId = semesterId === 'all' ? undefined : semesterId

  const { data, isLoading } = useEnrollments({
    page,
    pageSize,
    semesterId: effectiveSemesterId,
    status: status ?? 'pending',
  })
  const reviewMutation = useReviewEnrollment()

  const studentMap = React.useMemo(
    () => new Map((allStudents?.data ?? []).map((s) => [s.id, s.full_name])),
    [allStudents],
  )
  const subjectMap = React.useMemo(
    () => new Map((allSubjects ?? []).map((s) => [s.id, `${s.code} — ${s.name}`])),
    [allSubjects],
  )
  const semesterMap = React.useMemo(
    () => new Map((semesters ?? []).map((s) => [s.id, s.name])),
    [semesters],
  )

  // Client-side filter for subject and name search (semester is handled server-side)
  const filteredData = React.useMemo(() => {
    let rows = data?.data ?? []
    if (subjectId) rows = rows.filter((e) => e.subject_id === subjectId)
    if (search) {
      const q = search.toLowerCase()
      rows = rows.filter((e) => (studentMap.get(e.student_id) ?? '').toLowerCase().includes(q))
    }
    return rows
  }, [data, subjectId, search, studentMap])

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

  const columns = React.useMemo<ColumnDef<EnrollmentRequest>[]>(
    () => [
      {
        accessorKey: 'student_id',
        header: 'Student',
        cell: ({ row }) => (
          <span className="text-sm font-medium">
            {studentMap.get(row.original.student_id) ?? row.original.student_id.slice(0, 8)}
          </span>
        ),
      },
      {
        accessorKey: 'subject_id',
        header: 'Subject',
        cell: ({ row }) => (
          <span className="text-sm">
            {subjectMap.get(row.original.subject_id) ?? row.original.subject_id.slice(0, 8)}
          </span>
        ),
      },
      {
        accessorKey: 'semester_id',
        header: 'Semester',
        cell: ({ row }) => (
          <span className="text-sm">
            {semesterMap.get(row.original.semester_id) ?? row.original.semester_id.slice(0, 8)}
          </span>
        ),
      },
      {
        accessorKey: 'requested_at',
        header: 'Requested',
        cell: ({ row }) => new Date(row.original.requested_at).toLocaleDateString(),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => (
          <Badge variant={STATUS_VARIANT[row.original.status] ?? 'outline'}>
            {row.original.status}
          </Badge>
        ),
      },
      {
        id: 'actions',
        cell: ({ row }) => {
          if (row.original.status !== 'pending') return null
          return (
            <div className="flex gap-1">
              <Button
                size="sm"
                variant="outline"
                className="h-7 text-xs"
                onClick={() => approve(row.original.id)}
                disabled={reviewMutation.isPending}
              >
                <Check className="mr-1 h-3 w-3" /> Approve
              </Button>
              <Button
                size="sm"
                variant="outline"
                className="h-7 text-xs text-destructive"
                onClick={() => { setRejectId(row.original.id); setAdminNote('') }}
              >
                <X className="mr-1 h-3 w-3" /> Reject
              </Button>
            </div>
          )
        },
      },
    ],
    [studentMap, subjectMap, semesterMap, reviewMutation.isPending],
  )

  return (
    <div>
      <PageHeader
        title="Enrollment Requests"
        description="Review and process student enrollment requests."
      />

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
              void navigate({
                search: (prev) => ({ ...prev, page: 1, search: e.target.value || undefined }),
              })
            }
          />
        </div>

        <select
          className={SELECT_CLS + ' w-48'}
          value={semesterId ?? 'all'}
          onChange={(e) =>
            void navigate({
              // Use 'all' sentinel so semesterId is never undefined after user interaction
              search: (prev) => ({ ...prev, page: 1, semesterId: e.target.value || 'all' }),
            })
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
            void navigate({
              search: (prev) => ({ ...prev, page: 1, subjectId: e.target.value || undefined }),
            })
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

      {/* Reject dialog */}
      <Dialog open={Boolean(rejectId)} onOpenChange={(o) => !o && setRejectId(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Reject Enrollment</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1">
              <Label>Reason (optional)</Label>
              <Input
                value={adminNote}
                onChange={(e) => setAdminNote(e.target.value)}
                placeholder="Explain the reason…"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setRejectId(null)}>Cancel</Button>
              <Button variant="destructive" onClick={reject} disabled={reviewMutation.isPending}>
                Reject
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
