import { Bot } from 'lucide-react'
import { cn } from '@/lib/utils/cn'
import { Button } from '@/components/ui/button'

interface ChatToggleButtonProps {
  onClick: () => void
  isOpen: boolean
  isConnected: boolean
}

/**
 * Floating action button that opens/closes the chat panel.
 * Shows a green dot indicator when WebSocket is connected.
 */
export function ChatToggleButton({ onClick, isOpen, isConnected }: ChatToggleButtonProps) {
  return (
    <div className="relative">
      <Button
        onClick={onClick}
        size="icon"
        className={cn(
          'h-12 w-12 rounded-full shadow-lg transition-all',
          isOpen && 'ring-2 ring-primary ring-offset-2',
        )}
        aria-label={isOpen ? 'Close AI assistant' : 'Open AI assistant'}
      >
        <Bot className="h-5 w-5" />
      </Button>

      {/* Connection status indicator dot */}
      <span
        className={cn(
          'absolute right-0.5 top-0.5 h-2.5 w-2.5 rounded-full border-2 border-background',
          isConnected ? 'bg-green-500' : 'bg-muted-foreground/40',
        )}
        title={isConnected ? 'Connected' : 'Disconnected'}
      />
    </div>
  )
}
