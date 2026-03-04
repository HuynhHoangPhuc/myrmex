import * as React from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import {
  Bell, CheckCheck, Calendar, BookOpen, Users, GraduationCap,
  ShieldCheck, Megaphone, UserCheck, Clock, AlertCircle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { NotificationPreferences } from '@/notifications/components/notification-preferences'
import { useNotifications, type Notification } from '@/notifications/hooks/use-notifications'
import { cn } from '@/lib/utils/cn'

const searchSchema = z.object({
  page: z.number().catch(1),
  unreadOnly: z.boolean().catch(false),
  tab: z.enum(['all', 'preferences']).catch('all'),
})

export const Route = createFileRoute('/_authenticated/notifications/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: NotificationsPage,
})

const EVENT_ICON: Record<string, React.ElementType> = {
  'schedule.published': Calendar,
  'schedule.changed': Calendar,
  'enrollment.approved': CheckCheck,
  'enrollment.rejected': AlertCircle,
  'enrollment.requested': GraduationCap,
  'grade.posted': BookOpen,
  'assignment.changed': Users,
  'semester.created': Calendar,
  'semester.deadline': Clock,
  'role.changed': ShieldCheck,
  'system.announcement': Megaphone,
  'teacher.added': UserCheck,
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

function NotificationRow({
  notification,
  onRead,
}: {
  notification: Notification
  onRead: (id: string) => void
}) {
  const navigate = useNavigate()
  const isUnread = notification.read_at === null
  const link = notification.data?.link
  const Icon = EVENT_ICON[notification.type] ?? Bell

  function handleClick() {
    if (isUnread) onRead(notification.id)
    if (link) void navigate({ to: link })
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      className={cn(
        'flex w-full items-start gap-3 p-4 text-left transition-colors hover:bg-accent/50',
        isUnread ? 'bg-accent/20' : '',
      )}
    >
      <div className={cn('mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full', isUnread ? 'bg-primary/10' : 'bg-muted')}>
        <Icon className={cn('h-4 w-4', isUnread ? 'text-primary' : 'text-muted-foreground')} />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className={cn('text-sm leading-snug', isUnread ? 'font-medium' : 'font-normal')}>
            {notification.title}
          </p>
          <div className="flex shrink-0 items-center gap-1.5">
            <span className="text-xs text-muted-foreground/70">{timeAgo(notification.created_at)}</span>
            {isUnread && <span className="h-2 w-2 rounded-full bg-primary" />}
          </div>
        </div>
        <p className="mt-0.5 line-clamp-2 text-sm text-muted-foreground">{notification.body}</p>
      </div>
    </button>
  )
}

const PAGE_SIZE = 25

function NotificationsPage() {
  const { page, unreadOnly, tab } = Route.useSearch()
  const navigate = Route.useNavigate()

  const { notifications, total, unreadCount, markRead, markAllRead } = useNotifications(page, PAGE_SIZE)

  // Client-side unread filter (server doesn't support filter param yet)
  const displayed = unreadOnly ? notifications.filter((n) => n.read_at === null) : notifications
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  function setTab(t: 'all' | 'preferences') {
    void navigate({ search: (prev) => ({ ...prev, tab: t, page: 1 }) })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Notifications"
        description="Stay updated on schedule changes, enrollment status, and announcements."
        actions={
          tab === 'all' && unreadCount > 0 ? (
            <Button variant="outline" size="sm" className="gap-1.5" onClick={() => markAllRead()}>
              <CheckCheck className="h-4 w-4" />
              Mark all read
            </Button>
          ) : undefined
        }
      />

      {/* Tab bar */}
      <div className="flex gap-1 border-b">
        {(['all', 'preferences'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={cn(
              'px-4 pb-2 text-sm font-medium transition-colors',
              tab === t
                ? 'border-b-2 border-primary text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {t === 'all' ? (
              <span className="flex items-center gap-2">
                All notifications
                {unreadCount > 0 && (
                  <Badge variant="secondary" className="h-5 px-1.5 text-xs">
                    {unreadCount}
                  </Badge>
                )}
              </span>
            ) : (
              'Preferences'
            )}
          </button>
        ))}
      </div>

      {tab === 'preferences' ? (
        <NotificationPreferences />
      ) : (
        <Card className="overflow-hidden p-0">
          {/* Filters */}
          <div className="flex items-center gap-3 px-4 py-3">
            <button
              type="button"
              onClick={() => void navigate({ search: (prev) => ({ ...prev, unreadOnly: false, page: 1 }) })}
              className={cn(
                'rounded-md px-3 py-1 text-sm font-medium transition-colors',
                !unreadOnly ? 'bg-secondary text-secondary-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              All
            </button>
            <button
              type="button"
              onClick={() => void navigate({ search: (prev) => ({ ...prev, unreadOnly: true, page: 1 }) })}
              className={cn(
                'rounded-md px-3 py-1 text-sm font-medium transition-colors',
                unreadOnly ? 'bg-secondary text-secondary-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              Unread
            </button>
          </div>

          <Separator />

          <CardContent className="p-0">
            {displayed.length === 0 ? (
              <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
                <Bell className="h-10 w-10 opacity-30" />
                <p className="text-sm">{unreadOnly ? 'No unread notifications' : 'No notifications yet'}</p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {displayed.map((n) => (
                  <NotificationRow key={n.id} notification={n} onRead={markRead} />
                ))}
              </div>
            )}
          </CardContent>

          {/* Pagination */}
          {totalPages > 1 && (
            <>
              <Separator />
              <div className="flex items-center justify-between px-4 py-3 text-sm text-muted-foreground">
                <span>
                  Page {page} of {totalPages} · {total} total
                </span>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => void navigate({ search: (prev) => ({ ...prev, page: prev.page - 1 }) })}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => void navigate({ search: (prev) => ({ ...prev, page: prev.page + 1 }) })}
                  >
                    Next
                  </Button>
                </div>
              </div>
            </>
          )}
        </Card>
      )}
    </div>
  )
}
