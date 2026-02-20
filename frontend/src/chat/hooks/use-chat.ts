import { useCallback, useEffect, useRef, useState } from 'react'
import { authStore } from '@/lib/stores/auth-store'
import type { ChatMessage, WsClientMessage, WsServerEvent } from '@/chat/types'

// WebSocket base URL derived from the API base URL (http → ws, https → wss)
function getWsBaseUrl(): string {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  // Strip /api suffix if present
  const base = apiUrl.replace(/\/api$/, '')
  return base.replace(/^http/, 'ws')
}

const WS_URL = `${getWsBaseUrl()}/ws/chat`

// Reconnection config
const RECONNECT_DELAY_MS = 2_000
const MAX_RECONNECT_ATTEMPTS = 5

interface UseChatReturn {
  messages: ChatMessage[]
  isConnected: boolean
  isStreaming: boolean
  sendMessage: (content: string) => void
  clearMessages: () => void
}

/**
 * useChat manages a WebSocket connection to /ws/chat and exposes
 * message state + send function. Handles auto-reconnect with backoff.
 */
export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [isStreaming, setIsStreaming] = useState(false)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttempts = useRef(0)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  // Accumulate streaming text into the last assistant message
  const streamingIdRef = useRef<string | null>(null)

  const appendMessage = useCallback((msg: ChatMessage) => {
    setMessages((prev) => [...prev, msg])
  }, [])

  const connect = useCallback(() => {
    const token = authStore.getAccessToken()
    if (!token) return

    const ws = new WebSocket(`${WS_URL}?token=${encodeURIComponent(token)}`)
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
      reconnectAttempts.current = 0
    }

    ws.onclose = () => {
      setIsConnected(false)
      setIsStreaming(false)
      streamingIdRef.current = null

      // Auto-reconnect with cap
      if (reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        const delay = RECONNECT_DELAY_MS * Math.pow(2, reconnectAttempts.current)
        reconnectAttempts.current += 1
        reconnectTimer.current = setTimeout(connect, delay)
      }
    }

    ws.onerror = () => {
      // onerror is always followed by onclose, so reconnect is handled there
    }

    ws.onmessage = (event: MessageEvent) => {
      let data: WsServerEvent
      try {
        data = JSON.parse(event.data as string) as WsServerEvent
      } catch {
        return
      }

      switch (data.type) {
        case 'text': {
          const text = data.content ?? ''
          // Accumulate into existing streaming message or start a new one
          if (streamingIdRef.current) {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === streamingIdRef.current
                  ? { ...m, content: m.content + text }
                  : m,
              ),
            )
          } else {
            const id = crypto.randomUUID()
            streamingIdRef.current = id
            setIsStreaming(true)
            appendMessage({
              id,
              role: 'assistant',
              content: text,
              timestamp: new Date(),
            })
          }
          break
        }

        case 'tool_call': {
          appendMessage({
            id: crypto.randomUUID(),
            role: 'tool_call',
            content: '',
            toolName: data.tool,
            toolArgs: data.args,
            timestamp: new Date(),
          })
          break
        }

        case 'tool_result': {
          appendMessage({
            id: crypto.randomUUID(),
            role: 'tool_result',
            content: '',
            toolName: data.tool,
            toolResult: data.result,
            timestamp: new Date(),
          })
          break
        }

        case 'done': {
          streamingIdRef.current = null
          setIsStreaming(false)
          break
        }

        case 'error': {
          streamingIdRef.current = null
          setIsStreaming(false)
          appendMessage({
            id: crypto.randomUUID(),
            role: 'assistant',
            content: `Error: ${data.content ?? 'unknown error'}`,
            timestamp: new Date(),
          })
          break
        }
      }
    }
  }, [appendMessage])

  // Connect on mount; clean up on unmount
  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const sendMessage = useCallback((content: string) => {
    if (!content.trim()) return

    // Optimistically add user message to UI
    appendMessage({
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: new Date(),
    })

    // Reset streaming state for this new turn
    streamingIdRef.current = null

    const msg: WsClientMessage = { type: 'message', content }
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg))
    }
  }, [appendMessage])

  const clearMessages = useCallback(() => {
    setMessages([])
    streamingIdRef.current = null
  }, [])

  return { messages, isConnected, isStreaming, sendMessage, clearMessages }
}
