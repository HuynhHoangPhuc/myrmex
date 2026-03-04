import { useCallback, useEffect, useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { authStore } from '@/lib/stores/auth-store'

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  body: string
  data?: { resource_type?: string; resource_id?: string; link?: string }
  read_at: string | null
  created_at: string
}

export interface NotificationPreferences {
  user_id: string
  email_enabled: boolean
  inapp_enabled: boolean
  disabled_types: string[]
}

// WebSocket URL derived from current origin (same logic as use-chat.ts)
function getWsBaseUrl(): string {
  const apiUrl = import.meta.env.VITE_API_URL
  if (apiUrl) {
    return apiUrl.replace(/\/api$/, '').replace(/^http/, 'ws')
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}`
}

const NOTIF_WS_URL = `${getWsBaseUrl()}/ws/notifications`

/** Poll unread count + paginated list via REST, push-update via WebSocket. */
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

  // WebSocket subscription for real-time push
  const wsRef = useRef<WebSocket | null>(null)

  const connect = useCallback(() => {
    const token = authStore.getAccessToken()
    if (!token) return
    const ws = new WebSocket(`${NOTIF_WS_URL}?token=${encodeURIComponent(token)}`)
    wsRef.current = ws

    ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data as string) as { type: string }
        if (data.type === 'notification') {
          // Invalidate so list + unread badge refresh
          queryClient.invalidateQueries({ queryKey: ['notifications'] })
        }
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      // Reconnect after 3s on unexpected close
      setTimeout(connect, 3_000)
    }
  }, [queryClient])

  useEffect(() => {
    connect()
    return () => wsRef.current?.close()
  }, [connect])

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

  const { data: preferences } = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: () =>
      apiClient
        .get<NotificationPreferences>(ENDPOINTS.notifications.preferences)
        .then((r) => r.data),
  })

  const update = useMutation({
    mutationFn: (prefs: Omit<NotificationPreferences, 'user_id'>) =>
      apiClient.put(ENDPOINTS.notifications.preferences, prefs).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-preferences'] })
    },
  })

  return { preferences, updatePreferences: update.mutate }
}
