package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var feederDrumCmd = &cobra.Command{
	Use:   "drum",
	Short: "Drum management",
}

var feederDrumFullCmd = &cobra.Command{
	Use:   "full",
	Short: "Mark drum as full (resets fill level tracking)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getFeederClient(cmd.Context())
		if err != nil {
			return err
		}
		if err := client.MarkDrumFull(cmd.Context()); err != nil {
			return fmt.Errorf("marking drum full: %w", err)
		}
		fmt.Println("Drum marked as full.")
		return nil
	},
}

func init() {
	feederDrumCmd.AddCommand(feederDrumFullCmd)
	feederCmd.AddCommand(feederDrumCmd)
}
