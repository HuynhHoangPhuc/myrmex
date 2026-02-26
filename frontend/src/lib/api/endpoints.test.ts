import { ENDPOINTS } from './endpoints'

describe('ENDPOINTS', () => {
  it('has correct auth paths', () => {
    expect(ENDPOINTS.auth.login).toBe('/auth/login')
    expect(ENDPOINTS.auth.register).toBe('/auth/register')
    expect(ENDPOINTS.auth.me).toBe('/auth/me')
  })

  it('builds parameterized HR paths', () => {
    expect(ENDPOINTS.hr.teacher('abc-123')).toBe('/hr/teachers/abc-123')
    expect(ENDPOINTS.hr.department('dept-1')).toBe('/hr/departments/dept-1')
  })

  it('builds parameterized subject paths', () => {
    expect(ENDPOINTS.subjects.detail('sub-1')).toBe('/subjects/sub-1')
    expect(ENDPOINTS.subjects.prerequisites('sub-1')).toBe('/subjects/sub-1/prerequisites')
  })

  it('builds parameterized timetable paths', () => {
    expect(ENDPOINTS.timetable.semester('sem-1')).toBe('/timetable/semesters/sem-1')
    expect(ENDPOINTS.timetable.generate('sem-1')).toBe('/timetable/semesters/sem-1/generate')
    expect(ENDPOINTS.timetable.manualAssign('sch-1', 'ent-1')).toBe(
      '/timetable/schedules/sch-1/entries/ent-1',
    )
  })
})
