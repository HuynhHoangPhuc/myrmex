import { QueryClient } from '@tanstack/react-query'

// Global TanStack Query client with sensible defaults for ERP data
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Stale after 30s; refetch on window focus for fresh ERP data
      staleTime: 30_000,
      gcTime: 5 * 60_000,
      retry: (failureCount, error: unknown) => {
        // Don't retry 4xx errors
        const status = (error as { response?: { status?: number } })?.response?.status
        if (status && status >= 400 && status < 500) return false
        return failureCount < 2
      },
    },
    mutations: {
      retry: false,
    },
  },
})
