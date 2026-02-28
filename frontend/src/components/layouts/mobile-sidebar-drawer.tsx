import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { SidebarNav } from '@/components/layouts/sidebar-nav'

interface MobileSidebarDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function MobileSidebarDrawer({ open, onOpenChange }: MobileSidebarDrawerProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="left-0 top-0 h-screen w-60 max-w-[85vw] translate-x-0 translate-y-0 gap-0 rounded-none border-r border-l-0 bg-sidebar-background p-0 sm:rounded-none [&>button]:right-3 [&>button]:top-3 [&>button]:text-sidebar-foreground">
        <DialogTitle className="sr-only">Navigation</DialogTitle>
        <SidebarNav onNavigate={() => onOpenChange(false)} />
      </DialogContent>
    </Dialog>
  )
}
