// Dialog for rejecting an enrollment request with an optional admin note.
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface RejectEnrollmentDialogProps {
  open: boolean
  adminNote: string
  isPending: boolean
  onNoteChange: (value: string) => void
  onConfirm: () => void
  onClose: () => void
}

export function RejectEnrollmentDialog({
  open,
  adminNote,
  isPending,
  onNoteChange,
  onConfirm,
  onClose,
}: RejectEnrollmentDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Reject Enrollment</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-1">
            <Label>Reason (optional)</Label>
            <Input
              value={adminNote}
              onChange={(e) => onNoteChange(e.target.value)}
              placeholder="Explain the reason…"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button variant="destructive" onClick={onConfirm} disabled={isPending}>
              Reject
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
