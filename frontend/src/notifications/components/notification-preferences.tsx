import * as React from 'react'
import { cn } from '@/lib/utils/cn'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useNotificationPreferences, type PreferenceItem } from '@/notifications/hooks/use-notifications'
import { toast } from '@/lib/hooks/use-toast'

const EVENT_TYPE_LABELS: Record<string, string> = {
  'schedule.published': 'Schedule Published',
  'schedule.changed': 'Schedule Changed',
  'enrollment.approved': 'Enrollment Approved',
  'enrollment.rejected': 'Enrollment Rejected',
  'enrollment.requested': 'New Enrollment Request',
  'grade.posted': 'Grade Posted',
  'assignment.changed': 'Teaching Assignment Changed',
  'semester.created': 'New Semester Created',
  'semester.deadline': 'Semester Deadline Reminder',
  'role.changed': 'Role Changed',
  'system.announcement': 'System Announcement',
  'teacher.added': 'New Teacher Added',
}

// Groups for readability
const EVENT_GROUPS: Array<{ label: string; types: string[] }> = [
  {
    label: 'Schedule',
    types: ['schedule.published', 'schedule.changed', 'assignment.changed'],
  },
  {
    label: 'Enrollment',
    types: ['enrollment.approved', 'enrollment.rejected', 'enrollment.requested'],
  },
  {
    label: 'Academic',
    types: ['grade.posted', 'semester.created', 'semester.deadline'],
  },
  {
    label: 'System',
    types: ['role.changed', 'system.announcement', 'teacher.added'],
  },
]

const CHANNELS = [
  { key: 'in_app', label: 'In-App' },
  { key: 'email', label: 'Email' },
]

function Toggle({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        checked ? 'bg-primary' : 'bg-input',
      )}
    >
      <span
        className={cn(
          'pointer-events-none inline-block h-4 w-4 rounded-full bg-background shadow-lg ring-0 transition-transform',
          checked ? 'translate-x-4' : 'translate-x-0',
        )}
      />
    </button>
  )
}

export function NotificationPreferences() {
  const { preferences, isLoading, updatePreferences, isPending } = useNotificationPreferences()
  // Local state for immediate toggle feedback before save
  const [local, setLocal] = React.useState<PreferenceItem[] | null>(null)

  // Sync local state when data arrives
  React.useEffect(() => {
    if (preferences.length > 0 && local === null) {
      setLocal(preferences)
    }
  }, [preferences, local])

  const effective = local ?? preferences

  function isEnabled(eventType: string, channel: string): boolean {
    const found = effective.find((p) => p.event_type === eventType && p.channel === channel)
    return found?.enabled ?? true
  }

  function toggle(eventType: string, channel: string, enabled: boolean) {
    setLocal((prev) => {
      const base = prev ?? preferences
      const existing = base.find((p) => p.event_type === eventType && p.channel === channel)
      if (existing) {
        return base.map((p) =>
          p.event_type === eventType && p.channel === channel ? { ...p, enabled } : p,
        )
      }
      return [...base, { event_type: eventType, channel, enabled }]
    })
  }

  function handleSave() {
    if (!local) return
    updatePreferences(local, {
      onSuccess: () => toast({ title: 'Preferences saved' }),
      onError: () => toast({ title: 'Failed to save preferences', variant: 'destructive' }),
    })
  }

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-28 animate-pulse rounded-lg bg-muted" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {EVENT_GROUPS.map((group) => (
        <Card key={group.label}>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">{group.label}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-0">
            {/* Column headers */}
            <div className="mb-2 flex items-center gap-4">
              <div className="flex-1" />
              {CHANNELS.map((ch) => (
                <span key={ch.key} className="w-16 text-center text-xs font-medium text-muted-foreground">
                  {ch.label}
                </span>
              ))}
            </div>

            {group.types.map((eventType, idx) => (
              <div
                key={eventType}
                className={cn(
                  'flex items-center gap-4 py-2',
                  idx !== group.types.length - 1 && 'border-b border-border/50',
                )}
              >
                <span className="flex-1 text-sm">{EVENT_TYPE_LABELS[eventType] ?? eventType}</span>
                {CHANNELS.map((ch) => (
                  <div key={ch.key} className="flex w-16 justify-center">
                    <Toggle
                      checked={isEnabled(eventType, ch.key)}
                      onChange={(v) => toggle(eventType, ch.key, v)}
                    />
                  </div>
                ))}
              </div>
            ))}
          </CardContent>
        </Card>
      ))}

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={isPending || local === null}>
          {isPending ? 'Saving…' : 'Save preferences'}
        </Button>
      </div>
    </div>
  )
}
