package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Theme         string
	Thinking      string
	ShowBudget    bool
	TypingEffect  bool
	TypingDelayMs int
}

type settingsFile struct {
	UI settingsUI `json:"ui"`
}

type settingsUI struct {
	Theme         string `json:"theme"`
	Thinking      string `json:"thinking"`
	ShowBudget    *bool  `json:"showBudget"`
	TypingEffect  *bool  `json:"typingEffect"`
	TypingDelayMs int    `json:"typingDelayMs"`
}

func LoadConfig(root string) Config {
	cfg := Config{Theme: "default", Thinking: "minimal", TypingEffect: true, TypingDelayMs: 10}
	content, err := os.ReadFile(filepath.Join(root, ".shard", "settings.json"))
	if err != nil {
		return cfg
	}
	var parsed settingsFile
	if err := json.Unmarshal(content, &parsed); err != nil {
		return cfg
	}
	if parsed.UI.Theme != "" {
		cfg.Theme = parsed.UI.Theme
	}
	if parsed.UI.Thinking != "" {
		cfg.Thinking = parsed.UI.Thinking
	}
	if parsed.UI.ShowBudget != nil {
		cfg.ShowBudget = *parsed.UI.ShowBudget
	}
	if parsed.UI.TypingEffect != nil {
		cfg.TypingEffect = *parsed.UI.TypingEffect
	}
	if parsed.UI.TypingDelayMs > 0 {
		cfg.TypingDelayMs = parsed.UI.TypingDelayMs
	}
	return cfg
}
