package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/marzagao/aquadirector/internal/config"
	"github.com/marzagao/aquadirector/internal/discovery"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/pkg/eheim"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	discoverSubnet    string
	discoverThreads   int
	discoverSave      bool
	discoverEheimHost string
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Scan network for Red Sea and Eheim devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		subnet := discoverSubnet
		if subnet == "" {
			subnet = appConfig.Network.Subnet
		}
		threads := discoverThreads
		if threads == 0 {
			threads = appConfig.Network.ScanThreads
		}

		eheimHost := discoverEheimHost
		if eheimHost == "" {
			eheimHost = appConfig.Feeder.Host
		}

		fmt.Fprintf(os.Stderr, "Scanning %s...\n", subnet)

		// Run Red Sea and Eheim scans in parallel
		var (
			result      *discovery.ScanResult
			eheimResult []discovery.DiscoveredEheimDevice
			scanErr     error
			eheimErr    error
			wg          sync.WaitGroup
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			result, scanErr = discovery.Scan(cmd.Context(), subnet, threads)
		}()

		if eheimHost != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Fprintf(os.Stderr, "Scanning Eheim mesh at %s...\n", eheimHost)
				eheimResult, eheimErr = discovery.ScanEheim(cmd.Context(), eheimHost)
			}()
		}

		wg.Wait()

		if scanErr != nil {
			return fmt.Errorf("Red Sea scan failed: %w", scanErr)
		}

		format := output.ParseFormat(outputFmt)
		totalFound := len(result.Devices)

		// Red Sea results
		if len(result.Devices) > 0 {
			fmt.Fprintf(os.Stderr, "\nFound %d Red Sea device(s):\n\n", len(result.Devices))
			headers := []string{"IP", "MODEL", "NAME", "UUID", "FIRMWARE"}
			var rows [][]string
			for _, d := range result.Devices {
				rows = append(rows, []string{d.IP, d.HWModel, d.Name, d.UUID, d.Firmware})
			}
			if err := output.PrintList(os.Stdout, format, result.Devices, headers, rows); err != nil {
				return err
			}
		}

		// Eheim results
		if eheimErr != nil {
			fmt.Fprintf(os.Stderr, "\nEheim scan failed: %v\n", eheimErr)
		} else if len(eheimResult) > 0 {
			totalFound += len(eheimResult)
			fmt.Fprintf(os.Stderr, "\nFound %d Eheim device(s):\n\n", len(eheimResult))
			headers := []string{"HOST", "MAC", "NAME", "TYPE", "REVISION"}
			var rows [][]string
			for _, d := range eheimResult {
				rows = append(rows, []string{d.Host, d.MAC, d.Name, d.Type, d.Revision})
			}
			if err := output.PrintList(os.Stdout, format, eheimResult, headers, rows); err != nil {
				return err
			}
			result.EheimDevices = eheimResult
		}

		if totalFound == 0 {
			fmt.Fprintln(os.Stderr, "No devices found.")
			return nil
		}

		if discoverSave {
			if err := saveDevices(result.Devices); err != nil {
				return err
			}
			return saveEheimDevices(eheimResult, eheimHost)
		}

		return nil
	},
}

func saveDevices(devices []discovery.DiscoveredDevice) error {
	cfgPath := viper.ConfigFileUsed()
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("finding home directory: %w", err)
		}
		cfgPath = filepath.Join(home, ".config", "aquadirector", "aquadirector.yaml")
	}

	// Read existing config as raw YAML to preserve structure
	var raw map[string]any
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading config: %w", err)
		}
		raw = make(map[string]any)
	} else {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}
	}

	// Build device list from discovered devices
	var deviceList []config.DeviceConfig
	for _, d := range devices {
		deviceList = append(deviceList, config.DeviceConfig{
			Name: d.Name,
			IP:   d.IP,
			Type: d.HWModel,
			UUID: d.UUID,
		})
	}

	// Convert to yaml-compatible format
	var devicesYAML []map[string]string
	for _, d := range deviceList {
		devicesYAML = append(devicesYAML, map[string]string{
			"name": d.Name,
			"ip":   d.IP,
			"type": d.Type,
			"uuid": d.UUID,
		})
	}
	raw["devices"] = devicesYAML

	// Write back
	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(cfgPath, out, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nSaved %d device(s) to %s\n", len(devices), cfgPath)
	return nil
}

func saveEheimDevices(devices []discovery.DiscoveredEheimDevice, host string) error {
	if len(devices) == 0 {
		return nil
	}

	// Find the first feeder
	var feederMAC, feederIP string
	for _, d := range devices {
		if d.Version == eheim.DeviceTypeFeeder {
			feederMAC = d.MAC
			feederIP = d.IP
			break
		}
	}

	if feederMAC == "" {
		return nil
	}

	// Prefer IP over mDNS hostname to avoid slow .local DNS resolution
	if feederIP != "" {
		host = feederIP
	}

	cfgPath := viper.ConfigFileUsed()
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("finding home directory: %w", err)
		}
		cfgPath = filepath.Join(home, ".config", "aquadirector", "aquadirector.yaml")
	}

	var raw map[string]any
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading config: %w", err)
		}
		raw = make(map[string]any)
	} else {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}
	}

	raw["feeder"] = map[string]string{
		"host": host,
		"mac":  feederMAC,
	}

	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(cfgPath, out, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Saved Eheim feeder (MAC: %s) to %s\n", feederMAC, cfgPath)
	return nil
}

func init() {
	discoverCmd.Flags().StringVar(&discoverSubnet, "subnet", "", "subnet to scan (default from config)")
	discoverCmd.Flags().IntVar(&discoverThreads, "threads", 0, "number of concurrent probes (default from config)")
	discoverCmd.Flags().StringVar(&discoverEheimHost, "eheim-host", "", "Eheim hub host for mesh scan (default from config)")
	discoverCmd.Flags().BoolVar(&discoverSave, "save", false, "save discovered devices to config file")
	rootCmd.AddCommand(discoverCmd)
}
