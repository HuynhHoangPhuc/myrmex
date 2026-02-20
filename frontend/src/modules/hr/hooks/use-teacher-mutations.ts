import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Teacher, CreateTeacherInput, UpdateTeacherInput, TeacherAvailability } from '../types'

// Create a new teacher, invalidates teacher list
export function useCreateTeacher() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateTeacherInput) => {
      const { data } = await apiClient.post<Teacher>(ENDPOINTS.hr.teachers, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['teachers'] })
    },
  })
}

// Update teacher fields by id
export function useUpdateTeacher(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateTeacherInput) => {
      const { data } = await apiClient.patch<Teacher>(ENDPOINTS.hr.teacher(id), input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['teachers'] })
    },
  })
}

// Delete teacher by id
export function useDeleteTeacher() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(ENDPOINTS.hr.teacher(id))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['teachers'] })
    },
  })
}

// Update teacher weekly availability slots
export function useUpdateAvailability(teacherId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (slots: TeacherAvailability[]) => {
      const { data } = await apiClient.put<Teacher>(
        `${ENDPOINTS.hr.teacher(teacherId)}/availability`,
        { availability: slots },
      )
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['teachers', teacherId] })
    },
  })
}
