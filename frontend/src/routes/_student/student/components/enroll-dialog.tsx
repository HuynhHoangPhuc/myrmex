// Enrollment request modal with prerequisite checker and optional note input.
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useCheckPrerequisites } from '../-hooks/use-my-prerequisites'
import type { Subject } from '@/modules/subject/types'

export interface EnrollDialogState {
  subject: Subject
  semesterId: string
}

interface EnrollDialogProps {
  target: EnrollDialogState | null
  note: string
  isPending: boolean
  onNoteChange: (value: string) => void
  onConfirm: () => void
  onClose: () => void
}

export function EnrollDialog({
  target,
  note,
  isPending,
  onNoteChange,
  onConfirm,
  onClose,
}: EnrollDialogProps) {
  const { data: prereqResult, isLoading: prereqLoading } = useCheckPrerequisites(
    target?.subject.id ?? null,
  )

  const hasStrictMissing =
    prereqResult && !prereqResult.can_enroll &&
    prereqResult.missing.some((m) => m.type === 'strict')

  return (
    <Dialog open={Boolean(target)} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Request Enrollment</DialogTitle>
        </DialogHeader>
        {target && (
          <div className="space-y-4">
            <div className="rounded-md bg-muted px-4 py-2 text-sm">
              <p className="font-medium">{target.subject.name}</p>
              <p className="text-xs text-muted-foreground">{target.subject.code}</p>
            </div>

            {/* Prerequisite check */}
            {prereqLoading && (
              <p className="flex items-center gap-2 text-xs text-muted-foreground">
                <LoadingSpinner size="sm" /> Checking prerequisites…
              </p>
            )}
            {!prereqLoading && prereqResult && prereqResult.missing.length > 0 && (
              <div className={`rounded-md border px-3 py-2 text-xs ${hasStrictMissing ? 'border-destructive/50 bg-destructive/10 text-destructive' : 'border-amber-500/50 bg-amber-500/10 text-amber-700 dark:text-amber-400'}`}>
                <div className="flex items-center gap-1.5 font-medium mb-1">
                  <AlertTriangle className="h-3.5 w-3.5" />
                  {hasStrictMissing ? 'Missing required prerequisites' : 'Missing recommended prerequisites'}
                </div>
                <ul className="space-y-0.5 pl-1">
                  {prereqResult.missing.map((m) => (
                    <li key={m.subject_id}>
                      {m.subject_code} — {m.subject_name}
                      {m.type === 'recommended' && (
                        <span className="ml-1 opacity-70">(recommended)</span>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            <div className="space-y-1">
              <Label>Note (optional)</Label>
              <Input
                value={note}
                onChange={(e) => onNoteChange(e.target.value)}
                placeholder="Any special request…"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={onClose}>Cancel</Button>
              <Button
                onClick={onConfirm}
                disabled={isPending || Boolean(hasStrictMissing)}
              >
                Submit Request
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
