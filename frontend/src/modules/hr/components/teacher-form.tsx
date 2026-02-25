import * as React from 'react'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { X } from 'lucide-react'
import { TextInputField, FormField } from '@/components/shared/form-field'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useAllDepartments } from '../hooks/use-departments'
import type { Teacher, CreateTeacherInput } from '../types'

const teacherSchema = z.object({
  employee_code: z.string().min(1, 'Employee code is required'),
  full_name: z.string().min(2, 'Full name is required'),
  email: z.string().email('Valid email is required'),
  phone: z.string().optional(),
  department_id: z.string().min(1, 'Department is required'),
  max_hours_per_week: z.number().min(1).max(60),
  specializations: z.array(z.string()),
})

interface TeacherFormProps {
  defaultValues?: Partial<Teacher>
  onSubmit: (data: CreateTeacherInput) => void
  isLoading?: boolean
  submitLabel?: string
}

// Shared create/edit form for Teacher — Zod validated, department select
export function TeacherForm({ defaultValues, onSubmit, isLoading, submitLabel = 'Save' }: TeacherFormProps) {
  const { data: departments = [] } = useAllDepartments()
  const [specInput, setSpecInput] = React.useState('')

  const form = useForm({
    defaultValues: {
      employee_code: defaultValues?.employee_code ?? '',
      full_name: defaultValues?.full_name ?? '',
      email: defaultValues?.email ?? '',
      phone: defaultValues?.phone ?? '',
      department_id: defaultValues?.department_id ?? '',
      max_hours_per_week: defaultValues?.max_hours_per_week ?? 20,
      specializations: defaultValues?.specializations ?? [],
    },
    onSubmit: ({ value }) => onSubmit(value as CreateTeacherInput),
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        void form.handleSubmit()
      }}
      className="space-y-5"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <form.Field
          name="employee_code"
          validators={{ onChange: teacherSchema.shape.employee_code }}
          children={(field) => (
            <TextInputField
              label="Employee Code"
              required
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="full_name"
          validators={{ onChange: teacherSchema.shape.full_name }}
          children={(field) => (
            <TextInputField
              label="Full Name"
              required
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="email"
          validators={{ onChange: teacherSchema.shape.email }}
          children={(field) => (
            <TextInputField
              label="Email"
              type="email"
              required
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
        <form.Field
          name="phone"
          children={(field) => (
            <TextInputField
              label="Phone"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
            />
          )}
        />
        <form.Field
          name="department_id"
          validators={{ onChange: teacherSchema.shape.department_id }}
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
        <form.Field
          name="max_hours_per_week"
          validators={{ onChange: teacherSchema.shape.max_hours_per_week }}
          children={(field) => (
            <TextInputField
              label="Max Hours / Week"
              type="number"
              min={1}
              max={60}
              required
              value={String(field.state.value)}
              onChange={(e) => field.handleChange(Number(e.target.value))}
              error={field.state.meta.errors[0]?.toString()}
            />
          )}
        />
      </div>

      {/* Tag input for specializations */}
      <form.Field
        name="specializations"
        children={(field) => (
          <FormField label="Specializations" description="Press Enter to add a specialization">
            <div className="flex flex-wrap gap-1.5 rounded-md border border-input p-2 min-h-[40px]">
              {field.state.value.map((s) => (
                <span key={s} className="inline-flex items-center gap-1 rounded bg-secondary px-2 py-0.5 text-xs">
                  {s}
                  <button type="button" onClick={() => field.handleChange(field.state.value.filter((v) => v !== s))}>
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
              <input
                className="flex-1 bg-transparent text-sm outline-none min-w-[120px]"
                placeholder="Add specialization…"
                value={specInput}
                onChange={(e) => setSpecInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && specInput.trim()) {
                    e.preventDefault()
                    if (!field.state.value.includes(specInput.trim())) {
                      field.handleChange([...field.state.value, specInput.trim()])
                    }
                    setSpecInput('')
                  }
                }}
              />
            </div>
          </FormField>
        )}
      />

      <div className="flex justify-end gap-3 pt-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
          {submitLabel}
        </Button>
      </div>
    </form>
  )
}
