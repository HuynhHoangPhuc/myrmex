// Column definitions for the grade entry data table.
import { Button } from '@/components/ui/button'
import type { ColumnDef } from '@tanstack/react-table'
import type { EnrollmentRequest } from '@/modules/student/types'

interface BuildGradeColumnsOptions {
  studentMap: Map<string, string>
  subjectMap: Map<string, string>
  onAssign: (enrollment: EnrollmentRequest) => void
}

export function buildGradeEntryColumns({
  studentMap,
  subjectMap,
  onAssign,
}: BuildGradeColumnsOptions): ColumnDef<EnrollmentRequest>[] {
  return [
    {
      accessorKey: 'student_id',
      header: 'Student',
      cell: ({ row }) => (
        <span className="text-sm font-medium">
          {studentMap.get(row.original.student_id) ?? row.original.student_id.slice(0, 8)}
        </span>
      ),
    },
    {
      accessorKey: 'subject_id',
      header: 'Subject',
      cell: ({ row }) => (
        <span className="text-sm">
          {subjectMap.get(row.original.subject_id) ?? row.original.subject_id.slice(0, 8)}
        </span>
      ),
    },
    {
      id: 'actions',
      header: 'Grade',
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="outline"
          className="h-7 text-xs"
          onClick={() => onAssign(row.original)}
        >
          Assign Grade
        </Button>
      ),
    },
  ]
}
