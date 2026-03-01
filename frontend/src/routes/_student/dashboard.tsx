import { createFileRoute } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useMyStudentProfile } from './-hooks/use-my-profile'
import { useMyEnrollments } from './-hooks/use-my-enrollments'

export const Route = createFileRoute('/_student/dashboard')({
  component: StudentDashboardPage,
})

const STATUS_VARIANT = {
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
} as const

function StudentDashboardPage() {
  const { data: profile, isLoading: profileLoading } = useMyStudentProfile()
  const { data: enrollments, isLoading: enrollLoading } = useMyEnrollments()

  if (profileLoading || enrollLoading) return <LoadingSpinner />

  const enrolled = enrollments?.filter((e) => e.status === 'approved' || e.status === 'completed') ?? []
  const pending = enrollments?.filter((e) => e.status === 'pending') ?? []

  return (
    <div className="space-y-6">
      {/* Welcome header */}
      <div>
        <h1 className="text-2xl font-bold">Welcome back, {profile?.full_name ?? '…'}!</h1>
        <p className="text-sm text-muted-foreground">
          {profile?.student_code} · Enrollment Year {profile?.enrollment_year}
        </p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="Enrolled Subjects" value={enrolled.length} />
        <StatCard label="Pending Requests" value={pending.length} />
        <StatCard label="Status" value={profile?.status ?? '—'} />
      </div>

      {/* Recent enrollments */}
      <div>
        <h2 className="mb-3 text-base font-semibold">My Enrollments</h2>
        {enrollments?.length ? (
          <div className="space-y-2">
            {enrollments.map((e) => (
              <div
                key={e.id}
                className="flex items-center justify-between rounded-md border px-4 py-3 text-sm"
              >
                <div>
                  <span className="font-mono text-xs text-muted-foreground">
                    {e.subject_id.slice(0, 8)}…
                  </span>
                  {e.request_note && (
                    <p className="mt-0.5 text-xs text-muted-foreground">{e.request_note}</p>
                  )}
                </div>
                <Badge variant={STATUS_VARIANT[e.status] ?? 'outline'}>{e.status}</Badge>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No enrollments yet.</p>
        )}
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-lg border p-4">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  )
}
