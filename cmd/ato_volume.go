package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var atoVolumeSet int

var atoVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Get or set ATO reservoir volume",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getATOClient()
		if err != nil {
			return err
		}

		if atoVolumeSet > 0 {
			if err := client.SetVolume(cmd.Context(), atoVolumeSet); err != nil {
				return fmt.Errorf("setting volume: %w", err)
			}
			fmt.Printf("Volume set to %d ml.\n", atoVolumeSet)
			return nil
		}

		dash, err := client.Dashboard(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching dashboard: %w", err)
		}
		fmt.Printf("Volume left: %d ml\n", dash.VolumeLeft)
		return nil
	},
}

func init() {
	atoVolumeCmd.Flags().IntVar(&atoVolumeSet, "set", 0, "set reservoir volume in ml")
	atoCmd.AddCommand(atoVolumeCmd)
}
