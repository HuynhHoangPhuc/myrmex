// Column definitions for the enrollment queue data table.
import { Check, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { ColumnDef } from '@tanstack/react-table'
import type { EnrollmentRequest } from '@/modules/student/types'

const STATUS_VARIANT = {
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
} as const

interface BuildColumnsOptions {
  studentMap: Map<string, string>
  subjectMap: Map<string, string>
  semesterMap: Map<string, string>
  isPending: boolean
  onApprove: (id: string) => void
  onRejectOpen: (id: string) => void
}

export function buildEnrollmentColumns({
  studentMap,
  subjectMap,
  semesterMap,
  isPending,
  onApprove,
  onRejectOpen,
}: BuildColumnsOptions): ColumnDef<EnrollmentRequest>[] {
  return [
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
              onClick={() => onApprove(row.original.id)}
              disabled={isPending}
            >
              <Check className="mr-1 h-3 w-3" /> Approve
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="h-7 text-xs text-destructive"
              onClick={() => onRejectOpen(row.original.id)}
            >
              <X className="mr-1 h-3 w-3" /> Reject
            </Button>
          </div>
        )
      },
    },
  ]
}
