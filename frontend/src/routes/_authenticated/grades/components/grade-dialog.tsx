// Grade assignment modal with numeric input and letter grade preview.
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import type { EnrollmentRequest } from '@/modules/student/types'

interface GradeDialogProps {
  enrollment: EnrollmentRequest | null
  gradeValue: string
  notes: string
  preview: string | null
  isPending: boolean
  subjectMap: Map<string, string>
  studentMap: Map<string, string>
  onGradeChange: (value: string) => void
  onNotesChange: (value: string) => void
  onConfirm: () => void
  onClose: () => void
}

export function GradeDialog({
  enrollment,
  gradeValue,
  notes,
  preview,
  isPending,
  subjectMap,
  studentMap,
  onGradeChange,
  onNotesChange,
  onConfirm,
  onClose,
}: GradeDialogProps) {
  return (
    <Dialog open={Boolean(enrollment)} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>
            Assign Grade
            {enrollment && (
              <span className="block text-sm font-normal text-muted-foreground mt-0.5">
                {subjectMap.get(enrollment.subject_id) ?? 'Subject'} ·{' '}
                {studentMap.get(enrollment.student_id) ?? 'Student'}
              </span>
            )}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1">
            <Label>Grade (0 – 10)</Label>
            <div className="flex gap-2 items-center">
              <Input
                type="number"
                min={0}
                max={10}
                step={0.1}
                value={gradeValue}
                onChange={(e) => onGradeChange(e.target.value)}
                placeholder="e.g. 8.5"
                className="w-32"
              />
              {preview && (
                <span className="text-lg font-bold text-primary">{preview}</span>
              )}
            </div>
          </div>
          <div className="space-y-1">
            <Label>Notes (optional)</Label>
            <Input
              value={notes}
              onChange={(e) => onNotesChange(e.target.value)}
              placeholder="Optional remarks…"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button onClick={onConfirm} disabled={gradeValue === '' || isPending}>
              Save Grade
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
