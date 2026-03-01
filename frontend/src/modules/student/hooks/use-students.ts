import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Student } from '../types'

interface StudentListParams {
  page: number
  pageSize: number
  departmentId?: string
  status?: string
}

export const studentsQueryOptions = (params: StudentListParams) => ({
  queryKey: ['students', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Student>>(ENDPOINTS.students.list, {
      params: {
        page: params.page,
        page_size: params.pageSize,
        department_id: params.departmentId,
        status: params.status,
      },
    })
    return data
  },
})

export const studentDetailQueryOptions = (id: string) => ({
  queryKey: ['students', id] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<Student>(ENDPOINTS.students.detail(id))
    return data
  },
  enabled: Boolean(id),
})

export function useStudents(params: StudentListParams) {
  return useQuery(studentsQueryOptions(params))
}

export function useStudent(id: string) {
  return useQuery(studentDetailQueryOptions(id))
}
