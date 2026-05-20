package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"shard/internal/memory"
	"shard/internal/optimizer"
	"shard/internal/permissions"
	"shard/internal/project"
	"shard/internal/provider"
	"shard/internal/retrieval"
	"shard/internal/session"
	"shard/internal/tools"
	"shard/internal/ui"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive Shard chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := memory.NewStore(".")
		if err := store.Init(); err != nil {
			return err
		}

		selectedProvider, err := chooseProvider(providerName)
		if err != nil {
			return err
		}
		prov, err := provider.New(selectedProvider, modelName)
		if err != nil {
			return err
		}
		sess := session.New()
		tree, err := project.Scan(".")
		if err != nil {
			return err
		}
		renderer := ui.NewRenderer(os.Stdout, ui.LoadConfig("."))
		reader := bufio.NewReader(os.Stdin)
		toolCtx := newToolContext(renderer, reader)
		renderer.RenderHeader(ui.HeaderData{Model: prov.Name(), Provider: prov.Provider(), Effort: sess.Effort, ContextTokens: session.EstimateTokens(tree.Format(500))})
		renderer.RenderDetailedStatus(fmt.Sprintf("scanned project tree: %d files", len(tree.Files)))
		renderer.RenderSuccess("Shard chat started. Type /exit to quit.")

		for {
			fmt.Fprintln(os.Stdout)
			renderer.Prompt()
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			input := strings.TrimSpace(line)
			if input == "" {
				continue
			}

			if strings.HasPrefix(input, "/") {
				exit, err := handleSlash(context.Background(), input, sess, store, prov, toolCtx, renderer, tree)
				if err != nil {
					renderer.RenderError(err)
				}
				if exit {
					return nil
				}
				continue
			}

			if err := runChatTurn(context.Background(), input, sess, store, prov, toolCtx, tree); err != nil {
				renderer.RenderError(err)
				continue
			}

			if sess.PendingPairs() >= 10 {
				fmt.Println("Auto-optimizing session memory...")
				if summary, err := optimizer.Run(context.Background(), sess, store, prov); err != nil {
					renderer.RenderError(err)
				} else {
					renderer.RenderOptimizeSummary(summary.MessagesCompressed, summary.BeforeTokens, summary.AfterTokens, summary.UpdatedFiles)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func chooseProvider(configured string) (string, error) {
	if strings.TrimSpace(configured) != "" || strings.TrimSpace(provider.EnvValue("SHARD_PROVIDER")) != "" {
		return configured, nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return configured, nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Choose provider:")
		fmt.Println("1) OpenRouter")
		fmt.Println("2) Anthropic")
		fmt.Print("Provider [1]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		switch strings.ToLower(strings.TrimSpace(input)) {
		case "", "1", "openrouter":
			return "openrouter", nil
		case "2", "anthropic":
			return "anthropic", nil
		default:
			fmt.Println("Please choose 1/openrouter or 2/anthropic.")
		}
	}
}

type toolContext struct {
	permissions *permissions.Manager
	files       *tools.FileTools
	shell       *tools.Shell
	renderer    *ui.Renderer
	results     []toolResult
}

type toolResult struct {
	Text    string
	Visible bool
}

func newToolContext(renderer *ui.Renderer, in io.Reader) *toolContext {
	perms := permissions.NewManager()
	perms.UI = renderer
	return &toolContext{
		permissions: perms,
		renderer:    renderer,
		files: &tools.FileTools{
			Permissions: perms,
			In:          in,
			Out:         os.Stdout,
		},
		shell: &tools.Shell{
			Permissions: perms,
			In:          in,
			Out:         os.Stdout,
		},
	}
}

func (t *toolContext) addResult(result string, visible bool) {
	t.results = append(t.results, toolResult{Text: result, Visible: visible})
	if visible {
		t.renderer.RenderToolResult(tools.Truncate(result, 1200))
	}
}

func (t *toolContext) consumeResults() string {
	if len(t.results) == 0 {
		return ""
	}
	parts := make([]string, 0, len(t.results))
	for _, result := range t.results {
		parts = append(parts, result.Text)
	}
	out := strings.Join(parts, "\n\n")
	t.results = nil
	return out
}

func handleSlash(ctx context.Context, input string, sess *session.Session, store *memory.Store, prov provider.Provider, toolCtx *toolContext, renderer *ui.Renderer, tree project.Tree) (bool, error) {
	fields := strings.Fields(input)
	switch fields[0] {
	case "/help":
		renderer.RenderHelp()
		return false, nil
	case "/exit":
		return true, nil
	case "/optimize":
		summary, err := optimizer.Run(ctx, sess, store, prov)
		if err == nil {
			renderer.RenderOptimizeSummary(summary.MessagesCompressed, summary.BeforeTokens, summary.AfterTokens, summary.UpdatedFiles)
		}
		return false, err
	case "/context":
		showContext("", sess, store, renderer, tree)
		return false, nil
	case "/stats":
		renderer.RenderStats(sess, prov.Provider(), prov.Name())
		return false, nil
	case "/provider":
		renderer.RenderProvider(prov.Provider(), prov.Name(), sess.Effort)
		return false, nil
	case "/thinking":
		if len(fields) != 2 {
			return false, fmt.Errorf("usage: /thinking off|minimal|full")
		}
		value := strings.ToLower(fields[1])
		if value != "off" && value != "minimal" && value != "full" {
			return false, fmt.Errorf("usage: /thinking off|minimal|full")
		}
		renderer.Config.Thinking = value
		renderer.RenderSuccess("thinking: " + value)
		return false, nil
	case "/permissions":
		printPermissions(toolCtx.permissions, renderer)
		return false, nil
	case "/run":
		command := parseSlashPayload(input, "/run")
		return false, runCommandTool(command, toolCtx)
	case "/read":
		path := parseSlashPayload(input, "/read")
		return false, readFileTool(path, toolCtx)
	case "/read-all":
		return false, readProjectTool(tree, toolCtx)
	case "/effort":
		if len(fields) != 2 {
			return false, fmt.Errorf("usage: /effort low|medium|high|max")
		}
		return false, sess.SetEffort(fields[1])
	case "/focus":
		if len(fields) < 2 {
			sess.Focus = ""
			fmt.Println("Focus cleared.")
			return false, nil
		}
		focus := strings.ToLower(strings.TrimSpace(strings.Join(fields[1:], " ")))
		if !memory.IsCategory(focus) {
			return false, fmt.Errorf("unknown category %q", focus)
		}
		sess.Focus = focus
		fmt.Println("Focus:", sess.Focus)
		return false, nil
	default:
		renderer.RenderHelp()
		return false, fmt.Errorf("unknown slash command: %s", fields[0])
	}
}

func parseSlashPayload(input string, command string) string {
	return strings.TrimSpace(strings.TrimPrefix(input, command))
}

func runChatTurn(ctx context.Context, input string, sess *session.Session, store *memory.Store, prov provider.Provider, toolCtx *toolContext, tree project.Tree) error {
	files := retrieval.SelectFiles(input, sess.Focus, sess.Effort)
	toolCtx.renderer.RenderDetailedStatus("loading memory: " + strings.Join(files, ", "))
	memories, err := store.LoadFiles(files)
	if err != nil {
		return err
	}

	contextText := buildContextText(tree, memories, "")
	toolResults := toolCtx.consumeResults()
	if toolResults != "" {
		contextText += "\n\n" + toolResults
	}
	contextTokens := session.EstimateTokens(contextText)
	uiFiles := make([]string, 0, len(memories))
	for _, item := range memories {
		uiFiles = append(uiFiles, item.Name)
	}
	thinking := []string{"building context", "calling model"}
	if toolCtx.renderer.Config.Thinking == "full" {
		thinking = []string{
			"task: " + classifyTask(input),
			"memory: " + strings.Join(uiFiles, ", "),
			"provider: " + prov.Provider(),
			"model: " + prov.Name(),
			"effort: " + sess.Effort,
			"context: ~" + ui.HumanTokens(contextTokens) + " tokens",
		}
	}
	toolCtx.renderer.RenderThinking(thinking)
	toolCtx.renderer.RenderStatus("building context: ~" + ui.HumanTokens(contextTokens) + " tokens")
	toolCtx.renderer.RenderStatus("calling model: " + prov.Name())

	messages := []provider.Message{
		{Role: "system", Content: buildSystemPrompt(sess.Effort, contextText)},
	}
	for _, msg := range sess.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			messages = append(messages, provider.Message{Role: msg.Role, Content: msg.Content})
		}
	}
	messages = append(messages, provider.Message{Role: "user", Content: input})
	inputTokens := estimateProviderMessages(messages)

	sess.Add("user", input)
	forcedReq, forceTool := tools.InferToolRequest(input)
	reply, err := prov.Chat(ctx, messages)
	if err != nil {
		return err
	}
	for toolCalls := 0; toolCalls < 5; toolCalls++ {
		req, ok := tools.ParseToolRequest(reply)
		if !ok && forceTool {
			req = forcedReq
			ok = true
			forceTool = false
		}
		if !ok {
			break
		}
		result, err := executeToolRequest(req, toolCtx)
		if err != nil {
			return err
		}
		toolCtx.addResult(result, false)
		contextText += "\n\n" + result
		messages[0] = provider.Message{Role: "system", Content: buildSystemPrompt(sess.Effort, contextText)}
		messages = append(messages, provider.Message{Role: "assistant", Content: reply})
		messages = append(messages, provider.Message{Role: "user", Content: "Tool result provided. Explain the result to the user in plain language and answer the original request. Summarize relevant findings; do not dump raw tool output unless it is short and directly useful. If another tool is required, request it. Otherwise provide the final answer now."})
		inputTokens += estimateProviderMessages(messages)
		reply, err = prov.Chat(ctx, messages)
		if err != nil {
			return err
		}
	}
	if _, ok := tools.ParseToolRequest(reply); ok {
		return fmt.Errorf("model kept requesting tools after the safety limit")
	}
	toolCtx.renderer.RenderStatusSuccess("response ready")
	toolCtx.renderer.RenderAssistant(reply)
	toolCtx.renderer.RenderFooter(contextTokens, 10-sess.PendingPairs()-1)
	sess.Add("assistant", reply)
	sess.RecordModelCall(inputTokens, session.EstimateTokens(reply))
	sess.CurrentContextTokens = contextTokens
	return nil
}

func buildContextText(tree project.Tree, memories []memory.Item, toolResults string) string {
	contextText := tree.Format(500) + "\n\n" + memory.FormatForPrompt(memories)
	if strings.TrimSpace(toolResults) != "" {
		contextText += "\n\n" + strings.TrimSpace(toolResults)
	}
	return contextText
}

func classifyTask(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "bug"), strings.Contains(lower, "fix"), strings.Contains(lower, "error"), strings.Contains(lower, "fail"):
		return "debugging"
	case strings.Contains(lower, "design"), strings.Contains(lower, "architecture"):
		return "architecture"
	case strings.Contains(lower, "ui"), strings.Contains(lower, "theme"):
		return "ui"
	default:
		return "coding"
	}
}

