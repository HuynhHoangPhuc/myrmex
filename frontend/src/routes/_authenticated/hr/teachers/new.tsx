import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { TeacherForm } from '@/modules/hr/components/teacher-form'
import { useCreateTeacher } from '@/modules/hr/hooks/use-teacher-mutations'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/hr/teachers/new')({
  component: NewTeacherPage,
})

function NewTeacherPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const createMutation = useCreateTeacher()

  return (
    <div className="max-w-2xl">
      <PageHeader
        title="Add Teacher"
        description="Create a new teacher record."
      />
      <TeacherForm
        isLoading={createMutation.isPending}
        submitLabel="Create Teacher"
        onSubmit={(data) => {
          createMutation.mutate(data, {
            onSuccess: (teacher) => {
              toast({ title: 'Teacher created', description: teacher.full_name })
              void navigate({ to: '/hr/teachers/$id', params: { id: teacher.id } })
            },
            onError: () => {
              toast({ title: 'Failed to create teacher', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
