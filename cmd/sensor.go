package cmd

import (
	"github.com/spf13/cobra"
)

var sensorIP string

var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Kactoily water sensor commands",
}

func init() {
	sensorCmd.PersistentFlags().StringVar(&sensorIP, "ip", "", "sensor IP address (default from config)")
	rootCmd.AddCommand(sensorCmd)
}

func getSensorIP() string {
	if sensorIP != "" {
		return sensorIP
	}
	return appConfig.Sensor.IP
}
