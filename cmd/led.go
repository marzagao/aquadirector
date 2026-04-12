package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

var ledDevice string

var ledCmd = &cobra.Command{
	Use:   "led",
	Short: "ReefLED commands",
}

func init() {
	ledCmd.PersistentFlags().StringVar(&ledDevice, "device", "", "device name or IP")
	rootCmd.AddCommand(ledCmd)
}

func getLEDClient() (*redsea.LEDClient, error) {
	ip := resolveDeviceIP(ledDevice, "RSLED60")
	if ip == "" {
		// Try other LED models
		for _, model := range []string{"RSLED50", "RSLED90", "RSLED115", "RSLED160", "RSLED170"} {
			ip = resolveDeviceIP(ledDevice, model)
			if ip != "" {
				break
			}
		}
	}
	if ip == "" {
		return nil, fmt.Errorf("no LED device configured; use --device or add to config")
	}
	return redsea.NewLEDClient(ip, redsea.WithTimeout(appConfig.Network.DefaultTimeout),
		redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay)), nil
}