func executeToolRequest(req tools.Request, toolCtx *toolContext) (string, error) {
	switch req.Action {
	case "read_file":
		toolCtx.renderer.RenderStatus("reading file: " + req.Arg)
		content, err := toolCtx.files.ReadFile(req.Arg)
		if err != nil {
			return "", err
		}
		toolCtx.renderer.RenderStatusSuccess(fmt.Sprintf("read file: %s (~%s tokens)", req.Arg, ui.HumanTokens(session.EstimateTokens(content))))
		return fmt.Sprintf("Tool result:\nread file: %s\ncontent:\n%s", req.Arg, content), nil
	case "read_project":
		return readProjectToolResult(req.Arg, toolCtx)
	case "list_dir":
		toolCtx.renderer.RenderStatus("listing directory: " + req.Arg)
		content, err := toolCtx.files.ListDir(req.Arg)
		if err != nil {
			return "", err
		}
		toolCtx.renderer.RenderStatusSuccess(fmt.Sprintf("listed directory: %s (%d entries)", req.Arg, countLines(content)))
		return fmt.Sprintf("Tool result:\nlist dir: %s\nentries:\n%s", req.Arg, content), nil
	case "run_command":
		toolCtx.renderer.RenderStatus("running command: " + req.Arg)
		result, err := toolCtx.shell.Run(req.Arg)
		if err != nil {
			return "", err
		}
		if result.ExitCode == 0 {
			toolCtx.renderer.RenderStatusSuccess("command completed: " + req.Arg)
		} else {
			toolCtx.renderer.RenderError(fmt.Errorf("command failed: %s", req.Arg))
		}
		return tools.FormatCommandResult(result), nil
	default:
		return "", fmt.Errorf("unsupported tool request: %s", req.Action)
	}
}
func runCommandTool(command string, toolCtx *toolContext) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("usage: /run <command>")
	}
	toolCtx.renderer.RenderStatus("running command: " + command)
	result, err := toolCtx.shell.Run(command)
	if err != nil {
		return err
	}
	if result.ExitCode == 0 {
		toolCtx.renderer.RenderStatusSuccess("command completed: " + command)
	} else {
		toolCtx.renderer.RenderError(fmt.Errorf("command failed: %s", command))
	}
	toolCtx.addResult(tools.FormatCommandResult(result), true)
	return nil
}

