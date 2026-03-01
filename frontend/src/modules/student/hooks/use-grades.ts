import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Grade, AssignGradeInput, UpdateGradeInput } from '../types'

// Derive letter grade client-side (mirrors the DB generated column logic)
export function getLetterGrade(n: number): string {
  if (n >= 8.5) return 'A'
  if (n >= 7.0) return 'B'
  if (n >= 5.5) return 'C'
  if (n >= 4.0) return 'D'
  return 'F'
}

export function useAssignGrade() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: AssignGradeInput) => {
      const { data } = await apiClient.post<Grade>(ENDPOINTS.grades.assign, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['enrollments'] })
      void qc.invalidateQueries({ queryKey: ['grades'] })
    },
  })
}

export function useUpdateGrade(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateGradeInput) => {
      const { data } = await apiClient.patch<Grade>(ENDPOINTS.grades.update(id), input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['enrollments'] })
      void qc.invalidateQueries({ queryKey: ['grades'] })
    },
  })
}
