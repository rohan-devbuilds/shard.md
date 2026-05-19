package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"shard/internal/memory"
	"shard/internal/retrieval"
	"shard/internal/session"
)

var contextCmd = &cobra.Command{
	Use:   "context [prompt]",
	Short: "Show which memory files would be loaded",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := memory.NewStore(".")
		if err := store.Init(); err != nil {
			return err
		}
		sess := session.New()
		prompt := strings.Join(args, " ")
		files := retrieval.SelectFiles(prompt, sess.Focus, sess.Effort)
		memories, err := store.LoadFiles(files)
		if err != nil {
			return err
		}
		total := 0
		fmt.Println("Memory files:")
		for _, item := range memories {
			tokens := session.EstimateTokens(item.Content)
			total += tokens
			fmt.Printf("- %s (%d tokens)\n", item.Name, tokens)
		}
		fmt.Printf("Estimated context size: %d tokens\n", total)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
