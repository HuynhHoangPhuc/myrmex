import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Room } from '../types'

export function useRooms() {
  return useQuery({
    queryKey: ['timetable-rooms'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<{ data: Room[] }>(ENDPOINTS.timetable.rooms)
      return data.data
    },
  })
}
