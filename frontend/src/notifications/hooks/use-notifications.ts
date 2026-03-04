import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'

export interface Notification {
  id: string
  user_id: string
  type: string
  channel: string
  title: string
  body: string
  data?: { resource_type?: string; resource_id?: string; link?: string }
  read_at: string | null
  created_at: string
}

// Per-channel per-event preference row returned by the new API
export interface PreferenceItem {
  event_type: string
  channel: string
  enabled: boolean
}

/** Poll unread count + paginated list via REST. Real-time push handled by use-notification-ws. */
export function useNotifications(page = 1, pageSize = 20) {
  const queryClient = useQueryClient()

  const { data: listData } = useQuery({
    queryKey: ['notifications', page, pageSize],
    queryFn: () =>
      apiClient
        .get<{ data: Notification[]; total: number; page: number; page_size: number }>(
          ENDPOINTS.notifications.list,
          { params: { page, page_size: pageSize } },
        )
        .then((r) => r.data),
    staleTime: 30_000,
  })

  const { data: unreadCount = 0 } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () =>
      apiClient
        .get<{ unread_count: number }>(ENDPOINTS.notifications.unreadCount)
        .then((r) => r.data.unread_count),
    staleTime: 10_000,
  })

  const markRead = useMutation({
    mutationFn: (id: string) => apiClient.patch(ENDPOINTS.notifications.markRead(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
  })

  const markAllRead = useMutation({
    mutationFn: () => apiClient.post(ENDPOINTS.notifications.markAllRead),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
  })

  return {
    notifications: listData?.data ?? [],
    total: listData?.total ?? 0,
    unreadCount,
    markRead: markRead.mutate,
    markAllRead: markAllRead.mutate,
  }
}

export function useNotificationPreferences() {
  const queryClient = useQueryClient()

  const { data: preferences, isLoading } = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: () =>
      apiClient
        .get<{ preferences: PreferenceItem[] }>(ENDPOINTS.notifications.preferences)
        .then((r) => r.data.preferences),
  })

  const update = useMutation({
    mutationFn: (prefs: PreferenceItem[]) =>
      apiClient.put(ENDPOINTS.notifications.preferences, { preferences: prefs }).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-preferences'] })
    },
  })

  return { preferences: preferences ?? [], isLoading, updatePreferences: update.mutate, isPending: update.isPending }
}
