package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AnthropicProvider struct {
	model  string
	client *http.Client
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewAnthropicProvider(model string) *AnthropicProvider {
	return &AnthropicProvider{
		model: model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *AnthropicProvider) Name() string {
	return p.model
}

func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	key := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if key == "" {
		return "", errors.New("ANTHROPIC_API_KEY is not set; set it in your environment before using shard chat or shard optimize")
	}

	reqBody := anthropicRequest{
		Model:     p.model,
		MaxTokens: 2048,
		Messages:  make([]anthropicMessage, 0, len(messages)),
	}
	for _, msg := range messages {
		if msg.Role == "system" {
			if reqBody.System == "" {
				reqBody.System = msg.Content
			} else {
				reqBody.System += "\n\n" + msg.Content
			}
			continue
		}
		role := msg.Role
		if role != "assistant" {
			role = "user"
		}
		reqBody.Messages = append(reqBody.Messages, anthropicMessage{Role: role, Content: msg.Content})
	}
	if len(reqBody.Messages) == 0 {
		reqBody.Messages = append(reqBody.Messages, anthropicMessage{Role: "user", Content: "Summarize the current project state."})
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var decoded anthropicResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode Anthropic response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if decoded.Error != nil {
			return "", fmt.Errorf("anthropic error: %s", decoded.Error.Message)
		}
		return "", fmt.Errorf("anthropic error: HTTP %d", resp.StatusCode)
	}
	for _, block := range decoded.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return block.Text, nil
		}
	}
	return "", errors.New("anthropic returned no text content")
}
