package provider

import (
	"context"
	"fmt"
	"strings"
)

type Message struct {
	Role    string
	Content string
}

type Provider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	Name() string
	Provider() string
}

func New(providerName string, model string) (Provider, error) {
	providerName = strings.ToLower(strings.TrimSpace(firstNonEmpty(providerName, envValue("SHARD_PROVIDER"))))
	switch providerName {
	case "", "anthropic":
		return NewAnthropicProvider(firstNonEmpty(model, envValue("ANTHROPIC_MODEL"), "claude-3-5-sonnet-20241022")), nil
	case "openrouter":
		return NewOpenRouterProvider(firstNonEmpty(model, envValue("OPENROUTER_MODEL"), "anthropic/claude-3.5-sonnet")), nil
	default:
		return nil, fmt.Errorf("unknown provider %q; use anthropic or openrouter", providerName)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
