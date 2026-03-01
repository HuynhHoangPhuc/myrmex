import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Student, CreateStudentInput, UpdateStudentInput, CreateStudentResponse } from '../types'

export function useCreateStudent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateStudentInput) => {
      const { data } = await apiClient.post<CreateStudentResponse>(ENDPOINTS.students.list, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['students'] })
    },
  })
}

export function useUpdateStudent(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateStudentInput) => {
      const { data } = await apiClient.patch<Student>(ENDPOINTS.students.detail(id), input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['students'] })
    },
  })
}

export function useDeleteStudent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(ENDPOINTS.students.detail(id))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['students'] })
    },
  })
}
