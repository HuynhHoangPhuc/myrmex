import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { TextInputField } from '@/components/shared/form-field'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import type { CreateSemesterInput } from '../types'

const semesterSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  year: z.number().min(2000, 'Enter a valid year').max(2100),
  term: z.number().min(1).max(3),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string().min(1, 'End date is required'),
})

interface SemesterFormProps {
  onSubmit: (data: CreateSemesterInput) => void
  isLoading?: boolean
}

// Append T00:00:00Z so the backend can parse it as RFC3339
function toRFC3339(dateStr: string): string {
  return dateStr ? `${dateStr}T00:00:00Z` : ''
}

export function SemesterForm({ onSubmit, isLoading }: SemesterFormProps) {
  const form = useForm({
    defaultValues: {
      name: '',
      year: new Date().getFullYear(),
      term: 1,
      start_date: '',
      end_date: '',
    },
    onSubmit: ({ value }) => {
      onSubmit({
        name: value.name,
        year: value.year,
        term: value.term,
        start_date: toRFC3339(value.start_date),
        end_date: toRFC3339(value.end_date),
      })
    },
  })

  return (
    <form onSubmit={(e) => { e.preventDefault(); void form.handleSubmit() }} className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <form.Field name="name" validators={{ onChange: semesterSchema.shape.name }}
          children={(field) => (
            <TextInputField label="Semester Name" required placeholder="e.g. Spring 2026"
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />

        <form.Field name="year" validators={{ onChange: semesterSchema.shape.year }}
          children={(field) => (
            <TextInputField label="Year" type="number" required placeholder="e.g. 2026"
              value={String(field.state.value)} onChange={(e) => field.handleChange(Number(e.target.value))}
              error={field.state.meta.errors[0]?.toString()} />
          )} />

        <form.Field name="term"
          children={(field) => (
            <div className="space-y-1">
              <label className="text-sm font-medium">Term <span className="text-destructive">*</span></label>
              <select
                value={field.state.value}
                onChange={(e) => field.handleChange(Number(e.target.value))}
                className="h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm"
              >
                <option value={1}>Term 1</option>
                <option value={2}>Term 2</option>
                <option value={3}>Term 3</option>
              </select>
            </div>
          )} />

        <form.Field name="start_date" validators={{ onChange: semesterSchema.shape.start_date }}
          children={(field) => (
            <TextInputField label="Start Date" type="date" required
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />

        <form.Field name="end_date" validators={{ onChange: semesterSchema.shape.end_date }}
          children={(field) => (
            <TextInputField label="End Date" type="date" required
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />
      </div>

      <div className="flex justify-end pt-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
          Create Semester
        </Button>
      </div>
    </form>
  )
}
