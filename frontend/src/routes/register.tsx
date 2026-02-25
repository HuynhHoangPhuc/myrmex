import { createFileRoute, Link, redirect } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Building2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { TextInputField } from '@/components/shared/form-field'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useRegister } from '@/lib/hooks/use-auth'
import { authStore } from '@/lib/stores/auth-store'
import { toast } from '@/lib/hooks/use-toast'

const registerSchema = z
  .object({
    full_name: z.string().min(2, 'Name must be at least 2 characters'),
    email: z.string().email('Invalid email address'),
    password: z.string().min(6, 'Password must be at least 6 characters'),
    confirm_password: z.string(),
  })
  .refine((d) => d.password === d.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  })

export const Route = createFileRoute('/register')({
  beforeLoad: () => {
    if (authStore.isAuthenticated()) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: RegisterPage,
})

function RegisterPage() {
  const register = useRegister()

  const form = useForm({
    defaultValues: { full_name: '', email: '', password: '', confirm_password: '' },
    validators: { onSubmit: registerSchema },
    onSubmit: async ({ value }) => {
      try {
        await register.mutateAsync({
          full_name: value.full_name,
          email: value.email,
          password: value.password,
        })
        toast({ title: 'Account created', description: 'Please sign in to continue.' })
      } catch {
        toast({
          variant: 'destructive',
          title: 'Registration failed',
          description: 'This email may already be in use.',
        })
      }
    },
  })

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="flex flex-col items-center gap-2 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary">
            <Building2 className="h-6 w-6 text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-bold">Create account</h1>
          <p className="text-sm text-muted-foreground">Join Myrmex ERP</p>
        </div>

        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <form
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit()
            }}
            className="space-y-4"
          >
            <form.Field name="full_name" validators={{ onBlur: registerSchema.innerType().shape.full_name }}>
              {(field) => (
                <TextInputField
                  label="Full name"
                  placeholder="Jane Smith"
                  required
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  error={field.state.meta.errors[0]?.toString()}
                />
              )}
            </form.Field>

            <form.Field name="email" validators={{ onBlur: registerSchema.innerType().shape.email }}>
              {(field) => (
                <TextInputField
                  label="Email"
                  type="email"
                  placeholder="you@university.edu"
                  required
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  error={field.state.meta.errors[0]?.toString()}
                />
              )}
            </form.Field>

            <form.Field name="password" validators={{ onBlur: registerSchema.innerType().shape.password }}>
              {(field) => (
                <TextInputField
                  label="Password"
                  type="password"
                  placeholder="••••••••"
                  required
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  error={field.state.meta.errors[0]?.toString()}
                />
              )}
            </form.Field>

            <form.Field name="confirm_password">
              {(field) => (
                <TextInputField
                  label="Confirm password"
                  type="password"
                  placeholder="••••••••"
                  required
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  error={field.state.meta.errors[0]?.toString()}
                />
              )}
            </form.Field>

            <Button type="submit" className="w-full" disabled={register.isPending}>
              {register.isPending && <LoadingSpinner size="sm" className="mr-2" />}
              Create account
            </Button>
          </form>
        </div>

        <p className="text-center text-sm text-muted-foreground">
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-primary hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  )
}
