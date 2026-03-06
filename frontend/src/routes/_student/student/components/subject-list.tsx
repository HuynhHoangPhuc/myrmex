// Renders the filterable subject card list with enroll buttons.
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { Subject } from '@/modules/subject/types'

interface SubjectListProps {
  subjects: Subject[]
  enrolledSubjectIds: Set<string>
  selectedSemesterId: string
  onEnroll: (subject: Subject) => void
}

export function SubjectList({
  subjects,
  enrolledSubjectIds,
  selectedSemesterId,
  onEnroll,
}: SubjectListProps) {
  if (subjects.length === 0) {
    return <p className="text-sm text-muted-foreground">No subjects found.</p>
  }

  return (
    <div className="space-y-2">
      {subjects.map((subject) => {
        const alreadyEnrolled = enrolledSubjectIds.has(subject.id)
        return (
          <div
            key={subject.id}
            className="flex items-center justify-between rounded-md border px-4 py-3 text-sm"
          >
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <span className="font-mono text-xs font-bold">{subject.code}</span>
                <span className="font-medium">{subject.name}</span>
                {subject.credits != null && (
                  <Badge variant="outline" className="text-xs">{subject.credits} cr</Badge>
                )}
              </div>
            </div>
            {alreadyEnrolled ? (
              <Badge variant="secondary">Enrolled</Badge>
            ) : (
              <Button
                size="sm"
                variant="outline"
                className="h-7 text-xs"
                disabled={!selectedSemesterId}
                onClick={() => onEnroll(subject)}
              >
                Enroll →
              </Button>
            )}
          </div>
        )
      })}
    </div>
  )
}
