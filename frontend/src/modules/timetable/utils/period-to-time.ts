// Standard university period schedule — maps period number to start/end times.
const PERIOD_TIMES: Record<number, { start: string; end: string }> = {
  1: { start: '08:00', end: '09:30' },
  2: { start: '09:45', end: '11:15' },
  3: { start: '11:30', end: '13:00' },
  4: { start: '13:15', end: '14:45' },
  5: { start: '15:00', end: '16:30' },
  6: { start: '16:45', end: '18:15' },
  7: { start: '18:30', end: '20:00' },
  8: { start: '20:15', end: '21:45' },
}

export function periodToTimeLabel(startPeriod: number, endPeriod: number): string {
  const s = PERIOD_TIMES[startPeriod]?.start ?? `P${startPeriod}`
  const e = PERIOD_TIMES[endPeriod]?.end ?? `P${endPeriod}`
  return `${s}–${e}`
}

export function periodToStartTime(period: number): string {
  return PERIOD_TIMES[period]?.start ?? `P${period}`
}