func readFileTool(path string, toolCtx *toolContext) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("usage: /read <file>")
	}
	toolCtx.renderer.RenderStatus("reading file: " + path)
	content, err := toolCtx.files.ReadFile(path)
	if err != nil {
		return err
	}
	toolCtx.renderer.RenderStatusSuccess(fmt.Sprintf("read file: %s (~%s tokens)", path, ui.HumanTokens(session.EstimateTokens(content))))
	toolCtx.addResult(fmt.Sprintf("Tool result:\nread file: %s\ncontent:\n%s", path, content), true)
	return nil
}

func readProjectTool(tree project.Tree, toolCtx *toolContext) error {
	result, err := readProjectToolResult(tree.Root, toolCtx)
	if err != nil {
		return err
	}
	toolCtx.addResult(result, true)
	return nil
}

func readProjectToolResult(root string, toolCtx *toolContext) (string, error) {
	toolCtx.renderer.RenderStatus("reading project files: " + root)
	tree, err := project.Scan(root)
	if err != nil {
		return "", err
	}
	content, err := toolCtx.files.ReadProjectFiles(tree.Root, tree.Files)
	if err != nil {
		return "", err
	}
	toolCtx.renderer.RenderStatusSuccess(fmt.Sprintf("read project: %d files (~%s tokens)", len(tree.Files), ui.HumanTokens(session.EstimateTokens(content))))
	return fmt.Sprintf("Tool result:\nread project: %s\nfiles: %d\ncontent:\n%s", tree.Root, len(tree.Files), content), nil
}

