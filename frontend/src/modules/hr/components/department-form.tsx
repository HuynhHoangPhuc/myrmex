import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { TextInputField } from '@/components/shared/form-field'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import type { Department, CreateDepartmentInput } from '../types'

const departmentSchema = z.object({
  name: z.string().min(2, 'Name is required'),
  code: z.string().min(1, 'Code is required').max(10, 'Max 10 characters'),
  description: z.string().optional(),
})

interface DepartmentFormProps {
  defaultValues?: Partial<Department>
  onSubmit: (data: CreateDepartmentInput) => void
  isLoading?: boolean
  submitLabel?: string
}

export function DepartmentForm({ defaultValues, onSubmit, isLoading, submitLabel = 'Save' }: DepartmentFormProps) {
  const form = useForm({
    defaultValues: {
      name: defaultValues?.name ?? '',
      code: defaultValues?.code ?? '',
      description: defaultValues?.description ?? '',
    },
    onSubmit: ({ value }) => onSubmit(value as CreateDepartmentInput),
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        void form.handleSubmit()
      }}
      className="space-y-4"
    >
      <form.Field
        name="name"
        validators={{ onChange: departmentSchema.shape.name }}
        children={(field) => (
          <TextInputField
            label="Department Name"
            required
            value={field.state.value}
            onChange={(e) => field.handleChange(e.target.value)}
            error={field.state.meta.errors[0]?.toString()}
          />
        )}
      />
      <form.Field
        name="code"
        validators={{ onChange: departmentSchema.shape.code }}
        children={(field) => (
          <TextInputField
            label="Code"
            required
            placeholder="e.g. CS, EE"
            value={field.state.value}
            onChange={(e) => field.handleChange(e.target.value.toUpperCase())}
            error={field.state.meta.errors[0]?.toString()}
          />
        )}
      />
      <form.Field
        name="description"
        children={(field) => (
          <TextInputField
            label="Description"
            value={field.state.value}
            onChange={(e) => field.handleChange(e.target.value)}
          />
        )}
      />

      <div className="flex justify-end pt-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
          {submitLabel}
        </Button>
      </div>
    </form>
  )
}
