import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { EnrollmentRequest, RequestEnrollmentInput } from '@/modules/student/types'

export function useMyEnrollments() {
  return useQuery({
    queryKey: ['student-portal', 'enrollments'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<EnrollmentRequest[]>(ENDPOINTS.studentPortal.myEnrollments)
      return data
    },
  })
}

export function useRequestEnrollment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: RequestEnrollmentInput) => {
      const { data } = await apiClient.post<EnrollmentRequest>(
        ENDPOINTS.studentPortal.requestEnrollment,
        input,
      )
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['student-portal', 'enrollments'] })
    },
  })
}
