package llm

import "encoding/json"

// ChatRequest represents a request to the LLM provider.
type ChatRequest struct {
	SystemPrompt string
	Messages     []Message
	MaxTokens    int
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string // "user", "assistant", "tool_result"
	Content string
	// ToolCallID is set for tool_result messages linking back to the tool call.
	ToolCallID string
	// ToolCalls is set on assistant messages that triggered tool calls.
	// Used by providers (e.g. Gemini) that need explicit functionCall parts in history.
	ToolCalls []ToolCall
}

// Tool represents a tool definition available to the LLM.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage // JSON Schema object
}

// ToolCall represents a tool call requested by the LLM.
type ToolCall struct {
	ID        string
	ToolName  string
	Arguments json.RawMessage
	// ProviderMeta carries provider-specific metadata that must be preserved
	// for correct history replay (e.g. Gemini's thoughtSignature).
	ProviderMeta map[string]string
}

// ToolCallResponse is returned when the LLM responds with tool calls or text.
type ToolCallResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// StreamEvent is emitted during streaming responses.
type StreamEvent struct {
	Type     string    // "text", "tool_call", "done", "error"
	Content  string    // For "text" and "error" types
	ToolCall *ToolCall // For "tool_call" type
}
