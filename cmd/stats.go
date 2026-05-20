package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"shard/internal/session"
	"shard/internal/ui"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show session stats",
	Run: func(cmd *cobra.Command, args []string) {
		renderer := ui.NewRenderer(os.Stdout, ui.LoadConfig("."))
		renderer.RenderStats(session.New(), "none", "none")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
