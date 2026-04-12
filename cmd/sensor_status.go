package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/internal/sensor"
	"github.com/spf13/cobra"
)

var sensorWatchInterval time.Duration

var sensorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Read current water parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithWatch(cmd.Context(), sensorWatchInterval, func() error {
			configuredIP := getSensorIP()
			if configuredIP == "" {
				return fmt.Errorf("no sensor IP configured; use --ip or add to config")
			}

			deviceID := appConfig.Sensor.DeviceID
			localKey := appConfig.Sensor.LocalKey
			version := appConfig.Sensor.Version

			if deviceID == "" || localKey == "" {
				return fmt.Errorf("sensor device_id and local_key required in config; see aquadirector.yaml.example")
			}

			// Resolve IP: try configured IP first, fall back to MAC-based ARP discovery
			ip, rediscovered := sensor.ResolveIP(configuredIP, appConfig.Network.Subnet)
			if rediscovered {
				fmt.Fprintf(os.Stderr, "Device moved: found at %s (was %s)\n", ip, configuredIP)
			}

			client := sensor.NewClient(ip, deviceID, localKey, version, appConfig.Sensor.Calibration)
			wq, err := client.ReadWaterQuality(cmd.Context())
			if err != nil {
				return fmt.Errorf("reading sensor: %w", err)
			}

			format := output.ParseFormat(outputFmt)
			if format != output.Table {
				return output.Print(os.Stdout, format, wq, nil)
			}

			rows := []output.TableRow{
				{Label: "pH", Value: fmt.Sprintf("%.2f", wq.PH)},
				{Label: "Temperature", Value: fmt.Sprintf("%.1fC", wq.Temperature)},
				{Label: "TDS", Value: fmt.Sprintf("%d ppm", wq.TDS)},
				{Label: "EC", Value: fmt.Sprintf("%d uS/cm", wq.EC)},
				{Label: "ORP", Value: fmt.Sprintf("%d mV", wq.ORP)},
				{Label: "Salinity", Value: fmt.Sprintf("%.2f%%", wq.Salinity)},
				{Label: "SG", Value: fmt.Sprintf("%.3f", wq.SG)},
				{Label: "Battery", Value: fmt.Sprintf("%d%%", wq.Battery)},
			}

			return output.Print(os.Stdout, format, wq, rows)
		})
	},
}

func init() {
	addWatchFlag(sensorStatusCmd, &sensorWatchInterval)
	sensorCmd.AddCommand(sensorStatusCmd)
}
