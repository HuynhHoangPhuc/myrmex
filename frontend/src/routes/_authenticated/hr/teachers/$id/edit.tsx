import { createFileRoute, useNavigate, Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { TeacherForm } from '@/modules/hr/components/teacher-form'
import { useTeacher } from '@/modules/hr/hooks/use-teachers'
import { useUpdateTeacher } from '@/modules/hr/hooks/use-teacher-mutations'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/hr/teachers/$id/edit')({
  component: EditTeacherPage,
})

function EditTeacherPage() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const { toast } = useToast()
  const { data: teacher, isLoading } = useTeacher(id)
  const updateMutation = useUpdateTeacher(id)

  if (isLoading) return <LoadingSpinner />
  if (!teacher) return <p className="text-muted-foreground">Teacher not found.</p>

  return (
    <div className="max-w-2xl">
      <PageHeader
        title="Edit Teacher"
        description={teacher.full_name}
        actions={
          <Button variant="outline" asChild>
            <Link to="/hr/teachers/$id" params={{ id }}>
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
            </Link>
          </Button>
        }
      />
      <TeacherForm
        defaultValues={teacher}
        isLoading={updateMutation.isPending}
        submitLabel="Save Changes"
        onSubmit={(data) => {
          updateMutation.mutate(data, {
            onSuccess: () => {
              toast({ title: 'Teacher updated' })
              void navigate({ to: '/hr/teachers/$id', params: { id } })
            },
            onError: () => {
              toast({ title: 'Failed to update teacher', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
