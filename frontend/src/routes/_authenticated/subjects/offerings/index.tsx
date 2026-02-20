import { createFileRoute } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { OfferingManager } from '@/modules/subject/components/offering-manager'

export const Route = createFileRoute('/_authenticated/subjects/offerings/')({
  component: OfferingsPage,
})

function OfferingsPage() {
  return (
    <div className="max-w-3xl">
      <PageHeader
        title="Semester Offerings"
        description="Select which subjects are offered in each semester."
      />
      <OfferingManager />
    </div>
  )
}
