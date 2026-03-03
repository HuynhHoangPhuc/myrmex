import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { PrerequisiteCheckResult } from '@/modules/student/types'

export function useCheckPrerequisites(subjectId: string | null) {
  return useQuery({
    queryKey: ['student-portal', 'prerequisites', subjectId] as const,
    enabled: Boolean(subjectId),
    staleTime: 5 * 60 * 1000, // prereqs change rarely — avoid refetch flicker on each dialog open
    queryFn: async () => {
      const { data } = await apiClient.get<PrerequisiteCheckResult>(
        ENDPOINTS.studentPortal.checkPrerequisites,
        { params: { subject_id: subjectId } },
      )
      return data
    },
  })
}
