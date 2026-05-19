package cmd

import (
	"github.com/spf13/cobra"

	"shard/internal/session"
	"shard/internal/ui"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show session stats",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintStats(session.New())
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
