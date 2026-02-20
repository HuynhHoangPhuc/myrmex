import * as React from 'react'
import { Link } from '@tanstack/react-router'
import { Loader2, CheckCircle, XCircle, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useGenerateSchedule, useGenerationStatus } from '../hooks/use-schedules'
import type { ScheduleStatus } from '../types'

const STATUS_CONFIG: Record<ScheduleStatus, { label: string; icon: React.ElementType; color: string }> = {
  pending: { label: 'Pending', icon: Clock, color: 'text-muted-foreground' },
  generating: { label: 'Generating…', icon: Loader2, color: 'text-blue-600' },
  completed: { label: 'Completed', icon: CheckCircle, color: 'text-green-600' },
  failed: { label: 'Failed', icon: XCircle, color: 'text-destructive' },
}

interface GenerationPanelProps {
  semesterId: string
}

// Trigger schedule generation and poll status until done
export function GenerationPanel({ semesterId }: GenerationPanelProps) {
  const [scheduleId, setScheduleId] = React.useState<string | null>(null)
  const generateMutation = useGenerateSchedule()
  const { data: schedule } = useGenerationStatus(scheduleId)

  function handleGenerate() {
    generateMutation.mutate(
      { semester_id: semesterId, timeout_seconds: 60 },
      { onSuccess: (data) => setScheduleId(data.id) },
    )
  }

  const status = schedule?.status
  const statusCfg = status ? STATUS_CONFIG[status] : null
  const StatusIcon = statusCfg?.icon

  return (
    <div className="rounded-lg border p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold">Schedule Generation</h3>
          <p className="text-sm text-muted-foreground mt-0.5">
            CSP solver with backtracking + AC-3. Returns best partial solution on timeout.
          </p>
        </div>
        <Button onClick={handleGenerate} disabled={generateMutation.isPending || status === 'generating'}>
          {generateMutation.isPending ? (
            <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Starting…</>
          ) : 'Generate Schedule'}
        </Button>
      </div>

      {schedule && statusCfg && (
        <div className="rounded-md bg-muted/50 p-4 space-y-3">
          <div className="flex items-center gap-2">
            {StatusIcon && (
              <StatusIcon className={`h-4 w-4 ${statusCfg.color} ${status === 'generating' ? 'animate-spin' : ''}`} />
            )}
            <span className={`text-sm font-medium ${statusCfg.color}`}>{statusCfg.label}</span>
          </div>

          {status === 'completed' && (
            <div className="flex flex-wrap gap-2 items-center">
              <Badge variant="secondary">Score: {schedule.score.toFixed(2)}</Badge>
              <Badge variant={schedule.hard_violations === 0 ? 'secondary' : 'destructive'}>
                {schedule.hard_violations} hard violations
              </Badge>
              <Badge variant="outline">{schedule.soft_violations} soft violations</Badge>
              <Link
                to="/timetable/schedules/$id"
                params={{ id: scheduleId! }}
                className="ml-auto text-sm text-primary underline-offset-4 hover:underline"
              >
                View Schedule →
              </Link>
            </div>
          )}

          {status === 'failed' && (
            <p className="text-sm text-destructive">Generation failed. Try adjusting constraints or increasing timeout.</p>
          )}
        </div>
      )}
    </div>
  )
}
