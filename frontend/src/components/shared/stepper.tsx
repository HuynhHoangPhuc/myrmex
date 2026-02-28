import { Check } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

interface StepperStep {
  label: string
}

interface StepperProps {
  steps: StepperStep[]
  currentStep: number
  onStepClick?: (step: number) => void
}

export function Stepper({ steps, currentStep, onStepClick }: StepperProps) {
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-lg border bg-card p-3 sm:gap-3">
      {steps.map((step, index) => {
        const stepNumber = index + 1
        const isComplete = stepNumber < currentStep
        const isActive = stepNumber === currentStep
        const isClickable = Boolean(onStepClick) && stepNumber <= currentStep

        return (
          <div key={step.label} className="flex items-center gap-2">
            <button
              type="button"
              className={cn(
                'flex h-9 w-9 items-center justify-center rounded-full border text-sm font-semibold transition-colors',
                isActive && 'border-primary bg-primary text-primary-foreground',
                isComplete && 'border-primary/40 bg-primary/10 text-primary',
                !isActive && !isComplete && 'border-border bg-background text-muted-foreground',
                isClickable ? 'cursor-pointer' : 'cursor-default',
              )}
              onClick={() => isClickable && onStepClick?.(stepNumber)}
              disabled={!isClickable}
              aria-current={isActive ? 'step' : undefined}
            >
              {isComplete ? <Check className="h-4 w-4" /> : stepNumber}
            </button>
            <span className={cn('hidden text-sm sm:inline', isActive ? 'font-medium text-foreground' : 'text-muted-foreground')}>
              {step.label}
            </span>
            {index < steps.length - 1 && <div className="hidden h-px w-6 bg-border sm:block" />}
          </div>
        )
      })}
    </div>
  )
}
