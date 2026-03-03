import { useMutation } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { InviteCodeResponse } from '@/modules/student/types'

export function useGenerateInviteCode() {
  return useMutation({
    mutationFn: async (studentId: string) => {
      const { data } = await apiClient.post<InviteCodeResponse>(
        ENDPOINTS.students.inviteCode(studentId),
      )
      return data
    },
  })
}
