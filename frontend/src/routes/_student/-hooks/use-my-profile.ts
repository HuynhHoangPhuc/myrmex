import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Student } from '@/modules/student/types'

export function useMyStudentProfile() {
  return useQuery({
    queryKey: ['student-portal', 'me'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Student>(ENDPOINTS.studentPortal.me)
      return data
    },
  })
}
