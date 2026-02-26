import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { UtilizationStat } from '../types'

export function useUtilization(semesterId?: string) {
  return useQuery({
    queryKey: ['analytics', 'utilization', semesterId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<UtilizationStat[]>(ENDPOINTS.analytics.utilization, {
        params: semesterId ? { semester_id: semesterId } : undefined,
      })
      return data
    },
  })
}
