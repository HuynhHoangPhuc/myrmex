import { SidebarNav } from '@/components/layouts/sidebar-nav'
import { TopBar } from '@/components/layouts/top-bar'
import { Toaster } from '@/components/ui/toaster'
import { ChatPanel } from '@/chat/components/chat-panel'

interface AppLayoutProps {
  children: React.ReactNode
}

// Main authenticated shell: dark sidebar on left, top bar, scrollable content area
// Responsive: sidebar hides on mobile (future: add hamburger toggle)
export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <aside className="hidden w-60 shrink-0 overflow-y-auto bg-sidebar-background md:flex md:flex-col">
        <SidebarNav />
      </aside>

      {/* Main area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar />
        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>

      <Toaster />
      <ChatPanel />
    </div>
  )
}
