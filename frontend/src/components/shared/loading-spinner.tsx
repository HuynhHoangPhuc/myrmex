import { cn } from '@/lib/utils/cn'

interface LoadingSpinnerProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizeMap = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-10 w-10' }

// Simple CSS-based spinner — no extra dependencies
export function LoadingSpinner({ className, size = 'md' }: LoadingSpinnerProps) {
  return (
    <div
      role="status"
      aria-label="Loading"
      className={cn(
        'animate-spin rounded-full border-2 border-current border-t-transparent text-primary',
        sizeMap[size],
        className,
      )}
    />
  )
}

// Full-page centered overlay
export function LoadingPage() {
  return (
    <div className="flex h-full min-h-[400px] w-full items-center justify-center">
      <LoadingSpinner size="lg" />
    </div>
  )
}
