import { useNavigate, Link } from '@tanstack/react-router'
import { CheckCheck, Inbox } from 'lucide-react'

// Simple relative time without external deps
function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}
import { Button } from '@/components/ui/button'
import {
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { useNotifications, type Notification } from '@/notifications/hooks/use-notifications'

function NotificationItem({
  notification,
  onRead,
}: {
  notification: Notification
  onRead: (id: string) => void
}) {
  const navigate = useNavigate()
  const isUnread = notification.read_at === null
  const link = notification.data?.link

  function handleClick() {
    if (isUnread) onRead(notification.id)
    if (link) void navigate({ to: link })
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      className={[
        'flex w-full flex-col gap-1 px-4 py-3 text-left transition-colors hover:bg-accent',
        isUnread ? 'bg-accent/40' : '',
      ].join(' ')}
    >
      <div className="flex items-start justify-between gap-2">
        <p className={`text-sm leading-snug ${isUnread ? 'font-medium' : 'font-normal'}`}>
          {notification.title}
        </p>
        {isUnread && <span className="mt-1 h-2 w-2 shrink-0 rounded-full bg-primary" />}
      </div>
      <p className="line-clamp-2 text-xs text-muted-foreground">{notification.body}</p>
      <p className="text-xs text-muted-foreground/70">{timeAgo(notification.created_at)}</p>
    </button>
  )
}

export function NotificationPanel() {
  const { notifications, unreadCount, markRead, markAllRead } = useNotifications(1, 15)

  return (
    <DropdownMenuContent align="end" className="w-80 p-0">
      <DropdownMenuLabel className="flex items-center justify-between px-4 py-3">
        <span className="font-semibold">Notifications</span>
        {unreadCount > 0 && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 px-2 text-xs"
            onClick={() => markAllRead()}
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Mark all read
          </Button>
        )}
      </DropdownMenuLabel>
      <DropdownMenuSeparator className="my-0" />

      <div className="max-h-96 overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="flex flex-col items-center gap-2 px-4 py-8 text-muted-foreground">
            <Inbox className="h-8 w-8 opacity-40" />
            <p className="text-sm">No notifications yet</p>
          </div>
        ) : (
          notifications.map((n) => (
            <NotificationItem key={n.id} notification={n} onRead={markRead} />
          ))
        )}
      </div>

      <DropdownMenuSeparator className="my-0" />
      <div className="px-4 py-2">
        <Link
          to="/notifications"
          className="block text-center text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          View all notifications
        </Link>
      </div>
    </DropdownMenuContent>
  )
}
