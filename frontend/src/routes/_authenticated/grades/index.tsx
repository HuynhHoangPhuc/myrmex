import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useEnrollments } from '@/modules/student/hooks/use-enrollments'
import { useAssignGrade, getLetterGrade } from '@/modules/student/hooks/use-grades'
import { toast } from '@/lib/hooks/use-toast'
import { authStore } from '@/lib/stores/auth-store'
import type { ColumnDef } from '@tanstack/react-table'
import type { EnrollmentRequest } from '@/modules/student/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
})

export const Route = createFileRoute('/_authenticated/grades/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: GradeEntryPage,
})

function GradeEntryPage() {
  const { page, pageSize } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [targetEnrollment, setTargetEnrollment] = React.useState<EnrollmentRequest | null>(null)
  const [gradeValue, setGradeValue] = React.useState('')
  const [notes, setNotes] = React.useState('')

  // List approved enrollments that haven't been graded yet (status=approved)
  const { data, isLoading } = useEnrollments({ page, pageSize, status: 'approved' })
  const assignMutation = useAssignGrade()

  const preview = gradeValue !== '' ? getLetterGrade(Number(gradeValue)) : null

  function handleAssign() {
    if (!targetEnrollment || gradeValue === '') return
    const user = authStore.getUser()
    assignMutation.mutate(
      {
        enrollment_id: targetEnrollment.id,
        grade_numeric: Number(gradeValue),
        graded_by: user?.id ?? '',
        notes: notes || undefined,
      },
      {
        onSuccess: () => {
          toast({ title: 'Grade assigned', description: `${gradeValue} → ${preview}` })
          setTargetEnrollment(null)
          setGradeValue('')
          setNotes('')
        },
        onError: () => toast({ title: 'Failed to assign grade', variant: 'destructive' }),
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
        id: 'actions',
        header: 'Grade',
        cell: ({ row }) => (
          <Button
            size="sm"
            variant="outline"
            className="h-7 text-xs"
            onClick={() => {
              setTargetEnrollment(row.original)
              setGradeValue('')
              setNotes('')
            }}
          >
            Assign Grade
          </Button>
        ),
      },
    ],
    [],
  )

  return (
    <div>
      <PageHeader
        title="Grade Entry"
        description="Assign grades to approved enrollments."
      />

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: { page: p, pageSize } })}
      />

      <Dialog
        open={Boolean(targetEnrollment)}
        onOpenChange={(o) => !o && setTargetEnrollment(null)}
      >
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Assign Grade</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1">
              <Label>Grade (0 – 10)</Label>
              <div className="flex gap-2 items-center">
                <Input
                  type="number"
                  min={0}
                  max={10}
                  step={0.1}
                  value={gradeValue}
                  onChange={(e) => setGradeValue(e.target.value)}
                  placeholder="e.g. 8.5"
                  className="w-32"
                />
                {preview && (
                  <span className="text-lg font-bold text-primary">{preview}</span>
                )}
              </div>
            </div>
            <div className="space-y-1">
              <Label>Notes (optional)</Label>
              <Input
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Optional remarks…"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setTargetEnrollment(null)}>Cancel</Button>
              <Button
                onClick={handleAssign}
                disabled={gradeValue === '' || assignMutation.isPending}
              >
                Save Grade
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
