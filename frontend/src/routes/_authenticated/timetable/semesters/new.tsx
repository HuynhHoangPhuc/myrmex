import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/shared/page-header'
import { Stepper } from '@/components/shared/stepper'
import { GenerationPanel } from '@/modules/timetable/components/generation-panel'
import { SemesterForm } from '@/modules/timetable/components/semester-form'
import { SemesterWizardStepOfferings } from '@/modules/timetable/components/semester-wizard-step-offerings'
import { SemesterWizardStepSlots } from '@/modules/timetable/components/semester-wizard-step-slots'
import { useCreateSemester } from '@/modules/timetable/hooks/use-semesters'
import { useToast } from '@/lib/hooks/use-toast'

const searchSchema = z.object({
  step: z.number().min(1).max(4).catch(1),
  semesterId: z.string().optional().catch(undefined),
})

const STEPS = [
  { label: 'Create' },
  { label: 'Slots & rooms' },
  { label: 'Offerings' },
  { label: 'Generate' },
]

export const Route = createFileRoute('/_authenticated/timetable/semesters/new')({
  validateSearch: (search) => searchSchema.parse(search),
  component: NewSemesterPage,
})

function NewSemesterPage() {
  const { step, semesterId } = Route.useSearch()
  const navigate = Route.useNavigate()
  const { toast } = useToast()
  const createMutation = useCreateSemester()
  const [createdSemesterId, setCreatedSemesterId] = React.useState(semesterId ?? '')

  React.useEffect(() => {
    if (semesterId) setCreatedSemesterId(semesterId)
  }, [semesterId])

  const activeSemesterId = createdSemesterId || semesterId || ''
  const canContinue = step === 1 || Boolean(activeSemesterId)

  function goToStep(nextStep: number, nextSemesterId = activeSemesterId) {
    void navigate({
      search: {
        step: nextStep,
        semesterId: nextSemesterId || undefined,
      },
      replace: true,
    })
  }

  function renderStep() {
    if (!canContinue) {
      return (
        <div className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          Create a semester first to unlock the remaining setup steps.
        </div>
      )
    }

    switch (step) {
      case 1:
        return (
          <SemesterForm
            isLoading={createMutation.isPending}
            onSubmit={(data) => {
              createMutation.mutate(data, {
                onSuccess: (semester) => {
                  setCreatedSemesterId(semester.id)
                  toast({ title: 'Semester created', description: semester.name })
                  goToStep(2, semester.id)
                },
                onError: () => {
                  toast({ title: 'Failed to create semester', variant: 'destructive' })
                },
              })
            }}
          />
        )
      case 2:
        return <SemesterWizardStepSlots semesterId={activeSemesterId} />
      case 3:
        return (
          <SemesterWizardStepOfferings
            semesterId={activeSemesterId}
            onComplete={() => goToStep(4)}
          />
        )
      case 4:
        return <GenerationPanel semesterId={activeSemesterId} />
      default:
        return null
    }
  }

  return (
    <div className="max-w-4xl space-y-6">
      <PageHeader
        title="Semester Setup Wizard"
        description="Create a semester, review configuration, choose offerings, then generate a schedule without leaving this flow."
        actions={
          <Button variant="outline" asChild>
            <Link to="/timetable/semesters" search={{ page: 1, pageSize: 25 }}>
              Back to semesters
            </Link>
          </Button>
        }
      />

      <Stepper
        steps={STEPS}
        currentStep={step}
        onStepClick={(targetStep) => {
          if (targetStep <= step && (targetStep === 1 || activeSemesterId)) {
            goToStep(targetStep)
          }
        }}
      />

      {renderStep()}

      <div className="flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between">
        <Button
          type="button"
          variant="outline"
          onClick={() => goToStep(Math.max(1, step - 1))}
          disabled={step === 1}
        >
          Back
        </Button>

        {step > 1 && step < 4 && activeSemesterId && (
          <Button type="button" onClick={() => goToStep(step + 1)}>
            Next
          </Button>
        )}
      </div>
    </div>
  )
}
