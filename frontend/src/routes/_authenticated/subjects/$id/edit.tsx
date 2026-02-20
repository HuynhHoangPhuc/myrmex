import { createFileRoute, useNavigate, Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/shared/page-header'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { SubjectForm } from '@/modules/subject/components/subject-form'
import { useSubject, useUpdateSubject } from '@/modules/subject/hooks/use-subjects'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/subjects/$id/edit')({
  component: EditSubjectPage,
})

function EditSubjectPage() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const { toast } = useToast()
  const { data: subject, isLoading } = useSubject(id)
  const updateMutation = useUpdateSubject(id)

  if (isLoading) return <LoadingSpinner />
  if (!subject) return <p className="text-muted-foreground">Subject not found.</p>

  return (
    <div className="max-w-2xl">
      <PageHeader
        title="Edit Subject"
        description={subject.name}
        actions={
          <Button variant="outline" asChild>
            <Link to="/subjects/$id" params={{ id }}>
              <ArrowLeft className="mr-2 h-4 w-4" />Back
            </Link>
          </Button>
        }
      />
      <SubjectForm
        defaultValues={subject}
        isLoading={updateMutation.isPending}
        submitLabel="Save Changes"
        onSubmit={(data) => {
          updateMutation.mutate(data, {
            onSuccess: () => {
              toast({ title: 'Subject updated' })
              void navigate({ to: '/subjects/$id', params: { id } })
            },
            onError: () => {
              toast({ title: 'Failed to update subject', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
