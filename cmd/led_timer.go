package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

var (
	ledTimerWhite    int
	ledTimerBlue     int
	ledTimerMoon     int
	ledTimerDuration int
)

var ledTimerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Set LED temporary override with duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getLEDClient()
		if err != nil {
			return err
		}

		set := redsea.LEDTimerSet{
			White:    ledTimerWhite,
			Blue:     ledTimerBlue,
			Moon:     ledTimerMoon,
			Duration: ledTimerDuration * 1000, // seconds to milliseconds
		}

		if err := client.SetTimer(cmd.Context(), set); err != nil {
			return fmt.Errorf("setting timer: %w", err)
		}

		fmt.Printf("LED timer set: white=%d blue=%d moon=%d for %ds\n",
			set.White, set.Blue, set.Moon, ledTimerDuration)
		return nil
	},
}

func init() {
	ledTimerCmd.Flags().IntVar(&ledTimerWhite, "white", 0, "white channel (0-255)")
	ledTimerCmd.Flags().IntVar(&ledTimerBlue, "blue", 0, "blue channel (0-255)")
	ledTimerCmd.Flags().IntVar(&ledTimerMoon, "moon", 0, "moon channel (0-255)")
	ledTimerCmd.Flags().IntVar(&ledTimerDuration, "duration", 3600, "duration in seconds")
	ledCmd.AddCommand(ledTimerCmd)
}