func listDirTool(path string, toolCtx *toolContext) error {
	content, err := toolCtx.files.ListDir(path)
	if err != nil {
		return err
	}
	toolCtx.addResult(fmt.Sprintf("Tool result:\nlist dir: %s\nentries:\n%s", path, content), true)
	return nil
}

func countLines(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

func printPermissions(perms *permissions.Manager, renderer *ui.Renderer) {
	statuses := map[permissions.Action]string{}
	for _, action := range permissions.Actions() {
		statuses[action] = perms.Status(action)
	}
	renderer.RenderPermissions(statuses)
}

func buildSystemPrompt(effort string, contextText string) string {
	return fmt.Sprintf(`You are Shard, an AI coding CLI working inside a real project directory.

Response style:
- Use normal professional engineering language.
- Do not use caveman-style phrasing, roleplay, baby talk, or joke dialects.
- Be concise, direct, and specific.

You have access to:
- Project file structure
- Shard memory files
- Recent conversation
- Approved tool results

Rules:
- Do not assume file contents you have not seen
- If needed, request file reads or commands using only a tool_request block
- If broad project analysis is needed, request read_project
- If you are about to say you will list/read/run/check something, request the tool instead of narrating the plan
- For action requests, do the action through tools. Do not merely describe what you would do.
- After tool results are provided, explain the outcome clearly for a developer. Do not dump unexplained raw output.
- If you request a tool, do not also provide a final answer in the same message
- Be precise and actionable
- Focus on solving developer tasks
- Effort: %s

Tool request format:
`+"```tool_request\nread_file: path/to/file.go\n```"+`
`+"```tool_request\nread_project: .\n```"+`
`+"```tool_request\nlist_dir: .\n```"+`
`+"```tool_request\nrun_command: go test ./...\n```"+`

Context:
%s`, effort, contextText)
}

func estimateProviderMessages(messages []provider.Message) int {
	total := 0
	for _, msg := range messages {
		total += session.EstimateTokens(msg.Role)
		total += session.EstimateTokens(msg.Content)
		total += 4
	}
	return total
}

func showContext(prompt string, sess *session.Session, store *memory.Store, renderer *ui.Renderer, tree project.Tree) {
	files := retrieval.SelectFiles(prompt, sess.Focus, sess.Effort)
	memories, err := store.LoadFiles(files)
	if err != nil {
		renderer.RenderError(err)
		return
	}
	total := 0
	loaded := []string{}
	for _, item := range memories {
		tokens := session.EstimateTokens(item.Content)
		total += tokens
		loaded = append(loaded, item.Name)
	}
	skipped := []string{}
	loadedSet := map[string]bool{}
	for _, file := range loaded {
		loadedSet[file] = true
	}
	for _, category := range memory.Categories {
		file := category + ".md"
		if !loadedSet[file] {
			skipped = append(skipped, file)
		}
	}
	renderer.RenderContextDetails(tree.Root, len(tree.Files), loaded, skipped, len(sess.Messages), total+session.EstimateTokens(tree.Format(500)))
}
