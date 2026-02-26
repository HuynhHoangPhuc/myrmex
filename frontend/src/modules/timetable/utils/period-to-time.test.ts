import { periodToTimeLabel, periodToStartTime } from './period-to-time'

describe('periodToTimeLabel', () => {
  it('returns correct time range for valid periods', () => {
    expect(periodToTimeLabel(1, 2)).toBe('08:00–11:15')
    expect(periodToTimeLabel(3, 4)).toBe('11:30–14:45')
  })

  it('returns same start/end for single period', () => {
    expect(periodToTimeLabel(1, 1)).toBe('08:00–09:30')
  })

  it('falls back to P-notation for out-of-range periods', () => {
    expect(periodToTimeLabel(0, 9)).toBe('P0–P9')
  })
})

describe('periodToStartTime', () => {
  it('returns start time for known periods', () => {
    expect(periodToStartTime(1)).toBe('08:00')
    expect(periodToStartTime(5)).toBe('15:00')
    expect(periodToStartTime(8)).toBe('20:15')
  })

  it('falls back for unknown periods', () => {
    expect(periodToStartTime(0)).toBe('P0')
    expect(periodToStartTime(10)).toBe('P10')
  })
})
