// Enrollment history table showing subject, semester, and status per enrollment.
import { Badge } from '@/components/ui/badge'
import type { EnrollmentRequest } from '@/modules/student/types'

const STATUS_VARIANT = {
  pending: 'secondary',
  approved: 'default',
  rejected: 'destructive',
  completed: 'outline',
} as const

interface EnrollmentHistoryTableProps {
  enrollments: EnrollmentRequest[]
  subjectMap: Map<string, string>
  semesterMap: Map<string, string>
}

export function EnrollmentHistoryTable({
  enrollments,
  subjectMap,
  semesterMap,
}: EnrollmentHistoryTableProps) {
  if (enrollments.length === 0) return null

  return (
    <div className="space-y-2">
      <h2 className="text-base font-semibold">My Enrollment History</h2>
      <div className="rounded-md border">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/50">
            <tr>
              <th className="px-4 py-2 text-left font-medium">Subject</th>
              <th className="px-4 py-2 text-left font-medium">Semester</th>
              <th className="px-4 py-2 text-left font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {enrollments.map((e) => (
              <tr key={e.id} className="border-b last:border-0">
                <td className="px-4 py-2">
                  {subjectMap.get(e.subject_id) ?? e.subject_id.slice(0, 8)}
                </td>
                <td className="px-4 py-2 text-muted-foreground">
                  {semesterMap.get(e.semester_id) ?? e.semester_id.slice(0, 8)}
                </td>
                <td className="px-4 py-2">
                  <Badge variant={STATUS_VARIANT[e.status] ?? 'outline'}>{e.status}</Badge>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
