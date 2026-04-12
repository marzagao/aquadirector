package cmd

import (
	"fmt"
	"os"

	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/internal/sensor"
	"github.com/spf13/cobra"
)

var sensorProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Discover Kactoily sensor protocol",
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := getSensorIP()
		if ip == "" {
			return fmt.Errorf("no sensor IP configured; use --ip or add to config")
		}

		fmt.Fprintf(os.Stderr, "Probing %s...\n", ip)

		result, err := sensor.Probe(cmd.Context(), ip, nil)
		if err != nil {
			return fmt.Errorf("probe failed: %w", err)
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, result, nil)
		}

		fmt.Printf("IP: %s\n", result.IP)
		fmt.Printf("Protocol: %s\n", result.Protocol)
		fmt.Printf("\nOpen ports:\n")
		for _, p := range result.OpenPorts {
			fmt.Printf("  %d (%s)\n", p.Port, p.Service)
		}
		fmt.Printf("\n%s\n", result.Details)

		if result.Protocol != "" {
			fmt.Fprintf(os.Stderr, "\nUpdate your config with:\n  sensor:\n    protocol: %q\n", result.Protocol)
		}

		return nil
	},
}

func init() {
	sensorCmd.AddCommand(sensorProbeCmd)
}
