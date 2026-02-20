import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { authStore } from '@/lib/stores/auth-store'
import type { AuthResponse } from '@/lib/api/types'

interface LoginCredentials {
  email: string
  password: string
}

interface RegisterCredentials {
  full_name: string
  email: string
  password: string
}

// Login mutation: posts credentials, stores JWT + user, navigates to dashboard
export function useLogin() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  return useMutation({
    mutationFn: async (creds: LoginCredentials) => {
      const { data } = await apiClient.post<AuthResponse>(ENDPOINTS.auth.login, creds)
      return data
    },
    onSuccess: (data) => {
      authStore.setTokens(data.access_token, data.refresh_token)
      authStore.setUser(data.user)
      queryClient.setQueryData(['current-user'], data.user)
      void navigate({ to: '/dashboard' })
    },
  })
}

// Register mutation: creates account then redirects to login
export function useRegister() {
  const navigate = useNavigate()

  return useMutation({
    mutationFn: async (creds: RegisterCredentials) => {
      const { data } = await apiClient.post<AuthResponse>(ENDPOINTS.auth.register, creds)
      return data
    },
    onSuccess: () => {
      void navigate({ to: '/login' })
    },
  })
}

// Logout: clears store, invalidates all queries, navigates to login
export function useLogout() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  return () => {
    authStore.clear()
    queryClient.clear()
    void navigate({ to: '/login' })
  }
}
