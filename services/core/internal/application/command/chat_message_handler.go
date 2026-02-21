package command

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/agent"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

const (
	maxToolIterations = 5
	maxHistoryWindow  = 20
	defaultMaxTokens  = 4096
)

const systemPrompt = `You are Myrmex, an AI assistant for a university faculty management system.
You help administrators manage teachers, academic subjects, and timetables.

Available capabilities via tools:
- HR module: list teachers, get teacher details
- Subject module: list subjects, view prerequisite chains
- Timetable module: generate semester schedules, suggest available teachers

Guidelines:
- Always use tools to fetch real data; never make up teacher names, IDs, or schedules
- Be concise and structured in your responses
- When listing results, format them clearly with bullet points or numbered lists
- If a tool call fails, inform the user and suggest alternatives
- Limit tool call chains to what is necessary to answer the question`

// ChatMessageHandler orchestrates multi-turn LLM conversations with tool calling.
type ChatMessageHandler struct {
	llm      llm.LLMProvider
	registry *agent.ToolRegistry
	executor *agent.ToolExecutor
	log      *zap.Logger
}

// NewChatMessageHandler creates a new handler with all dependencies injected.
func NewChatMessageHandler(
	provider llm.LLMProvider,
	registry *agent.ToolRegistry,
	executor *agent.ToolExecutor,
	log *zap.Logger,
) *ChatMessageHandler {
	return &ChatMessageHandler{
		llm:      provider,
		registry: registry,
		executor: executor,
		log:      log,
	}
}

// Handle processes a user message with conversation history and streams events back.
// It implements a multi-turn tool-calling loop (max maxToolIterations iterations).
// The returned channel is closed when the response is complete or an error occurs.
func (h *ChatMessageHandler) Handle(ctx context.Context, userMsg string, history []llm.Message) (<-chan llm.StreamEvent, error) {
	events := make(chan llm.StreamEvent, 100)

	// Trim history to context window
	msgs := trimHistory(history, maxHistoryWindow)
	msgs = append(msgs, llm.Message{Role: "user", Content: userMsg})

	tools := h.registry.GetAllTools()

	go func() {
		defer close(events)

		for iteration := 0; iteration < maxToolIterations; iteration++ {
			h.log.Debug("LLM call", zap.Int("iteration", iteration), zap.Int("messages", len(msgs)))

			resp, err := h.llm.ChatWithTools(ctx, llm.ChatRequest{
				SystemPrompt: systemPrompt,
				Messages:     msgs,
				MaxTokens:    defaultMaxTokens,
			}, tools)
			if err != nil {
				events <- llm.StreamEvent{Type: "error", Content: fmt.Sprintf("LLM error: %s", err.Error())}
				return
			}

			// No tool calls — emit final text and finish
			if len(resp.ToolCalls) == 0 {
				events <- llm.StreamEvent{Type: "text", Content: resp.Content}
				events <- llm.StreamEvent{Type: "done"}
				return
			}

			// Append assistant turn with tool calls to history
			msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Content})

			// Execute each tool call, emit progress events, append results
			for _, tc := range resp.ToolCalls {
				toolCall := tc // capture loop var
				events <- llm.StreamEvent{Type: "tool_call", ToolCall: &toolCall}

				result, execErr := h.executor.Execute(ctx, tc)
				if execErr != nil {
					h.log.Warn("tool execution error",
						zap.String("tool", tc.ToolName),
						zap.Error(execErr),
					)
					result = fmt.Sprintf(`{"error": %q}`, execErr.Error())
				}

				events <- llm.StreamEvent{
					Type:    "tool_result",
					Content: result,
					ToolCall: &llm.ToolCall{
						ID:       tc.ID,
						ToolName: tc.ToolName,
					},
				}

				// Append tool result to message history so LLM can use it
				msgs = append(msgs, llm.Message{
					Role:       "tool_result",
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
			// Loop: send updated messages (including tool results) back to LLM
		}

		// Exceeded max iterations without a final text response
		events <- llm.StreamEvent{
			Type:    "text",
			Content: "I reached the maximum number of tool calls. Please try a more specific question.",
		}
		events <- llm.StreamEvent{Type: "done"}
	}()

	return events, nil
}

// trimHistory returns the last n messages to respect the context window limit.
func trimHistory(msgs []llm.Message, n int) []llm.Message {
	if len(msgs) <= n {
		result := make([]llm.Message, len(msgs))
		copy(result, msgs)
		return result
	}
	result := make([]llm.Message, n)
	copy(result, msgs[len(msgs)-n:])
	return result
}
