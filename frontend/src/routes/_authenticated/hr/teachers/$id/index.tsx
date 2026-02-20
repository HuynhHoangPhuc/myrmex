import { createFileRoute, Link } from '@tanstack/react-router'
import { Pencil, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { TeacherAvailabilityForm } from '@/modules/hr/components/teacher-availability-form'
import { useTeacher } from '@/modules/hr/hooks/use-teachers'

export const Route = createFileRoute('/_authenticated/hr/teachers/$id/')({
  component: TeacherDetailPage,
})

function TeacherDetailPage() {
  const { id } = Route.useParams()
  const { data: teacher, isLoading } = useTeacher(id)

  if (isLoading) return <LoadingSpinner />
  if (!teacher) return <p className="text-muted-foreground">Teacher not found.</p>

  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader
        title={teacher.full_name}
        description={`${teacher.employee_code} · ${teacher.department?.name ?? 'No department'}`}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" asChild>
              <Link to="/hr/teachers">
                <ArrowLeft className="mr-2 h-4 w-4" /> Back
              </Link>
            </Button>
            <Button asChild>
              <Link to="/hr/teachers/$id/edit" params={{ id }}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </Link>
            </Button>
          </div>
        }
      />

      {/* Info grid */}
      <div className="grid gap-4 sm:grid-cols-2 rounded-lg border p-5">
        <InfoRow label="Email" value={teacher.email} />
        <InfoRow label="Phone" value={teacher.phone ?? '—'} />
        <InfoRow label="Department" value={teacher.department?.name ?? '—'} />
        <InfoRow label="Max Hours / Week" value={`${teacher.max_hours_per_week}h`} />
        <div className="sm:col-span-2">
          <p className="text-xs text-muted-foreground mb-1">Specializations</p>
          {teacher.specializations?.length ? (
            <div className="flex flex-wrap gap-1">
              {teacher.specializations.map((s) => (
                <Badge key={s} variant="secondary">{s}</Badge>
              ))}
            </div>
          ) : (
            <span className="text-sm text-muted-foreground">None</span>
          )}
        </div>
      </div>

      {/* Availability grid */}
      <div>
        <h2 className="mb-3 text-lg font-semibold">Weekly Availability</h2>
        <TeacherAvailabilityForm
          teacherId={id}
          initialAvailability={teacher.availability ?? []}
        />
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-sm font-medium">{value}</p>
    </div>
  )
}
