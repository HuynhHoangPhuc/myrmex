import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ScheduleHeatmapCell } from '../types'

export function useScheduleHeatmap(semesterId?: string) {
  return useQuery({
    queryKey: ['analytics', 'schedule-heatmap', semesterId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<ScheduleHeatmapCell[]>(ENDPOINTS.analytics.scheduleHeatmap, {
        params: semesterId ? { semester_id: semesterId } : undefined,
      })
      return data
    },
  })
}
