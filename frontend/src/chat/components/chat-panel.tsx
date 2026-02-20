import { useEffect, useRef, useState } from 'react'
import { X, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ChatMessage } from '@/chat/components/chat-message'
import { ChatInput } from '@/chat/components/chat-input'
import { ChatToggleButton } from '@/chat/components/chat-toggle-button'
import { useChat } from '@/chat/hooks/use-chat'

/**
 * ChatPanel is the floating AI assistant widget.
 * Renders a toggle FAB (bottom-right) and a slide-up panel above it.
 * Mounts the WebSocket connection on first render (via useChat).
 */
export function ChatPanel() {
  const [isOpen, setIsOpen] = useState(false)
  const { messages, isConnected, isStreaming, sendMessage, clearMessages } = useChat()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to newest message
  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, isOpen])

  return (
    // Anchor container — fixed to bottom-right corner, above other overlays
    <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3">
      {/* Chat panel — visible only when open */}
      {isOpen && (
        <div className="flex h-[600px] w-96 flex-col overflow-hidden rounded-xl border bg-background shadow-2xl">
          {/* Header */}
          <div className="flex items-center justify-between border-b bg-background px-4 py-3">
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold">Myrmex AI Assistant</span>
              {isStreaming && (
                <span className="text-xs text-muted-foreground animate-pulse">
                  thinking…
                </span>
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
                onClick={() => setIsOpen(false)}
                title="Close"
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Messages area */}
          <div className="flex-1 overflow-y-auto p-4 space-y-3">
            {messages.length === 0 ? (
              <WelcomePrompt />
            ) : (
              messages.map((msg) => (
                <ChatMessage key={msg.id} message={msg} />
              ))
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
      )}

      {/* Floating toggle button */}
      <ChatToggleButton
        onClick={() => setIsOpen((v) => !v)}
        isOpen={isOpen}
        isConnected={isConnected}
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
