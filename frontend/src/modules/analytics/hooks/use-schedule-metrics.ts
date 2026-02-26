import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ScheduleMetric } from '../types'

export function useScheduleMetrics(semesterId?: string) {
  return useQuery({
    queryKey: ['analytics', 'schedule-metrics', semesterId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<ScheduleMetric[]>(ENDPOINTS.analytics.scheduleMetrics, {
        params: semesterId ? { semester_id: semesterId } : undefined,
      })
      return data
    },
  })
}
