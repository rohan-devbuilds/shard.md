package retrieval

import "strings"

func SelectFiles(prompt string, focus string, effort string) []string {
	prompt = strings.ToLower(prompt)
	focus = strings.ToLower(strings.TrimSpace(focus))
	files := []string{"current"}

	if containsAny(prompt, "bug", "error", "crash", "fail", "fix") {
		files = append(files, "bugs", "commands", "codebase")
	}
	if containsAny(prompt, "architecture", "design", "system") {
		files = append(files, "architecture", "decisions")
	}
	if containsAny(prompt, "ui", "interface", "terminal", "theme") {
		files = append(files, "ui")
	}
	if containsAny(prompt, "api", "provider", "model", "anthropic", "openai") {
		files = append(files, "api", "dependencies")
	}
	selected := limit(unique(files), limitForEffort(effort))
	if focus != "" && !has(selected, focus) {
		selected = append(selected, focus)
	}
	return selected
}

func containsAny(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func unique(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func limit(items []string, max int) []string {
	if max <= 0 || len(items) <= max {
		return items
	}
	return items[:max]
}

func limitForEffort(effort string) int {
	switch effort {
	case "low":
		return 3
	case "high":
		return 8
	case "max":
		return 12
	default:
		return 5
	}
}

func has(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
