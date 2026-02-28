// Manages time slots for a semester: add, delete, apply presets.
import { useState } from 'react'
import { Plus, Trash2, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useCreateTimeSlot, useDeleteTimeSlot, useApplyTimeSlotPreset } from '../hooks/use-semesters'
import type { TimeSlot, CreateTimeSlotInput, TimeSlotPreset } from '../types'

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const DAY_OPTIONS = [
  { value: 0, label: 'Monday' },
  { value: 1, label: 'Tuesday' },
  { value: 2, label: 'Wednesday' },
  { value: 3, label: 'Thursday' },
  { value: 4, label: 'Friday' },
  { value: 5, label: 'Saturday' },
]

const PERIOD_LABELS: Record<number, string> = {
  1: '1 (08:00)', 2: '2 (09:45)', 3: '3 (11:30)', 4: '4 (13:15)',
  5: '5 (15:00)', 6: '6 (16:45)', 7: '7 (18:30)', 8: '8 (20:15)',
}

const PRESETS: { value: TimeSlotPreset; label: string; description: string }[] = [
  { value: 'standard', label: 'Standard (Mon–Sat, 3 slots/day)', description: '18 slots across all weekdays, periods 1–6' },
  { value: 'mwf', label: 'MWF (Mon/Wed/Fri, 4 slots/day)', description: '12 slots, periods 1–8' },
  { value: 'tuth', label: 'TuTh (Tue/Thu, 4 slots/day)', description: '8 slots, periods 1–8' },
]

// Native <select> styled to match the design system
function NativeSelect({
  value,
  onChange,
  children,
}: {
  value: string | number
  onChange: (value: number) => void
  children: React.ReactNode
}) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(Number(e.target.value))}
      className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
    >
      {children}
    </select>
  )
}

interface AddSlotDialogProps {
  semesterId: string
}

function AddSlotDialog({ semesterId }: AddSlotDialogProps) {
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<CreateTimeSlotInput>({ day_of_week: 0, start_period: 1, end_period: 2 })
  const createSlot = useCreateTimeSlot(semesterId)

  const invalid = form.end_period <= form.start_period

  const handleSubmit = () => {
    createSlot.mutate(form, { onSuccess: () => setOpen(false) })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm" variant="outline">
          <Plus className="mr-1.5 h-3.5 w-3.5" />Add slot
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Add time slot</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>Day</Label>
            <NativeSelect
              value={form.day_of_week}
              onChange={(v) => setForm((f) => ({ ...f, day_of_week: v }))}
            >
              {DAY_OPTIONS.map((d) => (
                <option key={d.value} value={d.value}>{d.label}</option>
              ))}
            </NativeSelect>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>Start period</Label>
              <NativeSelect
                value={form.start_period}
                onChange={(v) => setForm((f) => ({ ...f, start_period: v }))}
              >
                {[1, 2, 3, 4, 5, 6, 7].map((p) => (
                  <option key={p} value={p}>{PERIOD_LABELS[p]}</option>
                ))}
              </NativeSelect>
            </div>
            <div className="space-y-1.5">
              <Label>End period</Label>
              <NativeSelect
                value={form.end_period}
                onChange={(v) => setForm((f) => ({ ...f, end_period: v }))}
              >
                {[2, 3, 4, 5, 6, 7, 8].map((p) => (
                  <option key={p} value={p}>{PERIOD_LABELS[p]}</option>
                ))}
              </NativeSelect>
            </div>
          </div>
          {invalid && (
            <p className="text-xs text-destructive">End period must be greater than start period.</p>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={createSlot.isPending || invalid}>
            {createSlot.isPending ? 'Adding…' : 'Add'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface PresetDialogProps {
  semesterId: string
  hasSlots: boolean
}

function PresetDialog({ semesterId, hasSlots }: PresetDialogProps) {
  const [open, setOpen] = useState(false)
  const [selected, setSelected] = useState<TimeSlotPreset>('standard')
  const applyPreset = useApplyTimeSlotPreset(semesterId)

  const handleApply = () => {
    applyPreset.mutate(selected, { onSuccess: () => setOpen(false) })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm" variant="outline">
          <Zap className="mr-1.5 h-3.5 w-3.5" />Use preset
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Apply time slot preset</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          {hasSlots && (
            <p className="text-sm text-amber-600 dark:text-amber-400">
              Applying a preset will replace all existing time slots.
            </p>
          )}
          {PRESETS.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => setSelected(p.value)}
              className={`w-full rounded-lg border px-4 py-3 text-left transition-colors ${
                selected === p.value
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:bg-muted/50'
              }`}
            >
              <p className="text-sm font-medium">{p.label}</p>
              <p className="text-xs text-muted-foreground">{p.description}</p>
            </button>
          ))}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={handleApply} disabled={applyPreset.isPending}>
            {applyPreset.isPending ? 'Applying…' : 'Apply preset'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface DeleteConfirmDialogProps {
  slot: TimeSlot
  onConfirm: () => void
  isPending: boolean
}

function DeleteConfirmDialog({ slot, onConfirm, isPending }: DeleteConfirmDialogProps) {
  const [open, setOpen] = useState(false)

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Delete time slot?</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          This will remove{' '}
          <span className="font-medium text-foreground">
            {DAY_NAMES[slot.day_of_week]} {slot.start_time}–{slot.end_time}
          </span>{' '}
          and may affect schedule entries that use it.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            variant="destructive"
            disabled={isPending}
            onClick={() => { onConfirm(); setOpen(false) }}
          >
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface TimeSlotManagerProps {
  semesterId: string
  slots: TimeSlot[]
}

export function TimeSlotManager({ semesterId, slots }: TimeSlotManagerProps) {
  const deleteSlot = useDeleteTimeSlot(semesterId)

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold">Time slots ({slots.length})</h4>
        <div className="flex gap-2">
          <PresetDialog semesterId={semesterId} hasSlots={slots.length > 0} />
          <AddSlotDialog semesterId={semesterId} />
        </div>
      </div>

      {!slots.length ? (
        <p className="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
          No time slots defined. Add one manually or use a preset.
        </p>
      ) : (
        <div className="rounded-lg border divide-y">
          {slots.map((slot, i) => (
            <div key={slot.id} className="flex items-center gap-3 px-4 py-2.5 text-sm">
              <span className="w-12 font-medium">{DAY_NAMES[slot.day_of_week]}</span>
              <span className="text-muted-foreground">
                {slot.start_time} – {slot.end_time}
              </span>
              <Badge variant="outline" className="ml-auto text-xs">Slot {i + 1}</Badge>
              <DeleteConfirmDialog
                slot={slot}
                isPending={deleteSlot.isPending}
                onConfirm={() => deleteSlot.mutate(slot.id)}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
