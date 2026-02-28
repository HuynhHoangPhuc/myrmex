import { useEffect, useState } from 'react'
import { SidebarNav } from '@/components/layouts/sidebar-nav'
import { TopBar } from '@/components/layouts/top-bar'
import { MobileSidebarDrawer } from '@/components/layouts/mobile-sidebar-drawer'
import { CommandPalette } from '@/components/shared/command-palette'
import { Toaster } from '@/components/ui/toaster'
import { ChatPanel } from '@/chat/components/chat-panel'

interface AppLayoutProps {
  children: React.ReactNode
}

// Main authenticated shell: responsive nav, top bar, scrollable content area, global command palette
export function AppLayout({ children }: AppLayoutProps) {
  const [chatOpen, setChatOpen] = useState(false)
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const [commandOpen, setCommandOpen] = useState(false)

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        setCommandOpen((open) => !open)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <MobileSidebarDrawer open={mobileNavOpen} onOpenChange={setMobileNavOpen} />

      <aside className="hidden w-60 shrink-0 overflow-y-auto bg-sidebar-background md:flex md:flex-col">
        <SidebarNav />
      </aside>

      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar
          chatOpen={chatOpen}
          onToggleChat={() => setChatOpen((open) => !open)}
          onOpenMobileNav={() => setMobileNavOpen(true)}
        />
        <main className="flex-1 overflow-y-auto p-4 md:p-6">
          <div className="animate-in fade-in duration-200">{children}</div>
        </main>
      </div>

      <Toaster />
      <CommandPalette open={commandOpen} onOpenChange={setCommandOpen} />
      <ChatPanel isOpen={chatOpen} onClose={() => setChatOpen(false)} />
    </div>
  )
}
