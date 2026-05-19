package optimizer

import (
	"context"
	"fmt"
	"strings"

	"shard/internal/memory"
	"shard/internal/provider"
	"shard/internal/session"
)

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

func Run(ctx context.Context, sess *session.Session, store *memory.Store, prov provider.Provider) error {
	transcript := sess.RecentTranscript()
	if strings.TrimSpace(transcript) == "" {
		return fmt.Errorf("nothing to optimize yet")
	}

	mode := fmt.Sprintf("Optimizer mode: %s effort. Be concise and preserve durable facts.", sess.Effort)
	result, err := prov.Chat(ctx, []provider.Message{
		{Role: "system", Content: optimizerPrompt + "\n\n" + mode},
		{Role: "user", Content: transcript},
	})
	if err != nil {
		return err
	}

	sections := ParseSections(result)
	for _, category := range memory.Categories {
		if err := store.AppendSection(category, sections[category]); err != nil {
			return err
		}
	}
	sess.Compact(sections["current"])
	fmt.Printf("Optimized memory. Before: %d tokens | After: %d tokens\n", sess.BeforeOptimizeTokens, sess.AfterOptimizeTokens)
	return nil
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
