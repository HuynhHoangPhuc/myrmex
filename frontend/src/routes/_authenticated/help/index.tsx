import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PageHeader } from '@/components/shared/page-header'

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

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card className="mb-4">
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground space-y-2">{children}</CardContent>
    </Card>
  )
}

function Steps({ items }: { items: string[] }) {
  return (
    <ol className="list-decimal list-inside space-y-1">
      {items.map((item, i) => <li key={i}>{item}</li>)}
    </ol>
  )
}

function InfoTable({ rows }: { rows: [string, string][] }) {
  return (
    <table className="w-full text-sm border-collapse">
      <tbody>
        {rows.map(([label, desc]) => (
          <tr key={label} className="border-b last:border-0">
            <td className="py-1.5 pr-4 font-medium text-foreground whitespace-nowrap">{label}</td>
            <td className="py-1.5 text-muted-foreground">{desc}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

// ── Guide content components ──────────────────────────────────────────────────

function OverviewGuide() {
  return (
    <>
      <Section title="What is Myrmex?">
        <p>Myrmex is a university faculty management system covering HR, Subjects, Timetable scheduling, and Student records. Access is role-based.</p>
      </Section>
      <Section title="Roles">
        <InfoTable rows={[
          ['Super Admin', 'Full system access'],
          ['Admin', 'User management, bulk import, audit logs, role assignment'],
          ['Dept Head', 'Department teachers, subjects, semester setup, schedule generation'],
          ['Teacher', 'View schedule, set availability, enter grades'],
          ['Student', 'View schedule, enrolled subjects, transcript, enrollment requests'],
        ]} />
      </Section>
      <Section title="Login Methods">
        <InfoTable rows={[
          ['Email + Password', 'All roles'],
          ['Google OAuth', 'Teachers, Admins'],
          ['Microsoft OAuth', 'Students (invite-code registration)'],
        ]} />
      </Section>
      <Section title="Navigation Map">
        <InfoTable rows={[
          ['/dashboard', 'Home overview (all roles)'],
          ['/hr', 'Teachers & Departments (Admin, Dept Head)'],
          ['/subjects', 'Subject catalog, prerequisites, offerings'],
          ['/timetable', 'Semesters, schedules, generation'],
          ['/students', 'Student list, enrollments (Admin)'],
          ['/grades', 'Grade entry (Teacher)'],
          ['/analytics', 'Workload & utilization charts'],
          ['/notifications', 'In-app notification center'],
          ['/admin', 'Role management, audit logs, bulk import'],
        ]} />
      </Section>
    </>
  )
}

function AdminGuide() {
  return (
    <>
      <Section title="First Login & Password Change">
        <Steps items={[
          'Receive credentials (email + temporary password) from Super Admin.',
          'Enter email and temporary password → Sign In.',
          'You will be prompted to change your password immediately.',
          'Enter new password (min 8 chars, one uppercase, one digit) → Save.',
        ]} />
        <p className="mt-2 text-xs">Google OAuth admins: click Continue with Google instead.</p>
      </Section>
      <Section title="Bulk User Import (Admin → Bulk Import)">
        <Steps items={[
          'Click Download Template to get the CSV format for Teachers or Students tab.',
          'Fill in the CSV: Teachers need email, full_name, department_id, employee_id.',
          'Students need email, full_name, student_id, program, year.',
          'Click Choose File → select CSV → Upload.',
          'Review the summary: total rows, imported count, failed rows.',
          'Click Download Error Report to get failed rows with error details.',
          'Fix errors and re-upload only failed rows.',
        ]} />
        <p className="mt-2 text-xs">Max 500 rows per upload. Duplicate emails are skipped.</p>
      </Section>
      <Section title="User Management">
        <Steps items={[
          'Navigate to HR → Teachers or Students.',
          'Use the search bar to filter by name or email.',
          'Click a row → Edit to modify name, department, or ID → Save.',
        ]} />
      </Section>
      <Section title="Role Management (Admin → Role Management)">
        <Steps items={[
          'Search for the user by email or name.',
          'Select the desired role from the dropdown.',
          'Click Assign Role → confirm in the dialog.',
        ]} />
        <p className="mt-2 text-xs">Role changes take effect on next login. Admins can assign up to admin role only.</p>
      </Section>
      <Section title="Audit Logs (Admin → Audit Logs)">
        <p>All system mutations are recorded immutably. Filter by resource type, action, or user ID. Click any row to expand and see before/after JSON snapshots.</p>
      </Section>
      <Section title="Broadcast Announcements">
        <Steps items={[
          'Navigate to Notifications → Broadcast.',
          'Enter title and body text.',
          'Select target audience: All Users, Teachers only, or Students only.',
          'Click Send → confirm.',
        ]} />
      </Section>
    </>
  )
}

function TeacherGuide() {
  return (
    <>
      <Section title="Login">
        <p><strong>Email:</strong> Enter institutional email + password → Sign In.</p>
        <p><strong>Google OAuth:</strong> Click Continue with Google → select institutional account.</p>
        <p className="text-xs mt-1">If "Account not found" appears, contact your admin — your email must be pre-registered.</p>
      </Section>
      <Section title="Viewing Your Schedule (Timetable → Schedules)">
        <Steps items={[
          'Your sessions are shown filtered to your assignments by default.',
          'Use the Semester dropdown to switch between semesters.',
          'Click a session row to see full details including enrolled students.',
        ]} />
      </Section>
      <Section title="Setting Weekly Availability">
        <Steps items={[
          'Navigate to Timetable → Schedules → My Availability.',
          'Select the target semester.',
          'Click cells on the weekly grid to toggle Available (green) / Unavailable (grey).',
          'Click Save Preferences.',
        ]} />
        <p className="text-xs mt-1">Submit preferences before the generation deadline set by the Department Head.</p>
      </Section>
      <Section title="Entering Student Grades (Grades)">
        <Steps items={[
          'Select the Semester and Subject from dropdowns.',
          'Enter grade values (0–10 scale) in the Grade column.',
          'Optionally add a Note per student.',
          'Click Save Grades at the bottom.',
        ]} />
        <InfoTable rows={[
          ['Draft', 'Entered but not yet published'],
          ['Published', 'Visible to student'],
          ['Locked', 'Grading window closed; contact admin to unlock'],
        ]} />
      </Section>
      <Section title="AI Chat Assistant">
        <p>Click the Chat icon in the bottom-right corner. Ask questions in natural language about your schedule, enrollments, or system features. The assistant is read-only and cannot modify data.</p>
      </Section>
    </>
  )
}

function StudentGuide() {
  return (
    <>
      <Section title="Registration">
        <p><strong>Invite Code + Email:</strong></p>
        <Steps items={[
          'Receive invite code and temporary password from admin.',
          'Go to login page → Register with Invite Code.',
          'Enter invite code, email, temporary password → Activate Account.',
          'Set a new permanent password → Save.',
        ]} />
        <p className="mt-2"><strong>Microsoft OAuth:</strong></p>
        <Steps items={[
          'Click Continue with Microsoft → sign in with institutional account.',
          'If your email matches a pre-registered invite, your account activates automatically.',
        ]} />
      </Section>
      <Section title="Viewing Your Schedule (Timetable → Schedules)">
        <Steps items={[
          'Your enrolled sessions show for the current semester by default.',
          'Use the Semester dropdown for past or upcoming semesters.',
          'Click a row for full session details: teacher, room, time slot.',
        ]} />
      </Section>
      <Section title="Requesting Subject Enrollment">
        <Steps items={[
          'Navigate to Subjects → Offerings.',
          'Check the Prerequisites column — all must be passed before enrolling.',
          'Click Request Enrollment on the desired subject.',
          'You will receive an in-app notification when approved or rejected.',
        ]} />
        <InfoTable rows={[
          ['Pending', 'Request submitted, awaiting admin approval'],
          ['Enrolled', 'Approved and confirmed'],
          ['Rejected', 'Denied — check notification for reason'],
          ['Waitlisted', 'Subject full; you are in queue'],
        ]} />
      </Section>
      <Section title="Checking Prerequisite Status (Subjects → Prerequisites)">
        <p>Select a subject to see its prerequisite chain as a dependency graph. Your completed subjects are highlighted in green; missing ones in red.</p>
      </Section>
      <Section title="Viewing Transcript & GPA">
        <Steps items={[
          'Current GPA is shown on the Dashboard summary card.',
          'Full transcript: Students → your profile → Transcript tab.',
          'Click Export PDF to download your transcript.',
        ]} />
      </Section>
    </>
  )
}

function DeptHeadGuide() {
  return (
    <>
      <Section title="Dashboard Overview">
        <InfoTable rows={[
          ['Teachers', 'Total teachers in your department'],
          ['Subjects', 'Active subjects this semester'],
          ['Schedule Coverage', '% of offered subjects with assigned teachers'],
          ['Workload', 'Average teaching hours per teacher this semester'],
        ]} />
      </Section>
      <Section title="Managing Department Teachers (HR → Teachers)">
        <Steps items={[
          'The list is pre-filtered to your department.',
          'To add: click Assign Teacher → search by name/email → Assign.',
          'To remove: click teacher row → Edit → clear department field → Save.',
        ]} />
      </Section>
      <Section title="Subject Management (Subjects)">
        <Steps items={[
          'Create: click New Subject → fill code, name, credits, description, department → Save.',
          'Edit: click subject row → Edit → modify fields → Save.',
          'Prerequisites: navigate to Subjects → Prerequisites → select subject → Add Prerequisite.',
          'The system rejects prerequisites that would create a cycle (cycle detection is automatic).',
        ]} />
      </Section>
      <Section title="Semester Setup (Timetable → Semesters)">
        <Steps items={[
          'Create semester: New Semester → name, start/end dates, registration deadline → Save.',
          'Add rooms: open semester → Rooms tab → Add Room → code, building, capacity.',
          'Add time slots: Time Slots tab → Add Time Slot → day, start time, end time.',
          'Add offerings: Offerings tab → Add Offering → select subject, set enrollment cap.',
        ]} />
      </Section>
      <Section title="Schedule Generation (Timetable → Generate)">
        <Steps items={[
          'Select the target semester.',
          'Verify the pre-flight checklist is all green (rooms, slots, teachers, availability).',
          'Click Generate Schedule — solver runs in background.',
          'When complete, you are redirected to the schedule view.',
          'Sessions marked Unscheduled need manual assignment (solver timed out on those).',
        ]} />
      </Section>
      <Section title="Manual Schedule Adjustments (Timetable → Schedules)">
        <Steps items={[
          'Click a session row → Edit.',
          'Change room, time slot, or assigned teacher.',
          'System checks conflicts in real time (room double-booking, teacher overlap, capacity).',
          'Click Save if no conflicts shown.',
        ]} />
      </Section>
      <Section title="Analytics">
        <InfoTable rows={[
          ['Workload Distribution', 'Teaching hours per teacher; highlights over/under-loaded'],
          ['Room Utilization', '% of time slots used per room'],
          ['Enrollment Trends', 'Subject enrollment counts over semesters'],
          ['Prerequisite Completion', '% of students completing prerequisites on time'],
        ]} />
        <p className="mt-2 text-xs">Use the Semester filter to compare across terms. Click Export CSV on any report.</p>
      </Section>
    </>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

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
