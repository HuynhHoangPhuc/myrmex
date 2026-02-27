import { useState } from 'react'
import { SidebarNav } from '@/components/layouts/sidebar-nav'
import { TopBar } from '@/components/layouts/top-bar'
import { Toaster } from '@/components/ui/toaster'
import { ChatPanel } from '@/chat/components/chat-panel'

interface AppLayoutProps {
  children: React.ReactNode
}

// Main authenticated shell: dark sidebar on left, top bar, scrollable content area
// Chat panel state is lifted here so TopBar can render the toggle button
export function AppLayout({ children }: AppLayoutProps) {
  const [chatOpen, setChatOpen] = useState(false)

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <aside className="hidden w-60 shrink-0 overflow-y-auto bg-sidebar-background md:flex md:flex-col">
        <SidebarNav />
      </aside>

      {/* Main area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar chatOpen={chatOpen} onToggleChat={() => setChatOpen((v) => !v)} />
        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>

      <Toaster />
      <ChatPanel isOpen={chatOpen} onClose={() => setChatOpen(false)} />
    </div>
  )
}
