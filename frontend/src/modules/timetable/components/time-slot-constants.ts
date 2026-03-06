// Static constants for time slot management: day names, options, period labels, presets.
import type { TimeSlotPreset } from '../types'

export const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export const DAY_OPTIONS = [
  { value: 0, label: 'Monday' },
  { value: 1, label: 'Tuesday' },
  { value: 2, label: 'Wednesday' },
  { value: 3, label: 'Thursday' },
  { value: 4, label: 'Friday' },
  { value: 5, label: 'Saturday' },
]

export const PERIOD_LABELS: Record<number, string> = {
  1: '1 (08:00)', 2: '2 (09:45)', 3: '3 (11:30)', 4: '4 (13:15)',
  5: '5 (15:00)', 6: '6 (16:45)', 7: '7 (18:30)', 8: '8 (20:15)',
}

export const PRESETS: { value: TimeSlotPreset; label: string; description: string }[] = [
  { value: 'standard', label: 'Standard (Mon–Sat, 3 slots/day)', description: '18 slots across all weekdays, periods 1–6' },
  { value: 'mwf', label: 'MWF (Mon/Wed/Fri, 4 slots/day)', description: '12 slots, periods 1–8' },
  { value: 'tuth', label: 'TuTh (Tue/Thu, 4 slots/day)', description: '8 slots, periods 1–8' },
]
