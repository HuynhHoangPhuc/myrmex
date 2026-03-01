import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { EnrollmentRequest, ReviewEnrollmentInput } from '../types'

interface EnrollmentListParams {
  page: number
  pageSize: number
  studentId?: string
  semesterId?: string
  status?: string
}

export function useEnrollments(params: EnrollmentListParams) {
  return useQuery({
    queryKey: ['enrollments', params] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<ListResponse<EnrollmentRequest>>(ENDPOINTS.enrollments.list, {
        params: {
          page: params.page,
          page_size: params.pageSize,
          student_id: params.studentId,
          semester_id: params.semesterId,
          status: params.status,
        },
      })
      return data
    },
  })
}

export function useReviewEnrollment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: ReviewEnrollmentInput }) => {
      const { data } = await apiClient.patch<EnrollmentRequest>(
        ENDPOINTS.enrollments.review(id),
        input,
      )
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['enrollments'] })
    },
  })
}
