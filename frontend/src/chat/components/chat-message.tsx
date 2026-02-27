import { CheckCircle2, Wrench } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { cn } from '@/lib/utils/cn'
import type { ChatMessage as ChatMessageType } from '@/chat/types'

interface ChatMessageProps {
  message: ChatMessageType
}

/**
 * Renders a single chat message.
 * - user: right-aligned blue bubble
 * - assistant: left-aligned gray bubble with Markdown
 * - tool_call: compact card with friendly action label
 * - tool_result: minimal success badge (no JSON)
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
      return <ToolResultCard result={message.toolResult} />
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
        <ReactMarkdown
          components={{
            p: ({ children }) => <p className="mb-2 last:mb-0 break-words">{children}</p>,
            ul: ({ children }) => <ul className="mb-2 ml-4 list-disc space-y-1 last:mb-0">{children}</ul>,
            ol: ({ children }) => <ol className="mb-2 ml-4 list-decimal space-y-1 last:mb-0">{children}</ol>,
            li: ({ children }) => <li className="break-words">{children}</li>,
            strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
            em: ({ children }) => <em className="italic">{children}</em>,
            h1: ({ children }) => <h1 className="mb-1 text-base font-bold">{children}</h1>,
            h2: ({ children }) => <h2 className="mb-1 text-sm font-bold">{children}</h2>,
            h3: ({ children }) => <h3 className="mb-1 text-sm font-semibold">{children}</h3>,
            code: ({ children, className }) => {
              const isBlock = className?.includes('language-')
              return isBlock ? (
                <code className="block overflow-x-auto rounded bg-background/60 px-3 py-2 font-mono text-xs">
                  {children}
                </code>
              ) : (
                <code className="rounded bg-background/60 px-1 py-0.5 font-mono text-xs">{children}</code>
              )
            },
            pre: ({ children }) => <pre className="mb-2 last:mb-0">{children}</pre>,
            hr: () => <hr className="my-2 border-border" />,
          }}
        >
          {content}
        </ReactMarkdown>
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
  const label = friendlyToolLabel(toolName)
  const paramSummary = args ? friendlyArgs(args) : null

  return (
    <div className="flex justify-start">
      <div className="flex max-w-[85%] items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs">
        <Wrench className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-600" />
        <div className="min-w-0">
          <span className="font-medium text-amber-800">{label}</span>
          {paramSummary && (
            <p className="mt-0.5 truncate text-amber-600">{paramSummary}</p>
          )}
        </div>
      </div>
    </div>
  )
}

function ToolResultCard({ result }: { result?: string }) {
  const summary = getResultSummary(result)

  return (
    <div className="flex justify-start">
      <div className="flex max-w-[85%] items-center gap-2 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-xs">
        <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-green-600" />
        <span className="text-green-700">{summary}</span>
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

/** Maps internal tool names to human-readable action labels. */
const TOOL_LABELS: Record<string, string> = {
  'hr.list_teachers': 'Searching teachers',
  'hr.get_teacher': 'Looking up teacher',
  'subject.list_subjects': 'Searching subjects',
  'subject.get_prerequisites': 'Loading prerequisites',
  'timetable.list_semesters': 'Checking semesters',
  'timetable.generate': 'Generating timetable',
  'timetable.suggest_teachers': 'Finding available teachers',
}

/** Returns a friendly label for a tool name, falling back to title-casing the raw name. */
function friendlyToolLabel(toolName?: string): string {
  if (!toolName) return 'Working…'
  return TOOL_LABELS[toolName] ?? toolName.replace(/[._]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

/**
 * Returns a short, human-readable summary of tool args.
 * Skips UUID values (not meaningful to display).
 */
function friendlyArgs(args: Record<string, unknown>): string | null {
  const meaningful = Object.entries(args).filter(
    ([, v]) => v !== null && v !== undefined && v !== '' && !UUID_RE.test(String(v)),
  )
  if (meaningful.length === 0) return null
  return meaningful.map(([k, v]) => `${k.replace(/_/g, ' ')}: ${String(v)}`).join(' · ')
}

/** Returns a plain-language summary of a tool result (no JSON). */
function getResultSummary(result?: string): string {
  if (!result) return 'Done'
  try {
    const parsed = JSON.parse(result)
    if (Array.isArray(parsed)) {
      return parsed.length === 0 ? 'No results found' : `Found ${parsed.length} result${parsed.length === 1 ? '' : 's'}`
    }
  } catch {
    // not JSON
  }
  return 'Done'
}
