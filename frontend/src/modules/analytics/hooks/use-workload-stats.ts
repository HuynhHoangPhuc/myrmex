import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { WorkloadStat } from '../types'

export function useWorkloadStats(semesterId?: string) {
  return useQuery({
    queryKey: ['analytics', 'workload', semesterId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<WorkloadStat[]>(ENDPOINTS.analytics.workload, {
        params: semesterId ? { semester_id: semesterId } : undefined,
      })
      return data
    },
  })
}
