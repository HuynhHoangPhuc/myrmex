import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Semester, CreateSemesterInput } from '../types'

interface SemesterListParams {
  page: number
  pageSize: number
}

export const semestersQueryOptions = (params: SemesterListParams) => ({
  queryKey: ['semesters', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Semester>>(ENDPOINTS.timetable.semesters, {
      params: { page: params.page, page_size: params.pageSize },
    })
    return data
  },
})

export const semesterDetailQueryOptions = (id: string) => ({
  queryKey: ['semesters', id] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<Semester>(ENDPOINTS.timetable.semester(id))
    return data
  },
  enabled: Boolean(id),
})

export function useSemesters(params: SemesterListParams) {
  return useQuery(semestersQueryOptions(params))
}

export function useSemester(id: string) {
  return useQuery(semesterDetailQueryOptions(id))
}

// All semesters without pagination — used for dropdowns
export function useAllSemesters() {
  return useQuery({
    queryKey: ['semesters', 'all'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<ListResponse<Semester>>(ENDPOINTS.timetable.semesters, {
        params: { page: 1, page_size: 100 },
      })
      return data.data
    },
  })
}

export function useCreateSemester() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateSemesterInput) => {
      const { data } = await apiClient.post<Semester>(ENDPOINTS.timetable.semesters, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['semesters'] })
    },
  })
}

export function useDeleteSemester() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(ENDPOINTS.timetable.semester(id))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['semesters'] })
    },
  })
}
