// Semester auto-select and client-side filter logic for enrollment queue.
import * as React from 'react'
import type { EnrollmentRequest } from '@/modules/student/types'
import type { Semester } from '@/modules/timetable/types'
import type { Subject } from '@/modules/subject/types'
import type { Student } from '@/modules/student/types'

interface UseEnrollmentFiltersOptions {
  semesters: Semester[] | undefined
  allSubjects: Subject[] | undefined
  allStudents: { data: Student[] } | undefined
  data: { data: EnrollmentRequest[] } | undefined
  semesterId: string | undefined
  subjectId: string | undefined
  search: string | undefined
}

interface UseEnrollmentFiltersResult {
  latestSemester: Semester | undefined
  effectiveSemesterId: string | undefined
  studentMap: Map<string, string>
  subjectMap: Map<string, string>
  semesterMap: Map<string, string>
  filteredData: EnrollmentRequest[]
}

export function useEnrollmentFilters({
  semesters,
  allSubjects,
  allStudents,
  data,
  semesterId,
  subjectId,
  search,
}: UseEnrollmentFiltersOptions): UseEnrollmentFiltersResult {
  // Auto-select latest (active) semester only when URL has no semesterId at all (first visit).
  const latestSemester = React.useMemo(() => {
    if (!semesters?.length) return undefined
    return (
      semesters.find((s) => s.is_active) ??
      [...semesters].sort((a, b) => b.start_date.localeCompare(a.start_date))[0]
    )
  }, [semesters])

  // 'all' sentinel means no semester filter; any real UUID is passed through
  const effectiveSemesterId = semesterId === 'all' ? undefined : semesterId

  const studentMap = React.useMemo(
    () => new Map((allStudents?.data ?? []).map((s) => [s.id, s.full_name])),
    [allStudents],
  )
  const subjectMap = React.useMemo(
    () => new Map((allSubjects ?? []).map((s) => [s.id, `${s.code} — ${s.name}`])),
    [allSubjects],
  )
  const semesterMap = React.useMemo(
    () => new Map((semesters ?? []).map((s) => [s.id, s.name])),
    [semesters],
  )

  // Client-side filter for subject and name search (semester is handled server-side)
  const filteredData = React.useMemo(() => {
    let rows = data?.data ?? []
    if (subjectId) rows = rows.filter((e) => e.subject_id === subjectId)
    if (search) {
      const q = search.toLowerCase()
      rows = rows.filter((e) => (studentMap.get(e.student_id) ?? '').toLowerCase().includes(q))
    }
    return rows
  }, [data, subjectId, search, studentMap])

  return { latestSemester, effectiveSemesterId, studentMap, subjectMap, semesterMap, filteredData }
}
