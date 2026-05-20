package optimizer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"shard/internal/memory"
	"shard/internal/provider"
	"shard/internal/session"
)

type Summary struct {
	MessagesCompressed int
	BeforeTokens       int
	AfterTokens        int
	UpdatedFiles       []string
}

const optimizerPrompt = `You are Shard's memory optimizer.

Convert the recent conversation into structured, durable project memory.

Extract only durable, reusable knowledge:
- goals
- requirements
- tasks
- bugs
- architecture decisions
- attempted fixes
- commands that worked or failed
- file paths
- APIs/providers/models
- UI decisions
- unresolved issues
- next steps

Ignore:
- small talk
- repeated wording
- temporary reasoning
- outdated ideas
- duplicate content

Output Markdown sections using this exact format:

## current
...

## tasks
...

## bugs
...

## architecture
...

## decisions
...

## codebase
...

## commands
...

## dependencies
...

## api
...

## ui
...

## agents
...

## changelog
...

Keep each section concise.`

func Run(ctx context.Context, sess *session.Session, store *memory.Store, prov provider.Provider) (Summary, error) {
	if err := store.Validate(); err != nil {
		return Summary{}, err
	}
	transcript := sess.RecentTranscript()
	if strings.TrimSpace(transcript) == "" {
		return Summary{}, ErrNothingToOptimize
	}
	messageCount := len(sess.Messages)

	mode := fmt.Sprintf("Optimizer mode: %s effort. Be concise and preserve durable facts.", sess.Effort)
	messages := []provider.Message{
		{Role: "system", Content: optimizerPrompt + "\n\n" + mode},
		{Role: "user", Content: transcript},
	}
	inputTokens := estimateMessages(messages)
	result, err := prov.Chat(ctx, messages)
	if err != nil {
		return Summary{}, err
	}
	sess.RecordModelCall(inputTokens, session.EstimateTokens(result))

	sections := ParseSections(result)
	debugOptimize(store, sections)
	updated, err := UpdateMemory(store, sections)
	if err != nil {
		return Summary{}, err
	}
	sess.Compact(sections["current"])
	return Summary{MessagesCompressed: messageCount, BeforeTokens: sess.BeforeOptimizeTokens, AfterTokens: sess.AfterOptimizeTokens, UpdatedFiles: updated}, nil
}

func UpdateMemory(store *memory.Store, sections map[string]string) ([]string, error) {
	return store.AppendSections(sections)
}

func debugOptimize(store *memory.Store, sections map[string]string) {
	if strings.ToLower(os.Getenv("SHARD_DEBUG")) != "1" && strings.ToLower(os.Getenv("SHARD_DEBUG")) != "true" {
		return
	}
	fmt.Fprintf(os.Stderr, "debug optimize: memory dir: %s\n", store.Dir())
	for _, category := range memory.Categories {
		if strings.TrimSpace(sections[category]) != "" {
			fmt.Fprintf(os.Stderr, "debug optimize: parsed section: %s (%d chars)\n", category, len(sections[category]))
			fmt.Fprintf(os.Stderr, "debug optimize: selected file: %s.md\n", category)
		}
	}
}

func estimateMessages(messages []provider.Message) int {
	total := 0
	for _, msg := range messages {
		total += session.EstimateTokens(msg.Role)
		total += session.EstimateTokens(msg.Content)
		total += 4
	}
	return total
}

func ParseSections(text string) map[string]string {
	sections := map[string]string{}
	current := ""
	var b strings.Builder

	flush := func() {
		if current != "" {
			sections[current] = strings.TrimSpace(b.String())
			b.Reset()
		}
	}

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			flush()
			if memory.IsCategory(name) {
				current = name
			} else {
				current = ""
			}
			continue
		}
		if current != "" {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	flush()
	return sections
}
