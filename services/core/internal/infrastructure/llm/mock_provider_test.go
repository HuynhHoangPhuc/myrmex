package llm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMockProvider_ChatWithTools_TextResponse(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.ChatWithTools(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello world"}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("expected text content, got empty")
	}
	if len(resp.ToolCalls) != 0 {
		t.Fatalf("expected no tool calls, got %d", len(resp.ToolCalls))
	}
}

func TestMockProvider_ChatWithTools_ToolCallResponse(t *testing.T) {
	tests := []struct {
		input    string
		toolName string
	}{
		{"please list teachers", "list_teachers"},
		{"can you list subjects", "list_subjects"},
		{"create subject for math", "create_subject"},
		{"create teacher named John", "create_teacher"},
		{"list semesters please", "list_semesters"},
	}

	p := NewMockProvider()
	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			resp, err := p.ChatWithTools(context.Background(), ChatRequest{
				Messages: []Message{{Role: "user", Content: tt.input}},
			}, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.ToolCalls) != 1 {
				t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
			}
			if resp.ToolCalls[0].ToolName != tt.toolName {
				t.Errorf("expected tool %q, got %q", tt.toolName, resp.ToolCalls[0].ToolName)
			}
			// Verify arguments are valid JSON
			if !json.Valid(resp.ToolCalls[0].Arguments) {
				t.Errorf("tool call arguments are not valid JSON: %s", resp.ToolCalls[0].Arguments)
			}
		})
	}
}

func TestMockProvider_StreamChat_TextEvents(t *testing.T) {
	p := NewMockProvider()
	ch, err := p.StreamChat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var textParts int
	var gotDone bool
	for ev := range ch {
		switch ev.Type {
		case "text":
			textParts++
		case "done":
			gotDone = true
		}
	}
	if textParts == 0 {
		t.Error("expected at least one text event")
	}
	if !gotDone {
		t.Error("expected done event")
	}
}

func TestMockProvider_StreamChat_ToolCallEvent(t *testing.T) {
	p := NewMockProvider()
	ch, err := p.StreamChat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "list teachers"}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotToolCall bool
	var gotDone bool
	for ev := range ch {
		switch ev.Type {
		case "tool_call":
			gotToolCall = true
			if ev.ToolCall.ToolName != "list_teachers" {
				t.Errorf("expected list_teachers, got %s", ev.ToolCall.ToolName)
			}
		case "done":
			gotDone = true
		}
	}
	if !gotToolCall {
		t.Error("expected tool_call event")
	}
	if !gotDone {
		t.Error("expected done event")
	}
}
