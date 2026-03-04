import * as React from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import { Building2 } from 'lucide-react'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { authStore } from '@/lib/stores/auth-store'
import { useQueryClient } from '@tanstack/react-query'
import type { AuthResponse } from '@/lib/api/types'

const searchSchema = z.object({
  code: z.string().optional(),
})

export const Route = createFileRoute('/auth/callback')({
  validateSearch: searchSchema,
  component: OAuthCallbackPage,
})

function OAuthCallbackPage() {
  const { code } = Route.useSearch()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [error, setError] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (!code) {
      setError('No auth code received.')
      return
    }

    // Error code from backend (oauth_handler.go redirectError)
    if (code.startsWith('error:')) {
      setError(decodeURIComponent(code.slice(6)))
      return
    }

    let cancelled = false

    async function exchange() {
      try {
        const { data } = await apiClient.post<AuthResponse>(ENDPOINTS.auth.oauthExchange, { code })
        if (cancelled) return

        authStore.setTokens(data.access_token, data.refresh_token)

        // Fetch the current user profile to populate the auth store
        const { data: me } = await apiClient.get<AuthResponse['user']>(ENDPOINTS.auth.me)
        if (cancelled) return

        authStore.setUser(me)
        queryClient.setQueryData(['current-user'], me)

        if (me.role === 'student') {
          void navigate({ to: '/student/dashboard' })
        } else {
          void navigate({ to: '/dashboard' })
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Authentication failed.')
        }
      }
    }

    void exchange()
    return () => { cancelled = true }
  }, [code, navigate, queryClient])

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
        <div className="w-full max-w-sm space-y-4 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-destructive mx-auto">
            <Building2 className="h-6 w-6 text-destructive-foreground" />
          </div>
          <h1 className="text-xl font-semibold">Login Failed</h1>
          <p className="text-sm text-muted-foreground">{error}</p>
          <a href="/login" className="text-sm font-medium text-primary hover:underline">
            Back to login
          </a>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <div className="w-full max-w-sm space-y-4 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary mx-auto">
          <Building2 className="h-6 w-6 text-primary-foreground" />
        </div>
        <h1 className="text-xl font-semibold">Signing in...</h1>
        <p className="text-sm text-muted-foreground">Please wait while we complete your login.</p>
      </div>
    </div>
  )
}
