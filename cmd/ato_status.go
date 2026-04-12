package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
)

var atoWatchInterval time.Duration

var atoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ATO dashboard and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithWatch(cmd.Context(), atoWatchInterval, func() error {
			client, err := getATOClient()
			if err != nil {
				return err
			}

			dash, err := client.Dashboard(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetching dashboard: %w", err)
			}

			format := output.ParseFormat(outputFmt)
			if format != output.Table {
				return output.Print(os.Stdout, format, dash, nil)
			}

			lastFill := "never"
			if dash.LastFillDate != nil {
				lastFill = time.Unix(*dash.LastFillDate, 0).Format(time.RFC3339)
			}

			tempStr := "N/A (probe not connected)"
			if dash.HasTemperature() {
				tempStr = fmt.Sprintf("%.1fC / %.1fF", dash.Temperature(), dash.Temperature()*9/5+32)
			}

			rows := []output.TableRow{
				{Label: "Mode", Value: dash.Mode},
				{Label: "Water Level", Value: dash.WaterLevel},
				{Label: "Volume Left", Value: fmt.Sprintf("%d ml", dash.VolumeLeft)},
				{Label: "Pump", Value: fmt.Sprintf("%s (on=%v, speed=%d%%)", dash.PumpState, dash.IsPumpOn, dash.PumpSpeed)},
				{Label: "Flow Rate", Value: fmt.Sprintf("%.0f ml/min", dash.FlowRate)},
				{Label: "Total Fills", Value: fmt.Sprintf("%d", dash.TotalFills)},
				{Label: "Today Fills", Value: fmt.Sprintf("%d (%d ml)", dash.TodayFills, dash.TodayVolumeUsage)},
				{Label: "Last Fill", Value: lastFill},
				{Label: "ATO Sensor", Value: fmt.Sprintf("level=%s connected=%v", dash.ATOSensor.CurrentLevel, dash.ATOSensor.Connected)},
				{Label: "Leak Sensor", Value: fmt.Sprintf("status=%s enabled=%v", dash.LeakSensor.Status, dash.LeakSensor.Enabled)},
				{Label: "Temperature", Value: tempStr},
			}

			return output.Print(os.Stdout, format, dash, rows)
		})
	},
}

func init() {
	addWatchFlag(atoStatusCmd, &atoWatchInterval)
	atoCmd.AddCommand(atoStatusCmd)
}
