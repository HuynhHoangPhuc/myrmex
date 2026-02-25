import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { DepartmentForm } from '@/modules/hr/components/department-form'
import { useCreateDepartment } from '@/modules/hr/hooks/use-departments'
import { useToast } from '@/lib/hooks/use-toast'

export const Route = createFileRoute('/_authenticated/hr/departments/new')({
  component: NewDepartmentPage,
})

function NewDepartmentPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const createMutation = useCreateDepartment()

  return (
    <div className="max-w-lg">
      <PageHeader title="Add Department" description="Create a new faculty department." />
      <DepartmentForm
        isLoading={createMutation.isPending}
        submitLabel="Create Department"
        onSubmit={(data) => {
          createMutation.mutate(data, {
            onSuccess: () => {
              toast({ title: 'Department created' })
              void navigate({ to: '/hr/departments', search: { page: 1, pageSize: 25 } })
            },
            onError: () => {
              toast({ title: 'Failed to create department', variant: 'destructive' })
            },
          })
        }}
      />
    </div>
  )
}
