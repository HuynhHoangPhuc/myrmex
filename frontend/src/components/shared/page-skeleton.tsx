import { cn } from '@/lib/utils/cn'

interface PageSkeletonProps {
  variant?: 'cards' | 'detail' | 'table'
  className?: string
}

function SkeletonBlock({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-md bg-muted', className)} />
}

export function PageSkeleton({ variant = 'detail', className }: PageSkeletonProps) {
  if (variant === 'cards') {
    return (
      <div className={cn('space-y-6', className)}>
        <div className="space-y-2">
          <SkeletonBlock className="h-8 w-48" />
          <SkeletonBlock className="h-4 w-80 max-w-full" />
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="rounded-lg border p-5">
              <SkeletonBlock className="h-4 w-24" />
              <SkeletonBlock className="mt-4 h-8 w-16" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (variant === 'table') {
    return (
      <div className={cn('space-y-4', className)}>
        <SkeletonBlock className="h-8 w-40" />
        <div className="rounded-lg border p-4">
          <SkeletonBlock className="h-10 w-full" />
          <div className="mt-4 space-y-3">
            {Array.from({ length: 5 }).map((_, index) => (
              <SkeletonBlock key={index} className="h-12 w-full" />
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={cn('space-y-6', className)}>
      <div className="space-y-2">
        <SkeletonBlock className="h-8 w-56" />
        <SkeletonBlock className="h-4 w-96 max-w-full" />
      </div>
      <div className="space-y-4 rounded-lg border p-5">
        <SkeletonBlock className="h-5 w-28" />
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, index) => (
            <SkeletonBlock key={index} className="h-11 w-full" />
          ))}
        </div>
      </div>
      <div className="space-y-4 rounded-lg border p-5">
        <SkeletonBlock className="h-5 w-20" />
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, index) => (
            <SkeletonBlock key={index} className="h-11 w-full" />
          ))}
        </div>
      </div>
    </div>
  )
}
