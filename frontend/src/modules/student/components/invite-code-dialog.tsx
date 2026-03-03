import * as React from 'react'
import { Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useGenerateInviteCode } from '@/modules/student/hooks/use-invite-code'
import { toast } from '@/lib/hooks/use-toast'

interface InviteCodeDialogProps {
  studentId: string
  studentName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function InviteCodeDialog({ studentId, studentName, open, onOpenChange }: InviteCodeDialogProps) {
  const [code, setCode] = React.useState<string | null>(null)
  const [expiresAt, setExpiresAt] = React.useState<string | null>(null)
  const [copied, setCopied] = React.useState(false)
  const mutation = useGenerateInviteCode()

  function handleGenerate() {
    mutation.mutate(studentId, {
      onSuccess: (data) => {
        setCode(data.code)
        setExpiresAt(data.expires_at)
        setCopied(false)
      },
      onError: () => toast({ title: 'Failed to generate invite code', variant: 'destructive' }),
    })
  }

  function handleCopy() {
    if (!code) return
    void navigator.clipboard.writeText(code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  function handleClose(open: boolean) {
    if (!open) {
      setCode(null)
      setExpiresAt(null)
      setCopied(false)
    }
    onOpenChange(open)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Invite Code — {studentName}</DialogTitle>
        </DialogHeader>

        {!code ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Generate a single-use invite code for this student to register their account.
              Any previously active codes will be invalidated.
            </p>
            <Button onClick={handleGenerate} disabled={mutation.isPending} className="w-full">
              {mutation.isPending ? 'Generating…' : 'Generate Code'}
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-xs text-amber-600 dark:text-amber-400 font-medium">
              This code will only be shown once. Share it securely.
            </p>
            <div className="flex items-center gap-2 rounded-md border bg-muted px-3 py-2">
              <code className="flex-1 font-mono text-sm tracking-widest break-all">{code}</code>
              <Button size="icon" variant="ghost" className="h-7 w-7 shrink-0" onClick={handleCopy}>
                {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            {expiresAt && (
              <p className="text-xs text-muted-foreground">
                Expires: {new Date(expiresAt).toLocaleString()}
              </p>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
