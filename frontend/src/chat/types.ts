// Chat domain types for the AI assistant panel

export type MessageRole = 'user' | 'assistant' | 'tool_call' | 'tool_result'

export interface ChatMessage {
  id: string
  role: MessageRole
  content: string
  /** Tool name — present for tool_call and tool_result messages */
  toolName?: string
  /** Tool arguments — present for tool_call messages */
  toolArgs?: Record<string, unknown>
  /** Tool result JSON string — present for tool_result messages */
  toolResult?: string
  timestamp: Date
}

/** Raw event shape received from the WebSocket server */
export interface WsServerEvent {
  type: 'text' | 'tool_call' | 'tool_result' | 'done' | 'error'
  content?: string
  tool?: string
  args?: Record<string, unknown>
  result?: string
}

/** Message sent from the browser to the WebSocket server */
export interface WsClientMessage {
  type: 'message'
  content: string
}
