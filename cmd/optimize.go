package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"shard/internal/memory"
	"shard/internal/optimizer"
	"shard/internal/provider"
	"shard/internal/session"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Compress the current session into .shard memory files",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := memory.NewStore(".")
		sess := session.New()
		prov, err := provider.New(providerName, modelName)
		if err != nil {
			return err
		}
		if _, err := optimizer.Run(context.Background(), sess, store, prov); err != nil {
			return err
		}
		fmt.Println("Optimization complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
}
