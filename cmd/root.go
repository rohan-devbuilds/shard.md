package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	modelName string
	rootCmd   = &cobra.Command{
		Use:   "shard",
		Short: "Compress AI coding conversations into Markdown project memory",
		Long:  "Shard is a CLI that keeps durable AI coding context in categorized Markdown files.",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&modelName, "model", "claude-3-5-sonnet-20241022", "Anthropic model to use")
}
