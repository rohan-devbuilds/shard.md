package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"shard/internal/memory"
	"shard/internal/optimizer"
	"shard/internal/provider"
	"shard/internal/retrieval"
	"shard/internal/session"
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

		prov := provider.NewAnthropicProvider(modelName)
		sess := session.New()
		reader := bufio.NewScanner(os.Stdin)
		fmt.Println("Shard chat started. Type /exit to quit.")

		for {
			fmt.Print("> ")
			if !reader.Scan() {
				break
			}
			input := strings.TrimSpace(reader.Text())
			if input == "" {
				continue
			}

			if strings.HasPrefix(input, "/") {
				exit, err := handleSlash(context.Background(), input, sess, store, prov)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				if exit {
					return nil
				}
				continue
			}

			if err := runChatTurn(context.Background(), input, sess, store, prov); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				continue
			}

			if sess.PendingPairs() >= 10 {
				fmt.Println("Auto-optimizing session memory...")
				if err := optimizer.Run(context.Background(), sess, store, prov); err != nil {
					fmt.Fprintln(os.Stderr, "Optimize error:", err)
				}
			}
		}
		return reader.Err()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func handleSlash(ctx context.Context, input string, sess *session.Session, store *memory.Store, prov provider.Provider) (bool, error) {
	fields := strings.Fields(input)
	switch fields[0] {
	case "/exit":
		return true, nil
	case "/optimize":
		return false, optimizer.Run(ctx, sess, store, prov)
	case "/context":
		showContext("", sess, store)
		return false, nil
	case "/stats":
		ui.PrintStats(sess)
		return false, nil
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
		return false, fmt.Errorf("unknown slash command: %s", fields[0])
	}
}

func runChatTurn(ctx context.Context, input string, sess *session.Session, store *memory.Store, prov provider.Provider) error {
	files := retrieval.SelectFiles(input, sess.Focus, sess.Effort)
	memories, err := store.LoadFiles(files)
	if err != nil {
		return err
	}

	contextText := memory.FormatForPrompt(memories)
	contextTokens := session.EstimateTokens(contextText)
	ui.PrintHeader(prov.Name(), sess.Effort, contextTokens)

	messages := []provider.Message{
		{Role: "system", Content: buildSystemPrompt(sess.Effort, contextText)},
	}
	for _, msg := range sess.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			messages = append(messages, provider.Message{Role: msg.Role, Content: msg.Content})
		}
	}
	messages = append(messages, provider.Message{Role: "user", Content: input})

	sess.Add("user", input)
	reply, err := prov.Chat(ctx, messages)
	if err != nil {
		return err
	}
	fmt.Println(reply)
	sess.Add("assistant", reply)
	sess.CurrentContextTokens = contextTokens
	return nil
}

func buildSystemPrompt(effort string, contextText string) string {
	return fmt.Sprintf("You are Shard, an AI coding assistant using compressed project memory. Effort: %s.\n\nRelevant project memory:\n%s", effort, contextText)
}

func showContext(prompt string, sess *session.Session, store *memory.Store) {
	files := retrieval.SelectFiles(prompt, sess.Focus, sess.Effort)
	memories, err := store.LoadFiles(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	total := 0
	fmt.Println("Memory files:")
	for _, item := range memories {
		tokens := session.EstimateTokens(item.Content)
		total += tokens
		fmt.Printf("- %s (%d tokens)\n", item.Name, tokens)
	}
	fmt.Printf("Estimated context size: %d tokens\n", total)
}
