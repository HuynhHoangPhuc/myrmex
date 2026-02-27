import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type {
  Subject,
  CreateSubjectInput,
  UpdateSubjectInput,
  AddPrerequisiteInput,
  Prerequisite,
  FullDAGResponse,
  CheckConflictsResponse,
} from '../types'

interface SubjectListParams {
  page: number
  pageSize: number
  search?: string
}

export const subjectsQueryOptions = (params: SubjectListParams) => ({
  queryKey: ['subjects', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Subject>>(ENDPOINTS.subjects.list, {
      params: { page: params.page, page_size: params.pageSize, search: params.search },
    })
    return data
  },
})

// All subjects without pagination — used for prerequisite selects
export const allSubjectsQueryOptions = () => ({
  queryKey: ['subjects', 'all'] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Subject>>(ENDPOINTS.subjects.list, {
      params: { page: 1, page_size: 500 },
    })
    return data.data
  },
})

export function useSubjects(params: SubjectListParams) {
  return useQuery(subjectsQueryOptions(params))
}

export function useAllSubjects() {
  return useQuery(allSubjectsQueryOptions())
}

export function useSubject(id: string) {
  return useQuery({
    queryKey: ['subjects', id] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Subject>(ENDPOINTS.subjects.detail(id))
      return data
    },
    enabled: Boolean(id),
  })
}

export function useSubjectPrerequisites(id: string) {
  return useQuery({
    queryKey: ['subjects', id, 'prerequisites'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Prerequisite[]>(ENDPOINTS.subjects.prerequisites(id))
      return data
    },
    enabled: Boolean(id),
  })
}

export function useCreateSubject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateSubjectInput) => {
      const { data } = await apiClient.post<Subject>(ENDPOINTS.subjects.list, input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['subjects'] })
    },
  })
}

export function useUpdateSubject(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateSubjectInput) => {
      const { data } = await apiClient.patch<Subject>(ENDPOINTS.subjects.detail(id), input)
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['subjects'] })
    },
  })
}

export function useDeleteSubject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(ENDPOINTS.subjects.detail(id))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['subjects'] })
    },
  })
}

export function useAddPrerequisite(subjectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: AddPrerequisiteInput) => {
      await apiClient.post(ENDPOINTS.subjects.prerequisites(subjectId), input)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['subjects', subjectId] })
    },
  })
}

export function useFullDAG() {
  return useQuery({
    queryKey: ['subjects', 'dag', 'full'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<FullDAGResponse>(ENDPOINTS.subjects.dag.full)
      return data
    },
    staleTime: 60_000, // DAG changes infrequently
  })
}

export function useCheckConflicts(subjectIds: string[]) {
  return useQuery({
    queryKey: ['subjects', 'dag', 'check-conflicts', subjectIds.slice().sort().join(',')] as const,
    queryFn: async () => {
      const { data } = await apiClient.post<CheckConflictsResponse>(
        ENDPOINTS.subjects.dag.checkConflicts,
        { subject_ids: subjectIds }
      )
      return data
    },
    enabled: subjectIds.length > 1,
    staleTime: 30_000,
  })
}

export function useRemovePrerequisite(subjectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (prerequisiteId: string) => {
      await apiClient.delete(`${ENDPOINTS.subjects.prerequisites(subjectId)}/${prerequisiteId}`)
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['subjects', subjectId] })
    },
  })
}
