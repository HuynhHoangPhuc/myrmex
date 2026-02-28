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
 * Supports expand-to-fullscreen toggle and mobile full-width layout.
 */
export function ChatPanel({ isOpen, onClose }: ChatPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const { messages, isConnected, isStreaming, isWaiting, sendMessage, clearMessages } = useChat()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, isOpen])

  useEffect(() => {
    if (!isOpen) setIsExpanded(false)
  }, [isOpen])

  if (!isOpen) return null

  return (
    <div
      className={cn(
        'fixed z-50 flex flex-col bg-background shadow-2xl sm:border-l',
        isExpanded ? 'inset-0' : 'inset-y-0 right-0 w-full sm:w-[380px]',
      )}
    >
      <div className="flex shrink-0 items-center justify-between border-b bg-background px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">Myrmex AI</span>
          {isStreaming && <span className="animate-pulse text-xs text-muted-foreground">thinking…</span>}
          {!isConnected && <span className="text-xs text-muted-foreground/60">offline</span>}
        </div>
        <div className="flex items-center gap-1">
          {messages.length > 0 && (
            <Button
              variant="ghost"
              size="icon"
              className="h-10 w-10 text-muted-foreground hover:text-destructive"
              onClick={clearMessages}
              title="Clear conversation"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="h-10 w-10 text-muted-foreground"
            onClick={() => setIsExpanded((expanded) => !expanded)}
            title={isExpanded ? 'Exit fullscreen' : 'Expand to fullscreen'}
          >
            {isExpanded ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-10 w-10 text-muted-foreground"
            onClick={onClose}
            title="Close"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="flex-1 space-y-3 overflow-y-auto p-4">
        {messages.length === 0 && !isWaiting ? (
          <WelcomePrompt />
        ) : (
          messages.map((message) => <ChatMessage key={message.id} message={message} />)
        )}
        {isWaiting && (
          <div className="flex justify-start">
            <div className="rounded-2xl rounded-bl-sm bg-muted px-4 py-3">
              <div className="flex gap-1">
                {[0, 1, 2].map((index) => (
                  <span
                    key={index}
                    className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground/50"
                    style={{ animationDelay: `${index * 150}ms` }}
                  />
                ))}
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

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
