import { useEffect, useRef, useState } from 'react'
import { X, Trash2, Maximize2, Minimize2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ChatMessage } from '@/chat/components/chat-message'
import { ChatInput } from '@/chat/components/chat-input'
import { useChat } from '@/chat/hooks/use-chat'
import { cn } from '@/lib/utils/cn'

interface ChatPanelProps {
  isOpen: boolean
  onClose: () => void
}

/**
 * Right-side AI chat panel.
 * Controlled from AppLayout — parent manages open/close state.
 * Supports expand-to-fullscreen toggle.
 */
export function ChatPanel({ isOpen, onClose }: ChatPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const { messages, isConnected, isStreaming, isWaiting, sendMessage, clearMessages } = useChat()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to newest message
  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, isOpen])

  // Reset expand state on close
  useEffect(() => {
    if (!isOpen) setIsExpanded(false)
  }, [isOpen])

  if (!isOpen) return null

  return (
    <div
      className={cn(
        'fixed z-50 flex flex-col border-l bg-background shadow-2xl',
        isExpanded ? 'inset-0' : 'right-0 top-0 h-screen w-[380px]',
      )}
    >
      {/* Header */}
      <div className="flex shrink-0 items-center justify-between border-b bg-background px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">Myrmex AI</span>
          {isStreaming && (
            <span className="animate-pulse text-xs text-muted-foreground">thinking…</span>
          )}
          {!isConnected && (
            <span className="text-xs text-muted-foreground/60">offline</span>
          )}
        </div>
        <div className="flex items-center gap-1">
          {messages.length > 0 && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 text-muted-foreground hover:text-destructive"
              onClick={clearMessages}
              title="Clear conversation"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground"
            onClick={() => setIsExpanded((v) => !v)}
            title={isExpanded ? 'Exit fullscreen' : 'Expand to fullscreen'}
          >
            {isExpanded ? (
              <Minimize2 className="h-3.5 w-3.5" />
            ) : (
              <Maximize2 className="h-3.5 w-3.5" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground"
            onClick={onClose}
            title="Close"
          >
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Messages area */}
      <div className="flex-1 space-y-3 overflow-y-auto p-4">
        {messages.length === 0 && !isWaiting ? (
          <WelcomePrompt />
        ) : (
          messages.map((msg) => <ChatMessage key={msg.id} message={msg} />)
        )}
        {/* Typing dots: visible from sendMessage until first text chunk arrives */}
        {isWaiting && (
          <div className="flex justify-start">
            <div className="rounded-2xl rounded-bl-sm bg-muted px-4 py-3">
              <div className="flex gap-1">
                {[0, 1, 2].map((i) => (
                  <span
                    key={i}
                    className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground/50"
                    style={{ animationDelay: `${i * 150}ms` }}
                  />
                ))}
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <ChatInput
        onSend={sendMessage}
        disabled={isStreaming || !isConnected}
        placeholder={
          !isConnected
            ? 'Connecting…'
            : isStreaming
              ? 'Waiting for response…'
              : 'Ask about teachers, subjects, or schedules…'
        }
      />
    </div>
  )
}

/** Empty-state prompt shown before the first message */
function WelcomePrompt() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 text-center text-muted-foreground">
      <p className="text-sm font-medium">How can I help you today?</p>
      <ul className="space-y-1 text-xs">
        <li>"Show all math teachers"</li>
        <li>"List subjects with 3 credits"</li>
        <li>"Generate schedule for Fall 2026"</li>
      </ul>
    </div>
  )
}
