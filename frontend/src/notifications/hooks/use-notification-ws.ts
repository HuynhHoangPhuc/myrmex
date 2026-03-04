import { useCallback, useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { authStore } from '@/lib/stores/auth-store'
import { toast } from '@/lib/hooks/use-toast'

function getWsBaseUrl(): string {
  const apiUrl = import.meta.env.VITE_API_URL
  if (apiUrl) {
    return apiUrl.replace(/\/api$/, '').replace(/^http/, 'ws')
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}`
}

const NOTIF_WS_URL = `${getWsBaseUrl()}/ws/notifications`

interface PushEvent {
  type: string
  id: string
  notif_type: string
  title: string
  body: string
  unread_count: number
}

/**
 * Maintains a single WebSocket connection to /ws/notifications.
 * On push: invalidates notification queries + shows a toast.
 * Mount once in AppLayout so there is only one WS connection.
 */
export function useNotificationWebSocket() {
  const queryClient = useQueryClient()
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const connect = useCallback(() => {
    const token = authStore.getAccessToken()
    if (!token) return

    const ws = new WebSocket(`${NOTIF_WS_URL}?token=${encodeURIComponent(token)}`)
    wsRef.current = ws

    ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data as string) as PushEvent
        if (data.type === 'notification') {
          // Refresh all notification queries
          void queryClient.invalidateQueries({ queryKey: ['notifications'] })
          // Show a toast for the incoming notification
          toast({ title: data.title, description: data.body })
        }
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      // Reconnect after 3s on unexpected close
      reconnectRef.current = setTimeout(connect, 3_000)
    }
  }, [queryClient])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectRef.current) clearTimeout(reconnectRef.current)
      wsRef.current?.close()
    }
  }, [connect])
}
