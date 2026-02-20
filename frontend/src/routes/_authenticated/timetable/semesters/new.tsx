import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { SemesterForm } from '@/modules/timetable/components/semester-form'
import { useCreateSemester } from '@/modules/timetable/hooks/use-semesters'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/timetable/semesters/new')({
  component: NewSemesterPage,
})

function NewSemesterPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const createMutation = useCreateSemester()

  return (
    <div className="max-w-3xl">
      <PageHeader
        title="New Semester"
        description="Define semester dates, weekly time slots, and available rooms."
      />
      <SemesterForm
        isLoading={createMutation.isPending}
        onSubmit={(data) => {
          createMutation.mutate(data, {
            onSuccess: (semester) => {
              toast({ title: 'Semester created', description: semester.name })
              void navigate({ to: '/timetable/semesters/$id', params: { id: semester.id } })
            },
            onError: () => {
              toast({ title: 'Failed to create semester', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
