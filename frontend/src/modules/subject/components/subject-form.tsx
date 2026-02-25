import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { TextInputField, FormField } from '@/components/shared/form-field'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useAllDepartments } from '@/modules/hr/hooks/use-departments'
import type { Subject, CreateSubjectInput } from '../types'

const subjectSchema = z.object({
  code: z.string().min(1, 'Code is required'),
  name: z.string().min(2, 'Name is required'),
  credits: z.number().min(1).max(10),
  description: z.string().optional(),
  department_id: z.string().min(1, 'Department is required'),
  weekly_hours: z.number().min(1).max(20),
})

interface SubjectFormProps {
  defaultValues?: Partial<Subject>
  onSubmit: (data: CreateSubjectInput) => void
  isLoading?: boolean
  submitLabel?: string
}

export function SubjectForm({ defaultValues, onSubmit, isLoading, submitLabel = 'Save' }: SubjectFormProps) {
  const { data: departments = [] } = useAllDepartments()

  const form = useForm({
    defaultValues: {
      code: defaultValues?.code ?? '',
      name: defaultValues?.name ?? '',
      credits: defaultValues?.credits ?? 3,
      description: defaultValues?.description ?? '',
      department_id: defaultValues?.department_id ?? '',
      weekly_hours: defaultValues?.weekly_hours ?? 3,
    },
    onSubmit: ({ value }) => onSubmit(value as CreateSubjectInput),
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        void form.handleSubmit()
      }}
      className="space-y-4"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <form.Field
          name="code"
          validators={{ onChange: subjectSchema.shape.code }}
          children={(field) => (
            <TextInputField
              label="Subject Code"
              required
              placeholder="e.g. CS101"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value.toUpperCase())}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="name"
          validators={{ onChange: subjectSchema.shape.name }}
          children={(field) => (
            <TextInputField
              label="Subject Name"
              required
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="credits"
          validators={{ onChange: subjectSchema.shape.credits }}
          children={(field) => (
            <TextInputField
              label="Credits"
              type="number"
              min={1}
              max={10}
              required
              value={String(field.state.value)}
              onChange={(e) => field.handleChange(Number(e.target.value))}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="weekly_hours"
          validators={{ onChange: subjectSchema.shape.weekly_hours }}
          children={(field) => (
            <TextInputField
              label="Weekly Hours"
              type="number"
              min={1}
              max={20}
              required
              value={String(field.state.value)}
              onChange={(e) => field.handleChange(Number(e.target.value))}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="department_id"
          validators={{ onChange: subjectSchema.shape.department_id }}
          children={(field) => (
            <FormField label="Department" required error={field.state.meta.errors[0]?.toString()}>
              <select
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              >
                <option value="">Select department…</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </select>
            </FormField>
          )}
        />
      </div>

      <form.Field
        name="description"
        children={(field) => (
          <FormField label="Description">
            <textarea
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              rows={3}
              className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm resize-none"
              placeholder="Optional description…"
            />
          </FormField>
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
