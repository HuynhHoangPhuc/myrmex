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

const claudeAPIVersion = "2023-06-01"
const claudeDefaultBaseURL = "https://api.anthropic.com/v1"

// ClaudeProvider implements LLMProvider for Anthropic Claude API.
type ClaudeProvider struct {
	apiKey  string
	model   string // e.g. "claude-haiku-4-5"
	baseURL string
	client  *http.Client
}

// NewClaudeProvider creates a new Claude provider.
func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: claudeDefaultBaseURL,
		client:  &http.Client{},
	}
}

// --- Claude API request/response types ---

type claudeMessage struct {
	Role    string        `json:"role"`
	Content claudeContent `json:"content"`
}

// claudeContent can be a string or a slice of content blocks.
// We use interface{} to support both forms.
type claudeContent = interface{}

type claudeTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeToolUseBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type claudeToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type claudeTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type claudeChatRequest struct {
	Model     string          `json:"model"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
	Tools     []claudeTool    `json:"tools,omitempty"`
	MaxTokens int             `json:"max_tokens"`
	Stream    bool            `json:"stream"`
}

type claudeAPIResponse struct {
	ID         string            `json:"id"`
	StopReason string            `json:"stop_reason"`
	Content    []json.RawMessage `json:"content"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// buildClaudeMessages converts internal messages to Claude format.
// Claude uses alternating user/assistant turns; tool results go in user messages.
func buildClaudeMessages(msgs []Message) []claudeMessage {
	result := make([]claudeMessage, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "tool_result":
			// Tool results are user-role messages with content blocks
			result = append(result, claudeMessage{
				Role: "user",
				Content: []claudeToolResultBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
		case "assistant":
			result = append(result, claudeMessage{
				Role:    "assistant",
				Content: m.Content,
			})
		default:
			result = append(result, claudeMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}
	return result
}

// buildClaudeTools converts internal Tool slice to Claude format.
func buildClaudeTools(tools []Tool) []claudeTool {
	result := make([]claudeTool, 0, len(tools))
	for _, t := range tools {
		params := t.Parameters
		if params == nil {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		result = append(result, claudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: params,
		})
	}
	return result
}

// doRequest executes a POST request against the Claude API.
func (p *ClaudeProvider) doRequest(ctx context.Context, payload claudeChatRequest) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)

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
		return nil, fmt.Errorf("Claude API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// ChatWithTools sends messages with tool definitions and returns text or tool calls.
func (p *ClaudeProvider) ChatWithTools(ctx context.Context, req ChatRequest, tools []Tool) (*ToolCallResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := claudeChatRequest{
		Model:     p.model,
		System:    req.SystemPrompt,
		Messages:  buildClaudeMessages(req.Messages),
		MaxTokens: maxTokens,
		Stream:    false,
	}
	if len(tools) > 0 {
		payload.Tools = buildClaudeTools(tools)
	}

	respBody, err := p.doRequest(ctx, payload)
	if err != nil {
		return nil, err
	}

	var apiResp claudeAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("Claude error: %s", apiResp.Error.Message)
	}

	result := &ToolCallResponse{}
	for _, rawBlock := range apiResp.Content {
		// Peek at type field
		var typeHolder struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawBlock, &typeHolder); err != nil {
			continue
		}
		switch typeHolder.Type {
		case "text":
			var tb claudeTextBlock
			if err := json.Unmarshal(rawBlock, &tb); err == nil {
				result.Content += tb.Text
			}
		case "tool_use":
			var tub claudeToolUseBlock
			if err := json.Unmarshal(rawBlock, &tub); err == nil {
				result.ToolCalls = append(result.ToolCalls, ToolCall{
					ID:        tub.ID,
					ToolName:  tub.Name,
					Arguments: tub.Input,
				})
			}
		}
	}

	return result, nil
}

// StreamChat streams responses from Claude API as StreamEvents.
// Claude SSE emits content_block_delta events for text and tool input deltas.
func (p *ClaudeProvider) StreamChat(ctx context.Context, req ChatRequest, tools []Tool) (<-chan StreamEvent, error) {
	events := make(chan StreamEvent, 100)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := claudeChatRequest{
		Model:     p.model,
		System:    req.SystemPrompt,
		Messages:  buildClaudeMessages(req.Messages),
		MaxTokens: maxTokens,
		Stream:    true,
	}
	if len(tools) > 0 {
		payload.Tools = buildClaudeTools(tools)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Claude API error: %d", resp.StatusCode)
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

			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				accumulated.Write(buf[:n])
				text := accumulated.String()
				lines := strings.Split(text, "\n")

				for _, line := range lines[:len(lines)-1] {
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					data := strings.TrimPrefix(line, "data: ")

					var event struct {
						Type  string `json:"type"`
						Delta *struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"delta"`
					}
					if jsonErr := json.Unmarshal([]byte(data), &event); jsonErr != nil {
						continue
					}

					switch event.Type {
					case "content_block_delta":
						if event.Delta != nil && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
							events <- StreamEvent{Type: "text", Content: event.Delta.Text}
						}
					case "message_stop":
						events <- StreamEvent{Type: "done"}
						return
					}
				}
				accumulated.Reset()
				accumulated.WriteString(lines[len(lines)-1])
			}

			if readErr == io.EOF {
				events <- StreamEvent{Type: "done"}
				return
			}
			if readErr != nil {
				events <- StreamEvent{Type: "error", Content: readErr.Error()}
				return
			}
		}
	}()

	return events, nil
}
