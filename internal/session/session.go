package session

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

type Session struct {
	Messages             []Message
	Effort               string
	Focus                string
	OptimizeCount        int
	CurrentContextTokens int
	SessionTokens        int
	BeforeOptimizeTokens int
	AfterOptimizeTokens  int
	pairsSinceOptimize   int
}

func New() *Session {
	return &Session{Effort: "medium"}
}

func (s *Session) Add(role string, content string) {
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	s.SessionTokens += EstimateTokens(content)
	if role == "assistant" {
		s.pairsSinceOptimize++
	}
}

func (s *Session) PendingPairs() int {
	return s.pairsSinceOptimize
}

func (s *Session) SetEffort(effort string) error {
	effort = strings.ToLower(strings.TrimSpace(effort))
	switch effort {
	case "low", "medium", "high", "max":
		s.Effort = effort
		return nil
	default:
		return fmt.Errorf("invalid effort %q; use low, medium, high, or max", effort)
	}
}

func (s *Session) RecentTranscript() string {
	var b strings.Builder
	for _, msg := range s.Messages {
		b.WriteString(strings.ToUpper(msg.Role))
		b.WriteString(": ")
		b.WriteString(strings.TrimSpace(msg.Content))
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func (s *Session) Compact(summary string) {
	s.BeforeOptimizeTokens = s.SessionTokens
	s.Messages = nil
	s.SessionTokens = 0
	if strings.TrimSpace(summary) != "" {
		s.Add("system", "Current summary:\n"+strings.TrimSpace(summary))
	}
	s.AfterOptimizeTokens = s.SessionTokens
	s.pairsSinceOptimize = 0
	s.OptimizeCount++
}

func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}
