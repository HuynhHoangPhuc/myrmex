import { useState } from 'react'
import { ChevronDown, ChevronRight, Wrench, Zap } from 'lucide-react'
import { cn } from '@/lib/utils/cn'
import type { ChatMessage as ChatMessageType } from '@/chat/types'

interface ChatMessageProps {
  message: ChatMessageType
}

/**
 * Renders a single chat message bubble.
 * - user: right-aligned blue bubble
 * - assistant: left-aligned gray bubble
 * - tool_call: compact card showing tool name + args
 * - tool_result: collapsible JSON result card
 */
export function ChatMessage({ message }: ChatMessageProps) {
  switch (message.role) {
    case 'user':
      return <UserBubble content={message.content} />
    case 'assistant':
      return <AssistantBubble content={message.content} />
    case 'tool_call':
      return <ToolCallCard toolName={message.toolName} args={message.toolArgs} />
    case 'tool_result':
      return <ToolResultCard toolName={message.toolName} result={message.toolResult} />
    default:
      return null
  }
}

// --- Sub-components ---

function UserBubble({ content }: { content: string }) {
  return (
    <div className="flex justify-end">
      <div className="max-w-[80%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm text-primary-foreground">
        <p className="whitespace-pre-wrap break-words">{content}</p>
      </div>
    </div>
  )
}

function AssistantBubble({ content }: { content: string }) {
  if (!content) {
    // Typing indicator while streaming
    return (
      <div className="flex justify-start">
        <div className="rounded-2xl rounded-bl-sm bg-muted px-4 py-3">
          <TypingIndicator />
        </div>
      </div>
    )
  }
  return (
    <div className="flex justify-start">
      <div className="max-w-[80%] rounded-2xl rounded-bl-sm bg-muted px-4 py-2.5 text-sm text-foreground">
        <p className="whitespace-pre-wrap break-words">{content}</p>
      </div>
    </div>
  )
}

function ToolCallCard({
  toolName,
  args,
}: {
  toolName?: string
  args?: Record<string, unknown>
}) {
  return (
    <div className="flex justify-start">
      <div className="flex max-w-[85%] items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs">
        <Wrench className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-600" />
        <div className="min-w-0">
          <span className="font-medium text-amber-800">{toolName ?? 'tool'}</span>
          {args && Object.keys(args).length > 0 && (
            <p className="mt-0.5 truncate text-amber-700">
              {Object.entries(args)
                .map(([k, v]) => `${k}: ${String(v)}`)
                .join(', ')}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

function ToolResultCard({
  toolName,
  result,
}: {
  toolName?: string
  result?: string
}) {
  const [expanded, setExpanded] = useState(false)

  const preview = formatResultPreview(result)

  return (
    <div className="flex justify-start">
      <div className="max-w-[85%] rounded-lg border border-green-200 bg-green-50 text-xs">
        <button
          className="flex w-full items-center gap-2 px-3 py-2 text-left"
          onClick={() => setExpanded((v) => !v)}
        >
          <Zap className="h-3.5 w-3.5 shrink-0 text-green-600" />
          <span className="font-medium text-green-800">{toolName ?? 'result'}</span>
          <span className="ml-1 truncate text-green-700">{preview}</span>
          {expanded ? (
            <ChevronDown className="ml-auto h-3 w-3 shrink-0 text-green-600" />
          ) : (
            <ChevronRight className="ml-auto h-3 w-3 shrink-0 text-green-600" />
          )}
        </button>
        {expanded && result && (
          <pre className="max-h-48 overflow-auto border-t border-green-200 px-3 py-2 text-green-900">
            {formatJSON(result)}
          </pre>
        )}
      </div>
    </div>
  )
}

function TypingIndicator() {
  return (
    <div className="flex gap-1">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className={cn(
            'h-1.5 w-1.5 rounded-full bg-muted-foreground/50',
            'animate-bounce',
          )}
          style={{ animationDelay: `${i * 150}ms` }}
        />
      ))}
    </div>
  )
}

// --- Helpers ---

function formatResultPreview(result?: string): string {
  if (!result) return ''
  try {
    const parsed = JSON.parse(result)
    if (Array.isArray(parsed)) return `${parsed.length} items`
    if (typeof parsed === 'object' && parsed !== null) {
      const keys = Object.keys(parsed)
      return keys.slice(0, 2).join(', ')
    }
  } catch {
    // not JSON
  }
  return result.slice(0, 40)
}

function formatJSON(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}
