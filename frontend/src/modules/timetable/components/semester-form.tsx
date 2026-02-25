import * as React from 'react'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Plus, Trash2 } from 'lucide-react'
import { TextInputField } from '@/components/shared/form-field'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import type { CreateSemesterInput } from '../types'

const semesterSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  academic_year: z.string().min(1, 'Academic year is required'),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string().min(1, 'End date is required'),
})

const DAY_NAMES = ['', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

interface TimeSlotRow {
  day_of_week: number
  start_time: string
  end_time: string
  slot_index: number
}

interface RoomRow {
  name: string
  capacity: number
  room_type: 'lecture' | 'lab' | 'seminar'
}

interface SemesterFormProps {
  onSubmit: (data: CreateSemesterInput) => void
  isLoading?: boolean
}

export function SemesterForm({ onSubmit, isLoading }: SemesterFormProps) {
  const [timeSlots, setTimeSlots] = React.useState<TimeSlotRow[]>([
    { day_of_week: 1, start_time: '07:00', end_time: '09:00', slot_index: 0 },
  ])
  const [rooms, setRooms] = React.useState<RoomRow[]>([
    { name: '', capacity: 40, room_type: 'lecture' },
  ])

  const form = useForm({
    defaultValues: {
      name: '',
      academic_year: '',
      start_date: '',
      end_date: '',
    },
    onSubmit: ({ value }) => {
      onSubmit({
        ...value,
        time_slots: timeSlots.map((ts, i) => ({ ...ts, slot_index: i })),
        rooms,
      } as CreateSemesterInput)
    },
  })

  function addSlot() {
    setTimeSlots((prev) => [
      ...prev,
      { day_of_week: 1, start_time: '07:00', end_time: '09:00', slot_index: prev.length },
    ])
  }

  function removeSlot(idx: number) {
    setTimeSlots((prev) => prev.filter((_, i) => i !== idx))
  }

  function updateSlot(idx: number, field: keyof TimeSlotRow, value: string | number) {
    setTimeSlots((prev) => prev.map((s, i) => i === idx ? { ...s, [field]: value } : s))
  }

  function addRoom() {
    setRooms((prev) => [...prev, { name: '', capacity: 40, room_type: 'lecture' }])
  }

  function removeRoom(idx: number) {
    setRooms((prev) => prev.filter((_, i) => i !== idx))
  }

  function updateRoom(idx: number, field: keyof RoomRow, value: string | number) {
    setRooms((prev) => prev.map((r, i) => i === idx ? { ...r, [field]: value } : r))
  }

  return (
    <form onSubmit={(e) => { e.preventDefault(); void form.handleSubmit() }} className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <form.Field name="name" validators={{ onChange: semesterSchema.shape.name }}
          children={(field) => (
            <TextInputField label="Semester Name" required placeholder="e.g. Fall 2024"
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />
        <form.Field name="academic_year" validators={{ onChange: semesterSchema.shape.academic_year }}
          children={(field) => (
            <TextInputField label="Academic Year" required placeholder="e.g. 2024-2025"
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />
        <form.Field name="start_date" validators={{ onChange: semesterSchema.shape.start_date }}
          children={(field) => (
            <TextInputField label="Start Date" type="date" required
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />
        <form.Field name="end_date" validators={{ onChange: semesterSchema.shape.end_date }}
          children={(field) => (
            <TextInputField label="End Date" type="date" required
              value={field.state.value} onChange={(e) => field.handleChange(e.target.value)}
              error={field.state.meta.errors[0]?.toString()} />
          )} />
      </div>

      {/* Time Slots */}
      <div>
        <div className="mb-2 flex items-center justify-between">
          <h3 className="text-sm font-semibold">Time Slots</h3>
          <Button type="button" variant="outline" size="sm" onClick={addSlot}>
            <Plus className="mr-1 h-3 w-3" /> Add Slot
          </Button>
        </div>
        <div className="space-y-2">
          {timeSlots.map((slot, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <select value={slot.day_of_week} onChange={(e) => updateSlot(idx, 'day_of_week', Number(e.target.value))}
                className="h-9 rounded-md border border-input bg-transparent px-2 text-sm">
                {DAY_NAMES.slice(1).map((d, i) => <option key={i+1} value={i+1}>{d}</option>)}
              </select>
              <input type="time" value={slot.start_time} onChange={(e) => updateSlot(idx, 'start_time', e.target.value)}
                className="h-9 rounded-md border border-input bg-transparent px-2 text-sm" />
              <span className="text-muted-foreground text-sm">–</span>
              <input type="time" value={slot.end_time} onChange={(e) => updateSlot(idx, 'end_time', e.target.value)}
                className="h-9 rounded-md border border-input bg-transparent px-2 text-sm" />
              <Button type="button" variant="ghost" size="icon" className="h-8 w-8 shrink-0"
                onClick={() => removeSlot(idx)}>
                <Trash2 className="h-3.5 w-3.5 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      </div>

      {/* Rooms */}
      <div>
        <div className="mb-2 flex items-center justify-between">
          <h3 className="text-sm font-semibold">Rooms</h3>
          <Button type="button" variant="outline" size="sm" onClick={addRoom}>
            <Plus className="mr-1 h-3 w-3" /> Add Room
          </Button>
        </div>
        <div className="space-y-2">
          {rooms.map((room, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <input placeholder="Room name" value={room.name}
                onChange={(e) => updateRoom(idx, 'name', e.target.value)}
                className="h-9 flex-1 rounded-md border border-input bg-transparent px-3 text-sm" />
              <input type="number" placeholder="Capacity" value={room.capacity} min={1}
                onChange={(e) => updateRoom(idx, 'capacity', Number(e.target.value))}
                className="h-9 w-24 rounded-md border border-input bg-transparent px-3 text-sm" />
              <select value={room.room_type} onChange={(e) => updateRoom(idx, 'room_type', e.target.value as RoomRow['room_type'])}
                className="h-9 rounded-md border border-input bg-transparent px-2 text-sm">
                <option value="lecture">Lecture</option>
                <option value="lab">Lab</option>
                <option value="seminar">Seminar</option>
              </select>
              <Button type="button" variant="ghost" size="icon" className="h-8 w-8 shrink-0"
                onClick={() => removeRoom(idx)}>
                <Trash2 className="h-3.5 w-3.5 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end pt-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
          Create Semester
        </Button>
      </div>
    </form>
  )
}
