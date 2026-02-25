package command

import (
	"context"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/agent"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
	"go.uber.org/zap"
)

// mockLLMProvider implements llm.LLMProvider for testing.
type mockLLMProvider struct {
	resp *llm.ToolCallResponse
	err  error
}

func (m *mockLLMProvider) ChatWithTools(_ context.Context, _ llm.ChatRequest, _ []llm.Tool) (*llm.ToolCallResponse, error) {
	return m.resp, m.err
}

func (m *mockLLMProvider) StreamChat(_ context.Context, _ llm.ChatRequest, _ []llm.Tool) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	return ch, nil
}

func TestTrimHistory_ShortHistory(t *testing.T) {
	msgs := []llm.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	}
	result := trimHistory(msgs, 10)
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
}

func TestTrimHistory_LongHistory(t *testing.T) {
	msgs := make([]llm.Message, 25)
	for i := range msgs {
		msgs[i] = llm.Message{Role: "user", Content: "msg"}
	}
	result := trimHistory(msgs, 20)
	if len(result) != 20 {
		t.Fatalf("expected 20 messages, got %d", len(result))
	}
}

func TestTrimHistory_ExactWindow(t *testing.T) {
	msgs := make([]llm.Message, 20)
	for i := range msgs {
		msgs[i] = llm.Message{Role: "user", Content: "msg"}
	}
	result := trimHistory(msgs, 20)
	if len(result) != 20 {
		t.Fatalf("expected 20 messages, got %d", len(result))
	}
}

func TestTrimHistory_Empty(t *testing.T) {
	result := trimHistory([]llm.Message{}, 10)
	if len(result) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(result))
	}
}

func TestChatMessageHandler_Handle_TextResponse(t *testing.T) {
	provider := &mockLLMProvider{
		resp: &llm.ToolCallResponse{Content: "Hello!", ToolCalls: nil},
	}
	registry := agent.NewToolRegistry()
	logger := zap.NewNop()
	h := NewChatMessageHandler(provider, registry, nil, logger)

	events, err := h.Handle(context.Background(), "Hi", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []llm.StreamEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(got))
	}
	if got[0].Type != "text" {
		t.Fatalf("expected first event type=text, got %s", got[0].Type)
	}
	if got[0].Content != "Hello!" {
		t.Fatalf("unexpected content: %s", got[0].Content)
	}
	if got[len(got)-1].Type != "done" {
		t.Fatalf("expected last event type=done, got %s", got[len(got)-1].Type)
	}
}

func TestChatMessageHandler_Handle_LLMError(t *testing.T) {
	provider := &mockLLMProvider{resp: nil, err: context.DeadlineExceeded}
	registry := agent.NewToolRegistry()
	logger := zap.NewNop()
	h := NewChatMessageHandler(provider, registry, nil, logger)

	events, err := h.Handle(context.Background(), "Hi", nil)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	var got []llm.StreamEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one error event")
	}
	if got[0].Type != "error" {
		t.Fatalf("expected error event, got %s", got[0].Type)
	}
}

func TestChatMessageHandler_Handle_WithHistory(t *testing.T) {
	provider := &mockLLMProvider{
		resp: &llm.ToolCallResponse{Content: "Got it", ToolCalls: nil},
	}
	registry := agent.NewToolRegistry()
	logger := zap.NewNop()
	h := NewChatMessageHandler(provider, registry, nil, logger)

	history := []llm.Message{
		{Role: "user", Content: "prev question"},
		{Role: "assistant", Content: "prev answer"},
	}

	events, err := h.Handle(context.Background(), "new question", history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []llm.StreamEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(got))
	}
	last := got[len(got)-1]
	if last.Type != "done" {
		t.Fatalf("expected done event, got %s", last.Type)
	}
}
