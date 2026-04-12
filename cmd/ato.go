package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

var atoDevice string

var atoCmd = &cobra.Command{
	Use:   "ato",
	Short: "ReefATO+ commands",
}

func init() {
	atoCmd.PersistentFlags().StringVar(&atoDevice, "device", "", "device name or IP")
	rootCmd.AddCommand(atoCmd)
}

func getATOClient() (*redsea.ATOClient, error) {
	ip := resolveDeviceIP(atoDevice, "RSATO+")
	if ip == "" {
		return nil, fmt.Errorf("no ATO device configured; use --device or add to config")
	}
	return redsea.NewATOClient(ip, redsea.WithTimeout(appConfig.Network.DefaultTimeout),
		redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay)), nil
}

func resolveDeviceIP(nameOrIP, deviceType string) string {
	if nameOrIP != "" {
		if dev := appConfig.DeviceByName(nameOrIP); dev != nil {
			return dev.IP
		}
		return nameOrIP // assume it's an IP
	}
	if dev := appConfig.DeviceByType(deviceType); dev != nil {
		return dev.IP
	}
	return ""
}
