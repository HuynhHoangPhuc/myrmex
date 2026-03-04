import * as React from 'react'
import { createFileRoute, Link, redirect } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Building2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { TextInputField } from '@/components/shared/form-field'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useLogin } from '@/lib/hooks/use-auth'
import { authStore } from '@/lib/stores/auth-store'
import { toast } from '@/lib/hooks/use-toast'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { apiClient } from '@/lib/api/client'

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

// Redirect already-authenticated users away from login page
export const Route = createFileRoute('/login')({
  beforeLoad: () => {
    if (authStore.isAuthenticated()) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: LoginPage,
})

// Detects the OAuth provider from the email domain.
// Returns "google" | "microsoft" | null (null = use password login).
function detectOAuthProvider(email: string): 'google' | 'microsoft' | null {
  const domain = email.split('@')[1]
  if (domain === 'hcmus.edu.vn') return 'google'
  if (domain === 'student.hcmus.edu.vn') return 'microsoft'
  return null
}

function OAuthButton({
  provider,
  label,
  hint,
}: {
  provider: 'google' | 'microsoft'
  label: string
  hint: string
}) {
  const [loading, setLoading] = React.useState(false)

  async function handleClick() {
    setLoading(true)
    try {
      // The backend endpoint performs a redirect to the provider.
      // We get the redirect URL and follow it by navigating the browser.
      const endpoint =
        provider === 'google'
          ? ENDPOINTS.auth.oauthGoogleLogin
          : ENDPOINTS.auth.oauthMicrosoftLogin
      // apiClient base URL + path — navigate browser directly to trigger redirect
      const baseURL = (apiClient.defaults.baseURL ?? '').replace(/\/$/, '')
      window.location.href = baseURL + endpoint
    } catch {
      setLoading(false)
    }
  }

  return (
    <Button
      type="button"
      variant="outline"
      className="w-full"
      onClick={handleClick}
      disabled={loading}
    >
      {loading && <LoadingSpinner size="sm" className="mr-2" />}
      {label}
      <span className="ml-1 text-xs text-muted-foreground">({hint})</span>
    </Button>
  )
}

function LoginPage() {
  const login = useLogin()
  const [emailValue, setEmailValue] = React.useState('')

  const detectedProvider = emailValue.includes('@') ? detectOAuthProvider(emailValue) : null

  const form = useForm({
    defaultValues: { email: '', password: '' },
    validators: { onSubmit: loginSchema },
    onSubmit: async ({ value }) => {
      try {
        await login.mutateAsync(value)
      } catch {
        toast({
          variant: 'destructive',
          title: 'Login failed',
          description: 'Invalid email or password. Please try again.',
        })
      }
    },
  })

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <div className="w-full max-w-sm space-y-6">
        {/* Logo */}
        <div className="flex flex-col items-center gap-2 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary">
            <Building2 className="h-6 w-6 text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-bold">Myrmex ERP</h1>
          <p className="text-sm text-muted-foreground">Sign in to your account</p>
        </div>

        {/* OAuth buttons */}
        <div className="rounded-xl border bg-card p-6 shadow-sm space-y-3">
          <p className="text-xs font-medium text-muted-foreground text-center uppercase tracking-wide">
            Institutional Login
          </p>
          <OAuthButton
            provider="google"
            label="Continue with Google"
            hint="@hcmus.edu.vn"
          />
          <OAuthButton
            provider="microsoft"
            label="Continue with Microsoft"
            hint="@student.hcmus.edu.vn"
          />
        </div>

        {/* Divider */}
        <div className="flex items-center gap-3">
          <div className="h-px flex-1 bg-border" />
          <span className="text-xs text-muted-foreground">or sign in with password</span>
          <div className="h-px flex-1 bg-border" />
        </div>

        {/* Password Form */}
        <div className="rounded-xl border bg-card p-6 shadow-sm">
          {detectedProvider && (
            <div className="mb-4 rounded-md bg-blue-50 dark:bg-blue-950/30 px-3 py-2 text-xs text-blue-700 dark:text-blue-300">
              💡 Tip: use{' '}
              <strong>
                {detectedProvider === 'google' ? 'Google' : 'Microsoft'} login
              </strong>{' '}
              for your {emailValue.split('@')[1]} account.
            </div>
          )}
          <form
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit()
            }}
            className="space-y-4"
          >
            <form.Field
              name="email"
              validators={{ onBlur: loginSchema.shape.email }}
            >
              {(field) => (
                <TextInputField
                  label="Email"
                  type="email"
                  placeholder="you@university.edu"
                  required
                  value={field.state.value}
                  onChange={(e) => {
                    field.handleChange(e.target.value)
                    setEmailValue(e.target.value)
                  }}
                  onBlur={field.handleBlur}
                  error={field.state.meta.errors[0]?.toString()}
                />
              )}
            </form.Field>

            <form.Field
              name="password"
              validators={{ onBlur: loginSchema.shape.password }}
            >
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

            <Button type="submit" className="w-full" disabled={login.isPending}>
              {login.isPending && <LoadingSpinner size="sm" className="mr-2" />}
              Sign in
            </Button>
          </form>
        </div>

        <p className="text-center text-sm text-muted-foreground">
          Don&apos;t have an account?{' '}
          <Link to="/register" className="font-medium text-primary hover:underline">
            Register
          </Link>
        </p>
      </div>
    </div>
  )
}
