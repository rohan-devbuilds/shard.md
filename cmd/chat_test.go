package cmd

import (
	"bytes"
	"strings"
	"testing"

	"shard/internal/memory"
	"shard/internal/project"
	"shard/internal/ui"
)

func TestParseRunAndReadCommands(t *testing.T) {
	run := parseSlashPayload("/run go test ./...", "/run")
	if run != "go test ./..." {
		t.Fatalf("run payload = %q", run)
	}
	read := parseSlashPayload("/read main.go", "/read")
	if read != "main.go" {
		t.Fatalf("read payload = %q", read)
	}
}

func TestContextBuilderIncludesProjectTreeAndCurrent(t *testing.T) {
	text := buildContextText(project.Tree{Root: "/repo", Files: []string{"main.go"}}, []memory.Item{{Name: "current.md", Content: "now"}}, "")
	if !strings.Contains(text, "Project root: /repo") {
		t.Fatalf("missing project root: %s", text)
	}
	if !strings.Contains(text, "- main.go") {
		t.Fatalf("missing file tree: %s", text)
	}
	if !strings.Contains(text, "current.md") {
		t.Fatalf("missing current memory: %s", text)
	}
}

func TestProviderCommandDoesNotRenderHeaderBox(t *testing.T) {
	var out bytes.Buffer
	renderer := ui.NewRenderer(&out, ui.Config{Theme: "minimal", Thinking: "minimal"})
	renderer.RenderHeader(ui.HeaderData{Model: "m", Provider: "p", Effort: "medium", ContextTokens: 1})
	renderer.RenderProvider("p", "m", "medium")

	if count := strings.Count(out.String(), "╭"); count != 1 {
		t.Fatalf("header-like boxes = %d, want 1\n%s", count, out.String())
	}
}

func TestToolContextCanHideRawToolOutput(t *testing.T) {
	var out bytes.Buffer
	renderer := ui.NewRenderer(&out, ui.Config{Theme: "minimal", Thinking: "minimal", TypingEffect: false})
	toolCtx := newToolContext(renderer, strings.NewReader(""))

	toolCtx.addResult("raw command output", false)
	injected := toolCtx.consumeResults()

	if injected != "raw command output" {
		t.Fatalf("injected = %q", injected)
	}
	if strings.Contains(out.String(), "raw command output") {
		t.Fatalf("hidden tool output was rendered: %s", out.String())
	}
}
