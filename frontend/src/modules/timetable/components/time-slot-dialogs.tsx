// Dialog components for time slot management: add slot, apply preset, delete confirm.
import { useState } from 'react'
import { Plus, Trash2, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
import { DAY_NAMES, DAY_OPTIONS, PERIOD_LABELS, PRESETS } from './time-slot-constants'
import type { TimeSlot, CreateTimeSlotInput, TimeSlotPreset } from '../types'

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

export function AddSlotDialog({ semesterId }: AddSlotDialogProps) {
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

export function PresetDialog({ semesterId, hasSlots }: PresetDialogProps) {
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
  semesterId: string
}

export function DeleteConfirmDialog({ slot, semesterId }: DeleteConfirmDialogProps) {
  const [open, setOpen] = useState(false)
  const deleteSlot = useDeleteTimeSlot(semesterId)

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
            disabled={deleteSlot.isPending}
            onClick={() => { deleteSlot.mutate(slot.id); setOpen(false) }}
          >
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
