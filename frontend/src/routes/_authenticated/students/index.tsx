import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { useStudents } from '@/modules/student/hooks/use-students'
import { useCreateStudent, useDeleteStudent } from '@/modules/student/hooks/use-student-mutations'
import { buildStudentColumns } from '@/modules/student/components/student-columns'
import { useDepartments } from '@/modules/hr/hooks/use-departments'
import { toast } from '@/lib/hooks/use-toast'
import type { CreateStudentInput } from '@/modules/student/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
})

export const Route = createFileRoute('/_authenticated/students/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: StudentListPage,
})

const DEFAULT_FORM: CreateStudentInput = {
  student_code: '',
  full_name: '',
  email: '',
  department_id: '',
  enrollment_year: new Date().getFullYear(),
}

function StudentListPage() {
  const { page, pageSize } = Route.useSearch()
  const navigate = Route.useNavigate()
  const [deleteId, setDeleteId] = React.useState<string | null>(null)
  const [createOpen, setCreateOpen] = React.useState(false)
  const [tempPassword, setTempPassword] = React.useState<string | null>(null)
  const [form, setForm] = React.useState<CreateStudentInput>(DEFAULT_FORM)

  const { data, isLoading } = useStudents({ page, pageSize })
  const { data: depts } = useDepartments({ page: 1, pageSize: 100 })
  const createMutation = useCreateStudent()
  const deleteMutation = useDeleteStudent()

  const columns = React.useMemo(() => buildStudentColumns(setDeleteId), [])

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    createMutation.mutate(form, {
      onSuccess: (res) => {
        setTempPassword(res.temp_password)
        setForm(DEFAULT_FORM)
        toast({ title: 'Student created', description: res.student.full_name })
      },
      onError: () => toast({ title: 'Failed to create student', variant: 'destructive' }),
    })
  }

  return (
    <div>
      <PageHeader
        title="Students"
        description="Manage enrolled students and their accounts."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" /> New Student
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: { page: p, pageSize } })}
      />

      {/* Create student dialog */}
      <Dialog open={createOpen} onOpenChange={(o) => { setCreateOpen(o); if (!o) setTempPassword(null) }}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>New Student</DialogTitle>
          </DialogHeader>

          {tempPassword ? (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">Student created successfully. Share this temporary password:</p>
              <div className="rounded-md bg-muted px-4 py-2 font-mono text-sm font-bold">{tempPassword}</div>
              <p className="text-xs text-muted-foreground">The student should change this on first login.</p>
              <Button className="w-full" onClick={() => { setCreateOpen(false); setTempPassword(null) }}>Done</Button>
            </div>
          ) : (
            <form onSubmit={handleCreate} className="space-y-4">
              <Field label="Student Code" required>
                <Input
                  value={form.student_code}
                  onChange={(e) => setForm({ ...form, student_code: e.target.value })}
                  placeholder="SV2025001"
                />
              </Field>
              <Field label="Full Name" required>
                <Input
                  value={form.full_name}
                  onChange={(e) => setForm({ ...form, full_name: e.target.value })}
                  placeholder="Nguyen Van A"
                />
              </Field>
              <Field label="Email" required>
                <Input
                  type="email"
                  value={form.email}
                  onChange={(e) => setForm({ ...form, email: e.target.value })}
                  placeholder="student@university.edu"
                />
              </Field>
              <Field label="Department" required>
                <select
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                  value={form.department_id}
                  onChange={(e) => setForm({ ...form, department_id: e.target.value })}
                  required
                >
                  <option value="">Select department…</option>
                  {depts?.data.map((d) => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </select>
              </Field>
              <Field label="Enrollment Year" required>
                <Input
                  type="number"
                  value={form.enrollment_year}
                  onChange={(e) => setForm({ ...form, enrollment_year: Number(e.target.value) })}
                  min={2000}
                  max={2100}
                />
              </Field>
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={createMutation.isPending}>Create</Button>
              </div>
            </form>
          )}
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(o) => !o && setDeleteId(null)}
        title="Delete Student"
        description="The student record will be deactivated. This cannot be undone."
        confirmLabel="Delete"
        variant="destructive"
        isLoading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteId) return
          deleteMutation.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />
    </div>
  )
}

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <Label>{label}{required && <span className="ml-0.5 text-destructive">*</span>}</Label>
      {children}
    </div>
  )
}
