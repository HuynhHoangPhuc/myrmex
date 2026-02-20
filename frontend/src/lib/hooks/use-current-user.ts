import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { authStore } from '@/lib/stores/auth-store'
import type { User } from '@/lib/api/types'

// Fetches current user from API; falls back to localStorage cache if available
export function useCurrentUser() {
  return useQuery({
    queryKey: ['current-user'],
    queryFn: async (): Promise<User> => {
      const { data } = await apiClient.get<User>(ENDPOINTS.auth.me)
      authStore.setUser(data)
      return data
    },
    // Seed initial data from localStorage to avoid flash of unauthenticated state
    initialData: () => authStore.getUser() ?? undefined,
    enabled: authStore.isAuthenticated(),
    staleTime: 5 * 60_000,
  })
}
