package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIProvider implements LLMProvider for OpenAI-compatible APIs.
// Works with OpenAI, Ollama, vLLM, and other OpenAI-compatible endpoints.
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI-compatible provider.
// baseURL: e.g. "https://api.openai.com/v1" or "http://localhost:11434/v1" for Ollama
func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// --- OpenAI API request/response types ---

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content"` // string or []contentPart
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function openAIToolFunction  `json:"function"`
}

type openAIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIChatRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	Tools     []openAITool    `json:"tools,omitempty"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Stream    bool            `json:"stream"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// mapMessages converts internal messages to OpenAI format.
func mapMessages(systemPrompt string, msgs []Message) []openAIMessage {
	result := make([]openAIMessage, 0, len(msgs)+1)
	if systemPrompt != "" {
		result = append(result, openAIMessage{Role: "system", Content: systemPrompt})
	}
	for _, m := range msgs {
		switch m.Role {
		case "tool_result":
			result = append(result, openAIMessage{
				Role:       "tool",
				Content:    m.Content,
				ToolCallID: m.ToolCallID,
			})
		default:
			result = append(result, openAIMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}
	return result
}

// mapTools converts internal Tool slice to OpenAI format.
func mapTools(tools []Tool) []openAITool {
	result := make([]openAITool, 0, len(tools))
	for _, t := range tools {
		params := t.Parameters
		if params == nil {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		result = append(result, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return result
}

// ChatWithTools sends a request with tool definitions and returns the response.
func (p *OpenAIProvider) ChatWithTools(ctx context.Context, req ChatRequest, tools []Tool) (*ToolCallResponse, error) {
	payload := openAIChatRequest{
		Model:     p.model,
		Messages:  mapMessages(req.SystemPrompt, req.Messages),
		MaxTokens: req.MaxTokens,
		Stream:    false,
	}
	if len(tools) > 0 {
		payload.Tools = mapTools(tools)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp openAIChatResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("LLM error: %s", apiResp.Error.Message)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from LLM")
	}

	choice := apiResp.Choices[0]
	result := &ToolCallResponse{}

	// Extract text content
	if s, ok := choice.Message.Content.(string); ok {
		result.Content = s
	}

	// Extract tool calls
	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			ToolName:  tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	return result, nil
}

// StreamChat sends a request and streams responses as StreamEvents.
// Uses SSE (Server-Sent Events) from the OpenAI streaming API.
func (p *OpenAIProvider) StreamChat(ctx context.Context, req ChatRequest, tools []Tool) (<-chan StreamEvent, error) {
	events := make(chan StreamEvent, 100)

	payload := openAIChatRequest{
		Model:     p.model,
		Messages:  mapMessages(req.SystemPrompt, req.Messages),
		MaxTokens: req.MaxTokens,
		Stream:    true,
	}
	if len(tools) > 0 {
		payload.Tools = mapTools(tools)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("LLM API error: %d", resp.StatusCode)
	}

	go func() {
		defer close(events)
		defer resp.Body.Close()

		buf := make([]byte, 4096)
		var accumulated strings.Builder

		for {
			select {
			case <-ctx.Done():
				events <- StreamEvent{Type: "error", Content: "context cancelled"}
				return
			default:
			}

			n, err := resp.Body.Read(buf)
			if n > 0 {
				accumulated.Write(buf[:n])
				text := accumulated.String()
				lines := strings.Split(text, "\n")

				// Process all complete lines; keep last partial line
				for i, line := range lines[:len(lines)-1] {
					_ = i
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					data := strings.TrimPrefix(line, "data: ")
					if data == "[DONE]" {
						events <- StreamEvent{Type: "done"}
						return
					}
					var chunk struct {
						Choices []struct {
							Delta struct {
								Content   string           `json:"content"`
								ToolCalls []openAIToolCall `json:"tool_calls"`
							} `json:"delta"`
						} `json:"choices"`
					}
					if jsonErr := json.Unmarshal([]byte(data), &chunk); jsonErr != nil {
						continue
					}
					if len(chunk.Choices) > 0 {
						delta := chunk.Choices[0].Delta
						if delta.Content != "" {
							events <- StreamEvent{Type: "text", Content: delta.Content}
						}
						for _, tc := range delta.ToolCalls {
							events <- StreamEvent{
								Type: "tool_call",
								ToolCall: &ToolCall{
									ID:        tc.ID,
									ToolName:  tc.Function.Name,
									Arguments: json.RawMessage(tc.Function.Arguments),
								},
							}
						}
					}
				}
				accumulated.Reset()
				accumulated.WriteString(lines[len(lines)-1])
			}

			if err == io.EOF {
				events <- StreamEvent{Type: "done"}
				return
			}
			if err != nil {
				events <- StreamEvent{Type: "error", Content: err.Error()}
				return
			}
		}
	}()

	return events, nil
}
