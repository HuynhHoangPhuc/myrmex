package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MockProvider implements LLMProvider with deterministic canned responses.
// Activated via LLM_PROVIDER=mock. Intended for E2E tests only — NOT for production.
type MockProvider struct{}

func NewMockProvider() *MockProvider { return &MockProvider{} }

// toolCallRoutes maps input keywords to canned tool call responses.
var toolCallRoutes = map[string]ToolCall{
	"list teachers": {
		ID:        "mock-tc-1",
		ToolName:  "list_teachers",
		Arguments: json.RawMessage(`{}`),
	},
	"list subjects": {
		ID:        "mock-tc-2",
		ToolName:  "list_subjects",
		Arguments: json.RawMessage(`{}`),
	},
	"create subject": {
		ID:        "mock-tc-3",
		ToolName:  "create_subject",
		Arguments: json.RawMessage(`{"name":"Mock Subject","code":"MOCK101","credits":3}`),
	},
	"create teacher": {
		ID:        "mock-tc-4",
		ToolName:  "create_teacher",
		Arguments: json.RawMessage(`{"full_name":"Mock Teacher","employee_code":"MT001","email":"mock@test.com"}`),
	},
	"list semesters": {
		ID:        "mock-tc-5",
		ToolName:  "list_semesters",
		Arguments: json.RawMessage(`{}`),
	},
}

// ChatWithTools returns canned text or tool-call responses based on input keywords.
func (m *MockProvider) ChatWithTools(_ context.Context, req ChatRequest, _ []Tool) (*ToolCallResponse, error) {
	input := lastUserMessage(req.Messages)
	lower := strings.ToLower(input)

	// Check for tool-call keyword matches
	for keyword, tc := range toolCallRoutes {
		if strings.Contains(lower, keyword) {
			return &ToolCallResponse{
				ToolCalls: []ToolCall{tc},
			}, nil
		}
	}

	// Default: return text response
	return &ToolCallResponse{
		Content: fmt.Sprintf("This is a mock response for: %s", input),
	}, nil
}

// StreamChat emits chunked text events simulating real provider streaming behavior.
func (m *MockProvider) StreamChat(_ context.Context, req ChatRequest, _ []Tool) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 10)
	input := lastUserMessage(req.Messages)
	lower := strings.ToLower(input)

	go func() {
		defer close(ch)

		// Check for tool-call keyword matches
		for keyword, tc := range toolCallRoutes {
			if strings.Contains(lower, keyword) {
				tcCopy := tc
				ch <- StreamEvent{Type: "tool_call", ToolCall: &tcCopy}
				ch <- StreamEvent{Type: "done"}
				return
			}
		}

		// Default: emit chunked text response
		response := fmt.Sprintf("This is a mock response for: %s", input)
		chunks := splitIntoChunks(response, 3)
		for _, chunk := range chunks {
			ch <- StreamEvent{Type: "text", Content: chunk}
			time.Sleep(50 * time.Millisecond) // simulate streaming delay
		}
		ch <- StreamEvent{Type: "done"}
	}()

	return ch, nil
}

// lastUserMessage extracts the last user message content from the message list.
func lastUserMessage(msgs []Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return msgs[i].Content
		}
	}
	return ""
}

// splitIntoChunks divides a string into n roughly equal parts.
func splitIntoChunks(s string, n int) []string {
	if n <= 0 || len(s) == 0 {
		return []string{s}
	}
	chunkSize := (len(s) + n - 1) / n
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
