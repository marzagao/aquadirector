package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

var (
	ledManualWhite int
	ledManualBlue  int
	ledManualMoon  int
)

var ledManualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Set LED manual output",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getLEDClient()
		if err != nil {
			return err
		}

		set := redsea.LEDManualSet{
			White: ledManualWhite,
			Blue:  ledManualBlue,
			Moon:  ledManualMoon,
		}

		if err := client.SetManual(cmd.Context(), set); err != nil {
			return fmt.Errorf("setting manual: %w", err)
		}

		fmt.Printf("LED set to white=%d blue=%d moon=%d\n", set.White, set.Blue, set.Moon)
		return nil
	},
}

func init() {
	ledManualCmd.Flags().IntVar(&ledManualWhite, "white", 0, "white channel (0-255)")
	ledManualCmd.Flags().IntVar(&ledManualBlue, "blue", 0, "blue channel (0-255)")
	ledManualCmd.Flags().IntVar(&ledManualMoon, "moon", 0, "moon channel (0-255)")
	ledCmd.AddCommand(ledManualCmd)
}
