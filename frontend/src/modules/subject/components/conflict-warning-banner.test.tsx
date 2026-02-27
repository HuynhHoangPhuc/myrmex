import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, test, expect } from 'vitest'
import { ConflictWarningBanner } from './conflict-warning-banner'
import type { ConflictDetail } from '../types'

const makeConflict = (subjectId: string, subjectName: string, missingCode: string): ConflictDetail => ({
  subject_id: subjectId,
  subject_name: subjectName,
  missing: [{ id: `missing-${missingCode}`, name: `${missingCode} Full Name`, code: missingCode, type: 'hard' }],
})

describe('ConflictWarningBanner', () => {
  test('renders nothing when conflicts array is empty', () => {
    const { container } = render(<ConflictWarningBanner conflicts={[]} />)
    expect(container.firstChild).toBeNull()
  })

  test('renders conflict count and subject names', () => {
    const conflicts = [makeConflict('s1', 'Calculus II', 'MATH101')]
    render(<ConflictWarningBanner conflicts={conflicts} />)
    expect(screen.getByText(/1 subject\(s\) have missing prerequisites/)).toBeTruthy()
    expect(screen.getByText(/Calculus II/)).toBeTruthy()
    expect(screen.getByText(/MATH101/)).toBeTruthy()
  })

  test('renders multiple conflicts', () => {
    const conflicts = [
      makeConflict('s1', 'Calculus II', 'MATH101'),
      makeConflict('s2', 'Physics II', 'PHYS101'),
    ]
    render(<ConflictWarningBanner conflicts={conflicts} />)
    expect(screen.getByText(/2 subject\(s\) have missing prerequisites/)).toBeTruthy()
    expect(screen.getByText(/Calculus II/)).toBeTruthy()
    expect(screen.getByText(/Physics II/)).toBeTruthy()
  })

  test('shows add missing button when onAddMissing provided', () => {
    const onAdd = vi.fn()
    const conflicts = [makeConflict('s1', 'Calc II', 'MATH101')]
    render(<ConflictWarningBanner conflicts={conflicts} onAddMissing={onAdd} />)
    expect(screen.getByText(/Add 1 missing prerequisite/)).toBeTruthy()
  })

  test('calls onAddMissing with all unique missing IDs', () => {
    const onAdd = vi.fn()
    const conflicts = [
      makeConflict('s1', 'Calc II', 'MATH101'),
      makeConflict('s2', 'Physics II', 'PHYS101'),
    ]
    render(<ConflictWarningBanner conflicts={conflicts} onAddMissing={onAdd} />)
    fireEvent.click(screen.getByText(/Add 2 missing prerequisite/))
    expect(onAdd).toHaveBeenCalledOnce()
    const calledWith: string[] = onAdd.mock.calls[0][0]
    expect(calledWith).toHaveLength(2)
    expect(calledWith).toContain('missing-MATH101')
    expect(calledWith).toContain('missing-PHYS101')
  })

  test('deduplicates missing IDs across conflicts', () => {
    const onAdd = vi.fn()
    // Both subjects missing the same prerequisite
    const conflicts: ConflictDetail[] = [
      { subject_id: 's1', subject_name: 'A', missing: [{ id: 'shared', name: 'Shared', code: 'SH101', type: 'hard' }] },
      { subject_id: 's2', subject_name: 'B', missing: [{ id: 'shared', name: 'Shared', code: 'SH101', type: 'hard' }] },
    ]
    render(<ConflictWarningBanner conflicts={conflicts} onAddMissing={onAdd} />)
    fireEvent.click(screen.getByText(/Add 1 missing prerequisite/))
    const calledWith: string[] = onAdd.mock.calls[0][0]
    expect(calledWith).toHaveLength(1)
    expect(calledWith[0]).toBe('shared')
  })

  test('hides add button when onAddMissing not provided', () => {
    const conflicts = [makeConflict('s1', 'Calc II', 'MATH101')]
    render(<ConflictWarningBanner conflicts={conflicts} />)
    expect(screen.queryByText(/Add/)).toBeNull()
  })
})
