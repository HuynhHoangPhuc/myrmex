import { createFileRoute } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useMyStudentProfile } from './-hooks/use-my-profile'

export const Route = createFileRoute('/_student/profile')({
  component: StudentProfilePage,
})

function StudentProfilePage() {
  const { data: profile, isLoading } = useMyStudentProfile()

  if (isLoading) return <LoadingSpinner />
  if (!profile) return <p className="text-muted-foreground">Profile not found.</p>

  return (
    <div className="max-w-lg space-y-6">
      <h1 className="text-2xl font-bold">My Profile</h1>

      <div className="grid gap-4 rounded-lg border p-5 sm:grid-cols-2">
        <InfoRow label="Full Name" value={profile.full_name} />
        <InfoRow label="Student Code" value={profile.student_code} />
        <InfoRow label="Email" value={profile.email} />
        <InfoRow label="Enrollment Year" value={String(profile.enrollment_year)} />
        <InfoRow label="Status">
          <Badge>{profile.status}</Badge>
        </InfoRow>
        <InfoRow label="Account" value={profile.user_id ? 'Activated' : 'Pending activation'} />
      </div>

      <p className="text-xs text-muted-foreground">
        To update your personal information, please contact your faculty administrator.
      </p>
    </div>
  )
}

function InfoRow({
  label,
  value,
  children,
}: {
  label: string
  value?: string
  children?: React.ReactNode
}) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-sm font-medium">{children ?? value}</p>
    </div>
  )
}
