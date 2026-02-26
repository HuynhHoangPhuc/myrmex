import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { DashboardSummary } from '../types'

export function useDashboardSummary() {
  return useQuery({
    queryKey: ['analytics', 'dashboard'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<DashboardSummary>(ENDPOINTS.analytics.dashboard)
      return data
    },
  })
}
