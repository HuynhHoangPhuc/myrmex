import * as React from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils/cn'

interface FormFieldProps {
  label: string
  error?: string
  description?: string
  required?: boolean
  className?: string
  children?: React.ReactNode
  htmlFor?: string
}

// Wrapper that adds label, description, and error message to any form control
export function FormField({
  label,
  error,
  description,
  required,
  className,
  children,
  htmlFor,
}: FormFieldProps) {
  return (
    <div className={cn('space-y-1.5', className)}>
      <Label htmlFor={htmlFor} className={cn(required && "after:ml-0.5 after:text-destructive after:content-['*']")}>
        {label}
      </Label>
      {children}
      {description && !error && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

interface TextInputFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  description?: string
}

// Convenience wrapper: FormField + Input in one
export function TextInputField({ label, error, description, className, id, ...inputProps }: TextInputFieldProps) {
  const generatedId = React.useId()
  const inputId = id ?? generatedId
  return (
    <FormField label={label} error={error} description={description} required={inputProps.required} htmlFor={inputId}>
      <Input
        id={inputId}
        className={cn(error && 'border-destructive focus-visible:ring-destructive', className)}
        aria-invalid={Boolean(error)}
        {...inputProps}
      />
    </FormField>
  )
}
