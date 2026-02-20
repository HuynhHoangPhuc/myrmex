import { createFileRoute } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { PrerequisiteGraph } from '@/modules/subject/components/prerequisite-graph'

export const Route = createFileRoute('/_authenticated/subjects/prerequisites')({
  component: PrerequisitesPage,
})

// Full DAG visualization of all subjects and their prerequisite relationships
function PrerequisitesPage() {
  return (
    <div>
      <PageHeader
        title="Prerequisite Graph"
        description="Visual overview of all subject prerequisite relationships. Subjects are arranged by dependency layer."
      />
      <div className="rounded-lg border overflow-hidden">
        <PrerequisiteGraph />
      </div>
    </div>
  )
}
