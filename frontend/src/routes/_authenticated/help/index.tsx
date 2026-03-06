import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { PageHeader } from '@/components/shared/page-header'
import { OverviewGuide, AdminGuide, TeacherGuide, StudentGuide, DeptHeadGuide } from './components/help-role-guides'

export const Route = createFileRoute('/_authenticated/help/')({
  component: HelpPage,
})

type GuideTab = 'overview' | 'admin' | 'teacher' | 'student' | 'depthead'

const TABS: { id: GuideTab; label: string }[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'admin', label: 'Admin' },
  { id: 'teacher', label: 'Teacher' },
  { id: 'student', label: 'Student' },
  { id: 'depthead', label: 'Dept Head' },
]

function HelpPage() {
  const [activeTab, setActiveTab] = React.useState<GuideTab>('overview')

  return (
    <div className="space-y-6">
      <PageHeader title="Help & User Guide" description="Documentation for all roles in Myrmex ERP" />

      {/* Tab bar */}
      <div className="flex gap-1 border-b">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={[
              'px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px',
              activeTab === tab.id
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground',
            ].join(' ')}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Guide content */}
      <div className="max-w-3xl">
        {activeTab === 'overview' && <OverviewGuide />}
        {activeTab === 'admin' && <AdminGuide />}
        {activeTab === 'teacher' && <TeacherGuide />}
        {activeTab === 'student' && <StudentGuide />}
        {activeTab === 'depthead' && <DeptHeadGuide />}
      </div>
    </div>
  )
}
