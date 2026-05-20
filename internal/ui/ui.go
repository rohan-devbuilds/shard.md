package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"shard/internal/permissions"
	"shard/internal/session"
)

type Renderer struct {
	Config Config
	Theme  Theme
	Out    io.Writer
}

type HeaderData struct {
	Model         string
	Provider      string
	Effort        string
	ContextTokens int
}

func NewRenderer(out io.Writer, cfg Config) *Renderer {
	return &Renderer{Config: cfg, Theme: NewTheme(cfg.Theme), Out: out}
}

func (r *Renderer) Prompt() {
	fmt.Fprint(r.Out, r.Theme.Prompt.Render("›")+" ")
}

func (r *Renderer) RenderHeader(data HeaderData) {
	body := fmt.Sprintf("%s\n%s | %s",
		r.Theme.Title.Render("Shard"),
		r.Theme.Metadata.Render(fmt.Sprintf("Model: %s | Provider: %s", data.Model, data.Provider)),
		r.Theme.Metadata.Render(fmt.Sprintf("Effort: %s | Context: ~%s tokens", data.Effort, HumanTokens(data.ContextTokens))),
	)
	fmt.Fprintln(r.Out, r.Theme.Box.Render(body))
}

func (r *Renderer) RenderProvider(provider string, model string, effort string) {
	fmt.Fprintf(r.Out, "%s %s | %s | effort: %s\n", r.Theme.Title.Render("Shard"), r.Theme.Metadata.Render("provider: "+provider), r.Theme.Metadata.Render("model: "+model), effort)
}

func (r *Renderer) RenderThinking(items []string) {
	if r.Config.Thinking == "off" {
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Metadata.Render("Thinking"))
	for _, item := range items {
		fmt.Fprintln(r.Out, r.Theme.Metadata.Render("  • "+item))
	}
	fmt.Fprintln(r.Out)
}

func (r *Renderer) RenderAssistant(text string) {
	if r.Config.TypingEffect {
		r.RenderStreamingAssistant(text, r.Config.TypingDelayMs)
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Assistant.Render("Shard"))
	fmt.Fprintln(r.Out, indent(strings.TrimSpace(text)))
	fmt.Fprintln(r.Out)
}

func (r *Renderer) RenderStreamingAssistant(text string, delayMs int) {
	fmt.Fprintln(r.Out, r.Theme.Assistant.Render("Shard"))
	fmt.Fprint(r.Out, "  ")
	for _, chunk := range streamingChunks(strings.TrimSpace(text)) {
		fmt.Fprint(r.Out, strings.ReplaceAll(chunk, "\n", "\n  "))
		flush(r.Out)
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
	fmt.Fprintln(r.Out)
	fmt.Fprintln(r.Out)
}

func (r *Renderer) RenderStatus(message string) {
	if r.Config.Thinking == "off" {
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Metadata.Render("• "+message))
}

func (r *Renderer) RenderDetailedStatus(message string) {
	if r.Config.Thinking != "full" {
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Metadata.Render("• "+message))
}

func (r *Renderer) RenderStatusSuccess(message string) {
	if r.Config.Thinking == "off" {
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Success.Render("✓ "+message))
}

func (r *Renderer) RenderFooter(contextTokens int, optimizeIn int) {
	if r.Config.Thinking == "off" {
		return
	}
	if optimizeIn < 0 {
		optimizeIn = 0
	}
	fmt.Fprintln(r.Out, r.Theme.Metadata.Render(fmt.Sprintf("tokens: ~%s context | optimize in %d messages", HumanTokens(contextTokens), optimizeIn)))
}

func (r *Renderer) RenderError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(r.Out, r.Theme.Error.Render("Error: "+cleanError(err.Error())))
	if strings.Contains(err.Error(), "API_KEY") {
		fmt.Fprintln(r.Out)
		fmt.Fprintln(r.Out, r.Theme.Metadata.Render("Set the provider API key in your shell or .env."))
	}
}

func (r *Renderer) RenderSuccess(message string) {
	fmt.Fprintln(r.Out, r.Theme.Success.Render("✓ "+message))
}

func (r *Renderer) RenderPermissionRequest(action permissions.Action, detail string) {
	body := fmt.Sprintf("Shard wants to %s:\n%s", permissionVerb(action), detail)
	fmt.Fprintln(r.Out, r.Theme.Box.Render(r.Theme.Title.Render("Permission Request")+"\n"+body))
	fmt.Fprintln(r.Out, "Allow? [y] yes  [n] no  [a] always this session")
}

func (r *Renderer) RenderPermissionApproved(action permissions.Action, detail string) {
	fmt.Fprintln(r.Out, r.Theme.Success.Render("✓ approved: "+permissionVerb(action)+": "+detail))
}

func (r *Renderer) RenderPermissionDenied() {
	fmt.Fprintln(r.Out, r.Theme.Error.Render("✗ permission denied"))
}

func (r *Renderer) RenderToolResult(result string) {
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "Tool result") {
		fmt.Fprintln(r.Out, r.Theme.Tool.Render("Tool result"))
		fmt.Fprintln(r.Out, strings.TrimSpace(strings.TrimPrefix(result, "Tool result:")))
	} else {
		fmt.Fprintln(r.Out, r.Theme.Tool.Render("Tool result"))
		fmt.Fprintln(r.Out, result)
	}
	fmt.Fprintln(r.Out)
}

