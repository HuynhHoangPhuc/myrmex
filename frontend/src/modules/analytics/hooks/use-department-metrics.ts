import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { DepartmentMetric } from '../types'

export function useDepartmentMetrics() {
  return useQuery({
    queryKey: ['analytics', 'department-metrics'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<DepartmentMetric[]>(ENDPOINTS.analytics.departmentMetrics)
      return data
    },
  })
}
