package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/marzagao/aquadirector/internal/sensor"
	"gopkg.in/yaml.v3"
)

var (
	rekeyClientID     string
	rekeyClientSecret string
)

var sensorRekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Fetch fresh local key from Tuya Cloud and update config",
	Long: `Fetches the current local key for the Kactoily sensor from the Tuya IoT
Developer Platform and updates the config file.

Requires Tuya Cloud API credentials (Access ID and Access Secret from iot.tuya.com).
Your Smart Life account must be linked to the Tuya IoT project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rekeyClientID == "" || rekeyClientSecret == "" {
			return fmt.Errorf("--client-id and --client-secret are required\n\nGet these from https://iot.tuya.com/ > Cloud > Development > your project")
		}

		deviceID := appConfig.Sensor.DeviceID
		if deviceID == "" {
			return fmt.Errorf("sensor.device_id not set in config")
		}

		fmt.Fprintf(os.Stderr, "Fetching devices from Tuya Cloud...\n")

		cloud := sensor.NewTuyaCloud(rekeyClientID, rekeyClientSecret)
		devices, err := cloud.GetDevices()
		if err != nil {
			return fmt.Errorf("fetching devices: %w", err)
		}

		if len(devices) == 0 {
			return fmt.Errorf("no devices found; re-link your Smart Life account at iot.tuya.com > Devices > Link Tuya App Account")
		}

		// Find our sensor
		var found *sensor.TuyaDeviceInfo
		for i, d := range devices {
			if d.ID == deviceID {
				found = &devices[i]
				break
			}
		}

		if found == nil {
			fmt.Fprintf(os.Stderr, "Device %s not found. Available devices:\n", deviceID)
			for _, d := range devices {
				fmt.Fprintf(os.Stderr, "  %s  %s  (key: %s)\n", d.ID, d.Name, d.LocalKey)
			}
			return fmt.Errorf("device %s not in linked account", deviceID)
		}

		fmt.Fprintf(os.Stderr, "Found: %s (%s)\n", found.Name, found.ID)
		fmt.Fprintf(os.Stderr, "New local key: %s\n", found.LocalKey)

		currentKey := appConfig.Sensor.LocalKey
		if currentKey == found.LocalKey {
			fmt.Fprintln(os.Stderr, "Key unchanged — no update needed.")
			return nil
		}

		fmt.Fprintf(os.Stderr, "Key changed: %s -> %s\n", currentKey, found.LocalKey)

		// Update config file
		if err := updateConfigKey(found.LocalKey, found.IP); err != nil {
			return fmt.Errorf("updating config: %w", err)
		}

		fmt.Fprintln(os.Stderr, "Config updated. Try `aquadirector sensor status` now.")
		return nil
	},
}

func updateConfigKey(newKey, newIP string) error {
	cfgPath := viper.ConfigFileUsed()
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cfgPath = filepath.Join(home, ".config", "aquadirector", "aquadirector.yaml")
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	sensorCfg, ok := raw["sensor"].(map[string]any)
	if !ok {
		sensorCfg = make(map[string]any)
		raw["sensor"] = sensorCfg
	}

	sensorCfg["local_key"] = newKey
	if newIP != "" {
		// Tuya cloud sometimes returns the IP; update if available
		currentIP, _ := sensorCfg["ip"].(string)
		if currentIP != "" && newIP != currentIP && !strings.HasPrefix(newIP, "0.") {
			sensorCfg["ip"] = newIP
			fmt.Fprintf(os.Stderr, "IP updated: %s -> %s\n", currentIP, newIP)
		}
	}

	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, out, 0644)
}

func init() {
	sensorRekeyCmd.Flags().StringVar(&rekeyClientID, "client-id", "", "Tuya Cloud Access ID / Client ID")
	sensorRekeyCmd.Flags().StringVar(&rekeyClientSecret, "client-secret", "", "Tuya Cloud Access Secret / Client Secret")
	sensorCmd.AddCommand(sensorRekeyCmd)
}
