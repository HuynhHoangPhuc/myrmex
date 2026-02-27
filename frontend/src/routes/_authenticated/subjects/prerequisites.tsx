import { createFileRoute } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { PrerequisiteDAG } from '@/modules/subject/components/prerequisite-dag'

export const Route = createFileRoute('/_authenticated/subjects/prerequisites')({
  component: PrerequisitesPage,
})

// Full interactive DAG visualization of all subjects and their prerequisite relationships
function PrerequisitesPage() {
  return (
    <div>
      <PageHeader
        title="Prerequisite Graph"
        description="Interactive DAG of all subject prerequisite relationships. Click a node to view the subject. Hover to highlight its prerequisite chain."
      />
      <div className="rounded-lg border overflow-hidden">
        <PrerequisiteDAG />
      </div>
    </div>
  )
}
