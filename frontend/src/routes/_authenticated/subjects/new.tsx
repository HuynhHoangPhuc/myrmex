import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { SubjectForm } from '@/modules/subject/components/subject-form'
import { useCreateSubject } from '@/modules/subject/hooks/use-subjects'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/subjects/new')({
  component: NewSubjectPage,
})

function NewSubjectPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const createMutation = useCreateSubject()

  return (
    <div className="max-w-2xl">
      <PageHeader title="Add Subject" description="Create a new course subject." />
      <SubjectForm
        isLoading={createMutation.isPending}
        submitLabel="Create Subject"
        onSubmit={(data) => {
          createMutation.mutate(data, {
            onSuccess: (subject) => {
              toast({ title: 'Subject created', description: subject.name })
              void navigate({ to: '/subjects/$id', params: { id: subject.id } })
            },
            onError: () => {
              toast({ title: 'Failed to create subject', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
