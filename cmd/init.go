package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"shard/internal/memory"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .shard memory files in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := memory.NewStore(".")
		if err := store.Init(); err != nil {
			return err
		}
		fmt.Println("Initialized .shard memory files.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
