package ui

import (
	"fmt"

	"shard/internal/session"
)

func PrintHeader(model string, effort string, contextTokens int) {
	fmt.Printf("\nModel: %s | Effort: %s | Context: %d tokens\n\n", model, effort, contextTokens)
}

func PrintStats(sess *session.Session) {
	focus := sess.Focus
	if focus == "" {
		focus = "none"
	}
	fmt.Println("Shard stats")
	fmt.Printf("Session messages: %d\n", len(sess.Messages))
	fmt.Printf("Optimize count: %d\n", sess.OptimizeCount)
	fmt.Printf("Estimated tokens: %d\n", sess.SessionTokens)
	fmt.Printf("Current context estimate: %d\n", sess.CurrentContextTokens)
	fmt.Printf("Effort: %s\n", sess.Effort)
	fmt.Printf("Focus: %s\n", focus)
}
