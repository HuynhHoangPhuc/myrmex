import { CheckCircle, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useTeacherSuggestions } from '../hooks/use-schedules'
import type { ScheduleEntry, TeacherSuggestion } from '../types'

interface TeacherSuggestionListProps {
  scheduleId: string
  entry: ScheduleEntry
  currentTeacherId?: string
  onSelect: (suggestion: TeacherSuggestion) => void
}

// Ranked list of AI-scored teacher suggestions for a schedule entry
export function TeacherSuggestionList({
  scheduleId,
  entry,
  currentTeacherId,
  onSelect,
}: TeacherSuggestionListProps) {
  const { data: suggestions = [], isLoading } = useTeacherSuggestions(scheduleId, entry)

  if (isLoading) return <LoadingSpinner />

  if (suggestions.length === 0) {
    return <p className="text-sm text-muted-foreground py-4 text-center">No suggestions available.</p>
  }

  return (
    <div className="divide-y rounded-md border">
      {suggestions.map((s, idx) => (
        <div
          key={s.teacher_id}
          className={`flex items-start gap-3 p-3 ${s.teacher_id === currentTeacherId ? 'bg-muted/40' : ''}`}
        >
          <span className="mt-0.5 text-xs text-muted-foreground w-5 shrink-0">#{idx + 1}</span>

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-medium">{s.teacher_name}</span>
              {s.teacher_id === currentTeacherId && (
                <Badge variant="secondary" className="text-xs">Current</Badge>
              )}
              {s.is_available ? (
                <CheckCircle className="h-3.5 w-3.5 text-green-600 shrink-0" />
              ) : (
                <XCircle className="h-3.5 w-3.5 text-destructive shrink-0" />
              )}
            </div>

            <div className="mt-1 flex flex-wrap gap-1">
              {s.reasons.map((r) => (
                <span key={r} className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                  {r}
                </span>
              ))}
            </div>
          </div>

          <div className="flex flex-col items-end gap-2 shrink-0">
            <Badge variant="outline" className="text-xs">
              {s.score.toFixed(1)}
            </Badge>
            <Button
              size="sm"
              variant={s.teacher_id === currentTeacherId ? 'secondary' : 'default'}
              disabled={!s.is_available}
              onClick={() => onSelect(s)}
              className="h-7 text-xs"
            >
              Assign
            </Button>
          </div>
        </div>
      ))}
    </div>
  )
}
