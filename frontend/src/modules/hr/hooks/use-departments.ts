import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Department, CreateDepartmentInput, UpdateDepartmentInput } from '../types'

interface DepartmentListParams {
  page: number
  pageSize: number
  search?: string
}

export const departmentsQueryOptions = (params: DepartmentListParams) => ({
  queryKey: ['departments', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Department>>(ENDPOINTS.hr.departments, {
      params: { page: params.page, page_size: params.pageSize, search: params.search },
    })
    return data
  },
})

// All departments (no pagination) — used in select dropdowns
export const allDepartmentsQueryOptions = () => ({
  queryKey: ['departments', 'all'] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Department>>(ENDPOINTS.hr.departments, {
      params: { page: 1, page_size: 200 },
    })
    return data.data
  },
})

export function useDepartments(params: DepartmentListParams) {
  return useQuery(departmentsQueryOptions(params))
}

export function useAllDepartments() {
  return useQuery(allDepartmentsQueryOptions())
}

export function useDepartment(id: string) {
  return useQuery({
    queryKey: ['departments', id] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Department>(ENDPOINTS.hr.department(id))
      return data
    },
    enabled: Boolean(id),
  })
}

export function useCreateDepartment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateDepartmentInput) => {
      const { data } = await apiClient.post<Department>(ENDPOINTS.hr.departments, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['departments'] })
    },
  })
}

export function useUpdateDepartment(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateDepartmentInput) => {
      const { data } = await apiClient.patch<Department>(ENDPOINTS.hr.department(id), input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['departments'] })
    },
  })
}

export function useDeleteDepartment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(ENDPOINTS.hr.department(id))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['departments'] })
    },
  })
}
