package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ledModeSet string

var ledModeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get or set LED operating mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getLEDClient()
		if err != nil {
			return err
		}

		if ledModeSet != "" {
			if err := client.SetMode(cmd.Context(), ledModeSet); err != nil {
				return fmt.Errorf("setting mode: %w", err)
			}
			fmt.Printf("Mode set to %s.\n", ledModeSet)
			return nil
		}

		mode, err := client.Mode(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching mode: %w", err)
		}
		fmt.Printf("Mode: %s\n", mode)
		return nil
	},
}

func init() {
	ledModeCmd.Flags().StringVar(&ledModeSet, "set", "", "set mode: auto, manual, timer")
	ledCmd.AddCommand(ledModeCmd)
}
