import { useNotificationWebSocket } from '@/notifications/hooks/use-notification-ws'

/**
 * Mounts the WebSocket connection for real-time notification push.
 * Renders nothing — side-effect only. Mount once in AppLayout.
 */
export function NotificationToast() {
  useNotificationWebSocket()
  return null
}
