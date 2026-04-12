package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/marzagao/aquadirector/internal/output"
	"github.com/spf13/cobra"
)

var atoConfigSet []string

var atoConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "View or update ATO configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getATOClient()
		if err != nil {
			return err
		}

		if len(atoConfigSet) > 0 {
			update := make(map[string]any)
			for _, kv := range atoConfigSet {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid key=value: %s", kv)
				}
				update[parts[0]] = parseValue(parts[1])
			}
			if err := client.SetConfiguration(cmd.Context(), update); err != nil {
				return fmt.Errorf("updating configuration: %w", err)
			}
			fmt.Println("Configuration updated.")
			return nil
		}

		cfg, err := client.Configuration(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching configuration: %w", err)
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, cfg, nil)
		}

		rows := []output.TableRow{
			{Label: "Auto Fill", Value: fmt.Sprintf("%v", cfg.AutoFill)},
			{Label: "Auto Delay", Value: fmt.Sprintf("%ds", cfg.AutoDelay)},
			{Label: "Buzzer", Value: fmt.Sprintf("enabled=%v freq=%d duty=%d%%", cfg.Buzzer.Enabled, cfg.Buzzer.Frequency, cfg.Buzzer.DutyCycle)},
			{Label: "Temp Range", Value: fmt.Sprintf("%.1f-%.1fC (desired: %.1f-%.1fC)", cfg.Temperature.AcceptableRangeLow, cfg.Temperature.AcceptableRangeHigh, cfg.Temperature.DesiredRangeLow, cfg.Temperature.DesiredRangeHigh)},
			{Label: "Leak Detection", Value: fmt.Sprintf("enabled=%v sensor=%v emergency=%v", cfg.Leak.Enabled, cfg.Leak.SensorEnabled, cfg.Leak.EmergencyShutdown)},
		}

		return output.Print(os.Stdout, format, cfg, rows)
	},
}

func parseValue(s string) any {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	return s
}

func init() {
	atoConfigCmd.Flags().StringArrayVar(&atoConfigSet, "set", nil, "set config key=value (repeatable)")
	atoCmd.AddCommand(atoConfigCmd)
}
