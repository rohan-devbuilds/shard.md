package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenRouterProvider struct {
	model  string
	client *http.Client
}

type openRouterRequest struct {
	Model     string              `json:"model"`
	Messages  []openRouterMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Text    string `json:"text,omitempty"`
		Message struct {
			Content          json.RawMessage `json:"content"`
			Text             string          `json:"text,omitempty"`
			Refusal          string          `json:"refusal,omitempty"`
			Reasoning        string          `json:"reasoning,omitempty"`
			ReasoningContent string          `json:"reasoning_content,omitempty"`
		} `json:"message"`
		Delta struct {
			Content json.RawMessage `json:"content,omitempty"`
			Text    string          `json:"text,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

func NewOpenRouterProvider(model string) *OpenRouterProvider {
	return &OpenRouterProvider{
		model: model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *OpenRouterProvider) Name() string {
	return p.model
}

func (p *OpenRouterProvider) Provider() string {
	return "openrouter"
}

func (p *OpenRouterProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	key := envValue("OPENROUTER_API_KEY")
	if key == "" {
		return "", errors.New("OPENROUTER_API_KEY is not set; export it in your environment or add OPENROUTER_API_KEY=your-key to a local .env file")
	}

	reqBody := openRouterRequest{
		Model:     p.model,
		MaxTokens: 2048,
		Messages:  make([]openRouterMessage, 0, len(messages)),
	}
	for _, msg := range messages {
		role := msg.Role
		if role != "system" && role != "assistant" {
			role = "user"
		}
		reqBody.Messages = append(reqBody.Messages, openRouterMessage{Role: role, Content: msg.Content})
	}
	if len(reqBody.Messages) == 0 {
		reqBody.Messages = append(reqBody.Messages, openRouterMessage{Role: "user", Content: "Summarize the current project state."})
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+key)
	req.Header.Set("http-referer", "https://github.com/shard-cli/shard")
	req.Header.Set("x-title", "Shard")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var decoded openRouterResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode OpenRouter response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return "", fmt.Errorf("openrouter error: %s", decoded.Error.Message)
		}
		return "", fmt.Errorf("openrouter error: HTTP %d", resp.StatusCode)
	}
	for _, choice := range decoded.Choices {
		if strings.TrimSpace(choice.Text) != "" {
			return strings.TrimSpace(choice.Text), nil
		}
		if content := openRouterContentText(choice.Message.Content); content != "" {
			return content, nil
		}
		if strings.TrimSpace(choice.Message.Text) != "" {
			return strings.TrimSpace(choice.Message.Text), nil
		}
		if strings.TrimSpace(choice.Message.Refusal) != "" {
			return strings.TrimSpace(choice.Message.Refusal), nil
		}
		if content := openRouterContentText(choice.Delta.Content); content != "" {
			return content, nil
		}
		if strings.TrimSpace(choice.Delta.Text) != "" {
			return strings.TrimSpace(choice.Delta.Text), nil
		}
		if strings.TrimSpace(choice.Message.ReasoningContent) != "" {
			return strings.TrimSpace(choice.Message.ReasoningContent), nil
		}
		if strings.TrimSpace(choice.Message.Reasoning) != "" {
			return strings.TrimSpace(choice.Message.Reasoning), nil
		}
	}
	return "", fmt.Errorf("openrouter returned no text content; raw response: %s", truncateForError(string(respBody), 1200))
}

func openRouterContentText(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	var parts []struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		Content    string `json:"content"`
		OutputText string `json:"output_text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, part := range parts {
			if strings.TrimSpace(part.Text) != "" {
				b.WriteString(part.Text)
			}
			if strings.TrimSpace(part.Content) != "" {
				b.WriteString(part.Content)
			}
			if strings.TrimSpace(part.OutputText) != "" {
				b.WriteString(part.OutputText)
			}
		}
		return strings.TrimSpace(b.String())
	}
	var generic []map[string]any
	if err := json.Unmarshal(raw, &generic); err == nil {
		var b strings.Builder
		for _, part := range generic {
			b.WriteString(extractText(part))
		}
		return strings.TrimSpace(b.String())
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err == nil {
		return strings.TrimSpace(extractText(object))
	}
	return ""
}

func extractText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		var b strings.Builder
		for _, item := range v {
			b.WriteString(extractText(item))
		}
		return b.String()
	case map[string]any:
		keys := []string{"text", "content", "output_text", "response", "message"}
		var b strings.Builder
		for _, key := range keys {
			if child, ok := v[key]; ok {
				b.WriteString(extractText(child))
			}
		}
		return b.String()
	default:
		return ""
	}
}

func truncateForError(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "... [truncated]"
}
