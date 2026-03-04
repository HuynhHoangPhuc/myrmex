import { Bell } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useNotifications } from '@/notifications/hooks/use-notifications'

interface NotificationBellProps {
  open: boolean
  onToggle: () => void
}

export function NotificationBell({ open, onToggle }: NotificationBellProps) {
  const { unreadCount } = useNotifications()

  return (
    <Button
      variant={open ? 'secondary' : 'ghost'}
      size="icon"
      className="relative h-9 w-9"
      onClick={onToggle}
      aria-label={`Notifications${unreadCount > 0 ? ` (${unreadCount} unread)` : ''}`}
    >
      <Bell className="h-4 w-4" />
      {unreadCount > 0 && (
        <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-destructive-foreground">
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      )}
    </Button>
  )
}
