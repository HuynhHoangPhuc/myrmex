// Grade assignment handler with validation and toast feedback.
import * as React from 'react'
import { useAssignGrade, getLetterGrade } from '@/modules/student/hooks/use-grades'
import { authStore } from '@/lib/stores/auth-store'
import { toast } from '@/lib/hooks/use-toast'
import type { EnrollmentRequest } from '@/modules/student/types'

interface UseGradeAssignmentResult {
  gradeValue: string
  notes: string
  preview: string | null
  isPending: boolean
  setGradeValue: (v: string) => void
  setNotes: (v: string) => void
  handleAssign: (enrollment: EnrollmentRequest | null, onSuccess: () => void) => void
  resetForm: () => void
}

export function useGradeAssignment(): UseGradeAssignmentResult {
  const [gradeValue, setGradeValue] = React.useState('')
  const [notes, setNotes] = React.useState('')
  const assignMutation = useAssignGrade()

  const preview = gradeValue !== '' ? getLetterGrade(Number(gradeValue)) : null

  function resetForm() {
    setGradeValue('')
    setNotes('')
  }

  function handleAssign(enrollment: EnrollmentRequest | null, onSuccess: () => void) {
    if (!enrollment || gradeValue === '') return
    const user = authStore.getUser()
    assignMutation.mutate(
      {
        enrollment_id: enrollment.id,
        grade_numeric: Number(gradeValue),
        graded_by: user?.id ?? '',
        notes: notes || undefined,
      },
      {
        onSuccess: () => {
          toast({ title: 'Grade assigned', description: `${gradeValue} → ${preview}` })
          resetForm()
          onSuccess()
        },
        onError: () => toast({ title: 'Failed to assign grade', variant: 'destructive' }),
      },
    )
  }

  return {
    gradeValue,
    notes,
    preview,
    isPending: assignMutation.isPending,
    setGradeValue,
    setNotes,
    handleAssign,
    resetForm,
  }
}