func (r *Renderer) RenderStats(sess *session.Session, provider string, model string) {
	focus := sess.Focus
	if focus == "" {
		focus = "none"
	}
	lines := []string{
		"Session Stats",
		fmt.Sprintf("- messages: %d", len(sess.Messages)),
		fmt.Sprintf("- optimizations: %d", sess.OptimizeCount),
		fmt.Sprintf("- effort: %s", sess.Effort),
		fmt.Sprintf("- focus: %s", focus),
		fmt.Sprintf("- estimated tokens: ~%s", HumanTokens(sess.EstimatedAPITokens())),
		fmt.Sprintf("- provider: %s", provider),
		fmt.Sprintf("- model: %s", model),
	}
	fmt.Fprintln(r.Out, r.Theme.Box.Render(strings.Join(lines, "\n")))
}

func (r *Renderer) RenderPermissions(statuses map[permissions.Action]string) {
	lines := []string{"Session Permissions"}
	for _, action := range permissions.Actions() {
		lines = append(lines, fmt.Sprintf("- %s: %s", action, statuses[action]))
	}
	fmt.Fprintln(r.Out, r.Theme.Box.Render(strings.Join(lines, "\n")))
}

func (r *Renderer) RenderContext(loaded []string, skipped []string, estimated int) {
	var b strings.Builder
	b.WriteString(r.Theme.Title.Render("Loaded Memory") + "\n")
	for _, file := range loaded {
		b.WriteString(r.Theme.Success.Render("✓ "+file) + "\n")
	}
	if len(skipped) > 0 {
		b.WriteString("\n" + r.Theme.Metadata.Render("Skipped") + "\n")
		for _, file := range skipped {
			b.WriteString("- " + file + "\n")
		}
	}
	b.WriteString(fmt.Sprintf("\nEstimated context: ~%s tokens", HumanTokens(estimated)))
	fmt.Fprintln(r.Out, r.Theme.Box.Render(strings.TrimSpace(b.String())))
}

func (r *Renderer) RenderContextDetails(root string, fileCount int, loaded []string, skipped []string, recentMessages int, estimated int) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Project root: %s\n", root))
	b.WriteString(fmt.Sprintf("Files indexed: %d\n\n", fileCount))
	b.WriteString(r.Theme.Title.Render("Loaded memory") + "\n")
	for _, file := range loaded {
		b.WriteString(r.Theme.Success.Render("✓ "+file) + "\n")
	}
	if len(skipped) > 0 {
		b.WriteString("\n" + r.Theme.Metadata.Render("Skipped") + "\n")
		for _, file := range skipped {
			b.WriteString("- " + file + "\n")
		}
	}
	b.WriteString(fmt.Sprintf("\nRecent messages: %d", recentMessages))
	b.WriteString(fmt.Sprintf("\nEstimated tokens: ~%s", HumanTokens(estimated)))
	fmt.Fprintln(r.Out, r.Theme.Box.Render(strings.TrimSpace(b.String())))
}

func (r *Renderer) RenderOptimizeSummary(messages int, before int, after int, updated []string) {
	fmt.Fprintln(r.Out, r.Theme.Metadata.Render("Optimizing memory..."))
	fmt.Fprintf(r.Out, "messages compressed: %d | before: ~%s tokens | after: ~%s tokens\n", messages, HumanTokens(before), HumanTokens(after))
	if len(updated) == 0 {
		fmt.Fprintln(r.Out, r.Theme.Warning.Render("No durable memory updates found."))
		return
	}
	fmt.Fprintln(r.Out, "Updated memory:")
	for _, file := range updated {
		fmt.Fprintln(r.Out, r.Theme.Success.Render("✓ "+file))
	}
}

func (r *Renderer) RenderHelp() {
	rows := []string{
		"Available commands",
		"",
		"/optimize      compress recent session into memory",
		"/context       show loaded memory files",
		"/stats         show session stats",
		"/effort        set effort: low, medium, high, max",
		"/focus         focus memory category",
		"/provider      show provider and model",
		"/thinking      set thinking: off, minimal, full",
		"/run           run shell command with approval",
		"/read          read file with approval",
		"/read-all      read indexed project files with approval",
		"/permissions   show session permissions",
		"/exit          save and exit",
	}
	fmt.Fprintln(r.Out, r.Theme.Box.Render(strings.Join(rows, "\n")))
}

func HumanTokens(tokens int) string {
	if tokens >= 1000 {
		return fmt.Sprintf("%.1fk", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

func indent(text string) string {
	if text == "" {
		return "  "
	}
	return "  " + strings.ReplaceAll(text, "\n", "\n  ")
}

func streamingChunks(text string) []string {
	if text == "" {
		return []string{""}
	}
	chunks := []string{}
	var b strings.Builder
	for _, r := range text {
		b.WriteRune(r)
		if r == ' ' || r == '\n' || r == '\t' || b.Len() >= 24 {
			chunks = append(chunks, b.String())
			b.Reset()
		}
	}
	if b.Len() > 0 {
		chunks = append(chunks, b.String())
	}
	return chunks
}

func flush(out io.Writer) {
	if f, ok := out.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
}

func cleanError(text string) string {
	text = strings.TrimPrefix(text, "anthropic error: ")
	text = strings.TrimPrefix(text, "openrouter error: ")
	return text
}

func permissionVerb(action permissions.Action) string {
	switch action {
	case permissions.ReadFile:
		return "read file"
	case permissions.WriteFile:
		return "write file"
	case permissions.ListDir:
		return "list directory"
	case permissions.RunCommand:
		return "run"
	default:
		return string(action)
	}
}
