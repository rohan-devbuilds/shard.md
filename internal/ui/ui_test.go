package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderStreamingAssistantPreservesText(t *testing.T) {
	var out bytes.Buffer
	renderer := NewRenderer(&out, Config{Theme: "minimal", Thinking: "minimal"})

	renderer.RenderStreamingAssistant("hello\n```go\nfmt.Println(\"x\")\n```", 0)

	text := out.String()
	for _, want := range []string{"Shard", "hello", "```go", "fmt.Println(\"x\")"} {
		if !strings.Contains(text, want) {
			t.Fatalf("streamed output missing %q: %s", want, text)
		}
	}
}

func TestRenderAssistantCanDisableTypingEffect(t *testing.T) {
	var out bytes.Buffer
	renderer := NewRenderer(&out, Config{Theme: "minimal", Thinking: "minimal", TypingEffect: false})

	renderer.RenderAssistant("hello")

	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("output missing response: %s", out.String())
	}
}

func TestStatusRespectsThinkingOff(t *testing.T) {
	var out bytes.Buffer
	renderer := NewRenderer(&out, Config{Theme: "minimal", Thinking: "off"})

	renderer.RenderStatus("building context")
	renderer.RenderStatusSuccess("ready")

	if out.Len() != 0 {
		t.Fatalf("expected no status output, got %q", out.String())
	}
}
