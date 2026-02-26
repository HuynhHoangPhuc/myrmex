import { formatDate, formatRelativeTime } from './format-date'

describe('formatDate', () => {
  it('formats ISO date to locale string', () => {
    // Use a fixed date to avoid timezone issues
    const result = formatDate('2026-02-20T00:00:00Z')
    expect(result).toContain('2026')
    expect(result).toContain('Feb')
    expect(result).toContain('20')
  })
})

describe('formatRelativeTime', () => {
  it('returns "just now" for recent timestamps', () => {
    const now = new Date().toISOString()
    expect(formatRelativeTime(now)).toBe('just now')
  })

  it('returns minutes ago for timestamps within an hour', () => {
    const fiveMinAgo = new Date(Date.now() - 5 * 60_000).toISOString()
    expect(formatRelativeTime(fiveMinAgo)).toBe('5m ago')
  })

  it('returns hours ago for timestamps within a day', () => {
    const twoHoursAgo = new Date(Date.now() - 2 * 3600_000).toISOString()
    expect(formatRelativeTime(twoHoursAgo)).toBe('2h ago')
  })

  it('returns days ago for older timestamps', () => {
    const threeDaysAgo = new Date(Date.now() - 3 * 86400_000).toISOString()
    expect(formatRelativeTime(threeDaysAgo)).toBe('3d ago')
  })
})
