import { useRef, useState, type KeyboardEvent } from 'react'
import { Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils/cn'

interface ChatInputProps {
  onSend: (content: string) => void
  disabled?: boolean
  placeholder?: string
}

/**
 * Chat input: auto-growing textarea with send button.
 * Enter sends, Shift+Enter inserts newline.
 */
export function ChatInput({
  onSend,
  disabled = false,
  placeholder = 'Ask about teachers, subjects, or schedules…',
}: ChatInputProps) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSend = () => {
    const trimmed = value.trim()
    if (!trimmed || disabled) return
    onSend(trimmed)
    setValue('')
    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value)
    // Auto-grow height up to 120px
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 120)}px`
  }

  const canSend = value.trim().length > 0 && !disabled

  return (
    <div className="flex items-end gap-2 border-t bg-background p-3">
      <textarea
        ref={textareaRef}
        rows={1}
        value={value}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        placeholder={placeholder}
        className={cn(
          'flex-1 resize-none rounded-md border bg-background px-3 py-2 text-sm',
          'placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring',
          'disabled:cursor-not-allowed disabled:opacity-50',
          'max-h-[120px] min-h-[38px]',
        )}
      />
      <Button
        size="sm"
        onClick={handleSend}
        disabled={!canSend}
        className="h-9 w-9 shrink-0 p-0"
        aria-label="Send message"
      >
        <Send className="h-4 w-4" />
      </Button>
    </div>
  )
}
