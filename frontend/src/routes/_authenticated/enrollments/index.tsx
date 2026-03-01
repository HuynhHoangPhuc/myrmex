import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Check, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useEnrollments } from '@/modules/student/hooks/use-enrollments'
import { useReviewEnrollment } from '@/modules/student/hooks/use-enrollments'
import { toast } from '@/lib/hooks/use-toast'
import type { ColumnDef } from '@tanstack/react-table'
import type { EnrollmentRequest } from '@/modules/student/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  status: z.string().optional().catch(undefined),
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

function EnrollmentQueuePage() {
  const { page, pageSize, status } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [rejectId, setRejectId] = React.useState<string | null>(null)
  const [adminNote, setAdminNote] = React.useState('')

  const { data, isLoading } = useEnrollments({
    page,
    pageSize,
    status: status ?? 'pending',
  })
  const reviewMutation = useReviewEnrollment()

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
          <span className="font-mono text-xs">{row.original.student_id.slice(0, 8)}…</span>
        ),
      },
      {
        accessorKey: 'subject_id',
        header: 'Subject',
        cell: ({ row }) => (
          <span className="font-mono text-xs">{row.original.subject_id.slice(0, 8)}…</span>
        ),
      },
      {
        accessorKey: 'semester_id',
        header: 'Semester',
        cell: ({ row }) => (
          <span className="font-mono text-xs">{row.original.semester_id.slice(0, 8)}…</span>
        ),
      },
      {
        accessorKey: 'requested_at',
        header: 'Requested',
        cell: ({ row }) =>
          new Date(row.original.requested_at).toLocaleDateString(),
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
    [reviewMutation.isPending],
  )

  return (
    <div>
      <PageHeader
        title="Enrollment Requests"
        description="Review and process student enrollment requests."
      />

      {/* Status filter */}
      <div className="mb-4 flex gap-2">
        {['pending', 'approved', 'rejected', 'completed'].map((s) => (
          <Button
            key={s}
            size="sm"
            variant={status === s || (!status && s === 'pending') ? 'default' : 'outline'}
            onClick={() => void navigate({ search: { page: 1, pageSize, status: s } })}
            className="capitalize"
          >
            {s}
          </Button>
        ))}
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: { page: p, pageSize, status } })}
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
