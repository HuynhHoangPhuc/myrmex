import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Teacher } from '../types'

interface TeacherListParams {
  page: number
  pageSize: number
  search?: string
}

// Query options factory — reusable for prefetching
export const teachersQueryOptions = (params: TeacherListParams) => ({
  queryKey: ['teachers', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Teacher>>(ENDPOINTS.hr.teachers, {
      params: { page: params.page, page_size: params.pageSize, search: params.search },
    })
    return data
  },
})

export const teacherDetailQueryOptions = (id: string) => ({
  queryKey: ['teachers', id] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<Teacher>(ENDPOINTS.hr.teacher(id))
    return data
  },
  enabled: Boolean(id),
})

// Paginated teacher list
export function useTeachers(params: TeacherListParams) {
  return useQuery(teachersQueryOptions(params))
}

// Single teacher detail
export function useTeacher(id: string) {
  return useQuery(teacherDetailQueryOptions(id))
}
