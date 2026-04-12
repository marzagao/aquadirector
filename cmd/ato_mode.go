package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var atoModeSet string

var atoModeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get or set ATO operating mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getATOClient()
		if err != nil {
			return err
		}

		if atoModeSet != "" {
			if err := client.SetMode(cmd.Context(), atoModeSet); err != nil {
				return fmt.Errorf("setting mode: %w", err)
			}
			fmt.Printf("Mode set to %s.\n", atoModeSet)
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
	atoModeCmd.Flags().StringVar(&atoModeSet, "set", "", "set mode: auto, manual, empty")
	atoCmd.AddCommand(atoModeCmd)
}
