import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Schedule, GenerateScheduleInput, TeacherSuggestion, AssignTeacherInput } from '../types'

interface ScheduleListParams {
  semesterId?: string
  page: number
  pageSize: number
}

export const schedulesQueryOptions = (params: ScheduleListParams) => ({
  queryKey: ['schedules', params] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<ListResponse<Schedule>>('/timetable/schedules', {
      params: {
        page: params.page,
        page_size: params.pageSize,
        semester_id: params.semesterId,
      },
    })
    return data
  },
})

export const scheduleDetailQueryOptions = (id: string) => ({
  queryKey: ['schedules', id] as const,
  queryFn: async () => {
    const { data } = await apiClient.get<Schedule>(`/timetable/schedules/${id}`)
    return data
  },
  enabled: Boolean(id),
})

export function useSchedules(params: ScheduleListParams) {
  return useQuery(schedulesQueryOptions(params))
}

export function useSchedule(id: string) {
  return useQuery(scheduleDetailQueryOptions(id))
}

// Poll generation status after triggering — enabled only when jobId present
export function useGenerationStatus(scheduleId: string | null) {
  return useQuery({
    queryKey: ['schedules', scheduleId, 'status'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Schedule>(`/timetable/schedules/${scheduleId}`)
      return data
    },
    enabled: Boolean(scheduleId),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      // Stop polling once done or failed
      if (status === 'completed' || status === 'failed') return false
      return 3000
    },
  })
}

// Trigger CSP schedule generation for a semester
export function useGenerateSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: GenerateScheduleInput) => {
      const { data } = await apiClient.post<Schedule>(
        ENDPOINTS.timetable.generate(input.semester_id),
        { timeout_seconds: input.timeout_seconds ?? 60 },
      )
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['schedules'] })
    },
  })
}

// Fetch AI-ranked teacher suggestions for a schedule entry
export function useTeacherSuggestions(scheduleId: string, entryId: string | null) {
  return useQuery({
    queryKey: ['schedules', scheduleId, 'suggestions', entryId] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<TeacherSuggestion[]>(
        `/timetable/schedules/${scheduleId}/entries/${entryId}/suggestions`,
      )
      return data
    },
    enabled: Boolean(scheduleId) && Boolean(entryId),
  })
}

// Manual override: assign a specific teacher to an entry
export function useAssignTeacher(scheduleId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: AssignTeacherInput) => {
      const { data } = await apiClient.put(
        `/timetable/schedules/${scheduleId}/entries/${input.entry_id}/assign`,
        { teacher_id: input.teacher_id },
      )
      return data
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['schedules', scheduleId] })
    },
  })
}
