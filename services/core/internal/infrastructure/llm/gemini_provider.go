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

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

// GeminiProvider implements LLMProvider for Google Gemini API.
type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

// --- Gemini API request/response types ---

type geminiPart struct {
	Text             string                `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResult `json:"functionResponse,omitempty"`
	// ThoughtSignature is set by thinking-enabled models on functionCall parts.
	// It must be preserved and replayed exactly when reconstructing history.
	ThoughtSignature string `json:"thoughtSignature,omitempty"`
}

type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type geminiFunctionResult struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"` // "user" or "model"
	Parts []geminiPart `json:"parts"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiFunctionDecl struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDecl `json:"functionDeclarations"`
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	Tools             []geminiTool             `json:"tools,omitempty"`
	GenerationConfig  *geminiGenerationConfig  `json:"generationConfig,omitempty"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int                  `json:"maxOutputTokens,omitempty"`
	ThinkingConfig  *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

// geminiThinkingConfig controls the model's internal reasoning budget.
// Set ThinkingBudget=0 to disable thinking, which avoids the thought_signature
// requirement when replaying function calls in conversation history.
type geminiThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// buildGeminiContents converts internal messages to Gemini format.
// Gemini uses "user"/"model" roles. Tool calls must appear as functionCall parts in the
// model turn; tool results must appear as functionResponse parts in a user turn.
func buildGeminiContents(msgs []Message) []geminiContent {
	contents := make([]geminiContent, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "tool_result":
			// Tool results are user-role functionResponse parts.
			part := geminiPart{
				FunctionResponse: &geminiFunctionResult{
					Name:     m.ToolCallID, // ToolCallID holds the tool name for Gemini
					Response: map[string]any{"result": m.Content},
				},
			}
			// Merge consecutive tool results into the same user turn.
			if len(contents) > 0 && contents[len(contents)-1].Role == "user" {
				last := &contents[len(contents)-1]
				if len(last.Parts) > 0 && last.Parts[0].FunctionResponse != nil {
					last.Parts = append(last.Parts, part)
					continue
				}
			}
			contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{part}})

		case "assistant":
			// Build model turn: text (if any) + functionCall parts (if any).
			var parts []geminiPart
			if m.Content != "" {
				parts = append(parts, geminiPart{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var args map[string]any
				_ = json.Unmarshal(tc.Arguments, &args)
				part := geminiPart{
					FunctionCall: &geminiFunctionCall{Name: tc.ToolName, Args: args},
				}
				// Replay thoughtSignature required by thinking-enabled models.
				if tc.ProviderMeta != nil {
					part.ThoughtSignature = tc.ProviderMeta["thoughtSignature"]
				}
				parts = append(parts, part)
			}
			if len(parts) == 0 {
				// Skip fully empty assistant turns (shouldn't happen, but guard anyway).
				continue
			}
			contents = append(contents, geminiContent{Role: "model", Parts: parts})

		default: // "user"
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}
	return contents
}

// buildGeminiTools converts internal Tool slice to Gemini function declarations.
func buildGeminiTools(tools []Tool) []geminiTool {
	if len(tools) == 0 {
		return nil
	}
	decls := make([]geminiFunctionDecl, 0, len(tools))
	for _, t := range tools {
		params := t.Parameters
		if params == nil {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		decls = append(decls, geminiFunctionDecl{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  params,
		})
	}
	return []geminiTool{{FunctionDeclarations: decls}}
}

// parseGeminiCandidate extracts text and function calls from a response candidate.
func parseGeminiCandidate(content geminiContent) *ToolCallResponse {
	result := &ToolCallResponse{}
	for _, part := range content.Parts {
		if part.Text != "" {
			result.Content += part.Text
		}
		if part.FunctionCall != nil {
			argsJSON, _ := json.Marshal(part.FunctionCall.Args)
			tc := ToolCall{
				ID:        part.FunctionCall.Name, // Gemini has no separate call ID
				ToolName:  part.FunctionCall.Name,
				Arguments: argsJSON,
			}
			// Preserve thoughtSignature so it can be replayed in history.
			// Thinking-enabled models require this field on replayed functionCall parts.
			if part.ThoughtSignature != "" {
				tc.ProviderMeta = map[string]string{"thoughtSignature": part.ThoughtSignature}
			}
			result.ToolCalls = append(result.ToolCalls, tc)
		}
	}
	return result
}

// doRequest executes a POST to the Gemini generateContent endpoint.
func (p *GeminiProvider) doRequest(ctx context.Context, endpoint string, payload geminiRequest) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:%s?key=%s", geminiBaseURL, p.model, endpoint, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// buildRequest constructs the common geminiRequest payload.
// Thinking is disabled (ThinkingBudget=0) to avoid the thought_signature requirement
// when replaying functionCall parts in multi-turn conversation history.
func (p *GeminiProvider) buildRequest(req ChatRequest, tools []Tool) geminiRequest {
	payload := geminiRequest{
		Contents: buildGeminiContents(req.Messages),
		Tools:    buildGeminiTools(tools),
		GenerationConfig: &geminiGenerationConfig{
			ThinkingConfig: &geminiThinkingConfig{ThinkingBudget: 0},
		},
	}
	if req.SystemPrompt != "" {
		payload.SystemInstruction = &geminiSystemInstruction{
			Parts: []geminiPart{{Text: req.SystemPrompt}},
		}
	}
	if req.MaxTokens > 0 {
		payload.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}
	return payload
}

// ChatWithTools sends a non-streaming request and returns text or tool calls.
func (p *GeminiProvider) ChatWithTools(ctx context.Context, req ChatRequest, tools []Tool) (*ToolCallResponse, error) {
	respBody, err := p.doRequest(ctx, "generateContent", p.buildRequest(req, tools))
	if err != nil {
		return nil, err
	}

	var apiResp geminiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("Gemini error: %s", apiResp.Error.Message)
	}
	if len(apiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	return parseGeminiCandidate(apiResp.Candidates[0].Content), nil
}

// StreamChat streams responses from Gemini as StreamEvents via SSE.
func (p *GeminiProvider) StreamChat(ctx context.Context, req ChatRequest, tools []Tool) (<-chan StreamEvent, error) {
	events := make(chan StreamEvent, 100)

	payload := p.buildRequest(req, tools)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse&key=%s", geminiBaseURL, p.model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(b))
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

					var chunk geminiResponse
					if jsonErr := json.Unmarshal([]byte(data), &chunk); jsonErr != nil {
						continue
					}
					if chunk.Error != nil {
						events <- StreamEvent{Type: "error", Content: chunk.Error.Message}
						return
					}
					if len(chunk.Candidates) == 0 {
						continue
					}
					for _, part := range chunk.Candidates[0].Content.Parts {
						if part.Text != "" {
							events <- StreamEvent{Type: "text", Content: part.Text}
						}
						if part.FunctionCall != nil {
							argsJSON, _ := json.Marshal(part.FunctionCall.Args)
							tc := &ToolCall{
								ID:        part.FunctionCall.Name,
								ToolName:  part.FunctionCall.Name,
								Arguments: argsJSON,
							}
							if part.ThoughtSignature != "" {
								tc.ProviderMeta = map[string]string{"thoughtSignature": part.ThoughtSignature}
							}
							events <- StreamEvent{Type: "tool_call", ToolCall: tc}
						}
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
