import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Search } from 'lucide-react'
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
import { useStudents } from '@/modules/student/hooks/use-students'
import { useAssignGrade, getLetterGrade } from '@/modules/student/hooks/use-grades'
import { useAllSubjects } from '@/modules/subject/hooks/use-subjects'
import { useAllSemesters } from '@/modules/timetable/hooks/use-semesters'
import { toast } from '@/lib/hooks/use-toast'
import { authStore } from '@/lib/stores/auth-store'
import type { ColumnDef } from '@tanstack/react-table'
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
  const [gradeValue, setGradeValue] = React.useState('')
  const [notes, setNotes] = React.useState('')

  const { data: semesters } = useAllSemesters()
  const { data: allSubjects } = useAllSubjects()
  const { data: allStudents } = useStudents({ page: 1, pageSize: 1000 })

  // Auto-select latest (active) semester only when URL has no semesterId at all (first visit).
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

  // List approved enrollments that haven't been graded yet (status=approved)
  const { data, isLoading } = useEnrollments({ page, pageSize, semesterId: effectiveSemesterId, status: 'approved' })
  const assignMutation = useAssignGrade()

  const studentMap = React.useMemo(
    () => new Map((allStudents?.data ?? []).map((s) => [s.id, s.full_name])),
    [allStudents],
  )
  const subjectMap = React.useMemo(
    () => new Map((allSubjects ?? []).map((s) => [s.id, `${s.code} — ${s.name}`])),
    [allSubjects],
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
    [studentMap, subjectMap],
  )

  return (
    <div>
      <PageHeader
        title="Grade Entry"
        description="Assign grades to approved enrollments."
      />

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

      <Dialog
        open={Boolean(targetEnrollment)}
        onOpenChange={(o) => !o && setTargetEnrollment(null)}
      >
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>
              Assign Grade
              {targetEnrollment && (
                <span className="block text-sm font-normal text-muted-foreground mt-0.5">
                  {subjectMap.get(targetEnrollment.subject_id) ?? 'Subject'} ·{' '}
                  {studentMap.get(targetEnrollment.student_id) ?? 'Student'}
                </span>
              )}
            </DialogTitle>
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
