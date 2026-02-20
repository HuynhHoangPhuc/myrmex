package llm

import "context"

// LLMProvider is the provider-agnostic interface for LLM interactions.
// Implementations: ClaudeProvider, OpenAIProvider.
type LLMProvider interface {
	// ChatWithTools sends messages with tool definitions and returns text or tool calls.
	ChatWithTools(ctx context.Context, req ChatRequest, tools []Tool) (*ToolCallResponse, error)
	// StreamChat sends messages and emits streaming events (text tokens + tool calls).
	StreamChat(ctx context.Context, req ChatRequest, tools []Tool) (<-chan StreamEvent, error)
}
