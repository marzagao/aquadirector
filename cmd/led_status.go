package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
)

var ledWatchInterval time.Duration

var ledStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show LED state",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithWatch(cmd.Context(), ledWatchInterval, func() error {
			client, err := getLEDClient()
			if err != nil {
				return err
			}

			manual, err := client.ManualState(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetching manual state: %w", err)
			}

			mode, _ := client.Mode(cmd.Context())

			format := output.ParseFormat(outputFmt)

			statusData := struct {
				Mode string `json:"mode" yaml:"mode"`
				*output.LEDStatusData
			}{
				Mode: mode,
				LEDStatusData: &output.LEDStatusData{
					White:       manual.White,
					Blue:        manual.Blue,
					Moon:        manual.Moon,
					Temperature: manual.Temperature,
				},
			}

			if format != output.Table {
				return output.Print(os.Stdout, format, statusData, nil)
			}

			rows := []output.TableRow{
				{Label: "Mode", Value: mode},
				{Label: "White", Value: fmt.Sprintf("%d", manual.White)},
				{Label: "Blue", Value: fmt.Sprintf("%d", manual.Blue)},
				{Label: "Moon", Value: fmt.Sprintf("%d", manual.Moon)},
				{Label: "LED Temp", Value: fmt.Sprintf("%.1fC", manual.Temperature)},
			}

			return output.Print(os.Stdout, format, statusData, rows)
		})
	},
}

func init() {
	addWatchFlag(ledStatusCmd, &ledWatchInterval)
	ledCmd.AddCommand(ledStatusCmd)
}
