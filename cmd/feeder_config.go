package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/pkg/eheim"
)

var feederConfigSet []string

var feederConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "View or update feeder configuration",
	Long: `View or update autofeeder settings.

Settable keys:
  overfeeding=on|off     Overfeeding protection (max 3 feedings/day)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getFeederClient(cmd.Context())
		if err != nil {
			return err
		}

		if len(feederConfigSet) > 0 {
			update := eheim.FeederConfigUpdate{}
			for _, kv := range feederConfigSet {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid key=value: %s", kv)
				}
				key, val := parts[0], parts[1]
				switch key {
				case "overfeeding":
					b, err := parseBoolFlag(val)
					if err != nil {
						return fmt.Errorf("invalid value for %s: %w", key, err)
					}
					update.Overfeeding = &b
				default:
					return fmt.Errorf("unknown config key: %s (valid: overfeeding)", key)
				}
			}

			if err := client.SetConfig(cmd.Context(), update); err != nil {
				return fmt.Errorf("updating configuration: %w", err)
			}
			fmt.Println("Configuration updated.")
			return nil
		}

		fd, err := client.Status(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching feeder status: %w", err)
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, fd, nil)
		}

		rows := []output.TableRow{
			{Label: "Overfeeding Protection", Value: onOff(fd.Overfeeding)},
			{Label: "Feeding Break", Value: onOff(fd.FeedingBreak)},
			{Label: "Is Break Day", Value: onOff(fd.IsBreakDay)},
		}

		if fd.SyncFilter != "" {
			rows = append(rows, output.TableRow{
				Label: "Filter Sync",
				Value: fd.SyncFilter,
			})
		}

		return output.Print(os.Stdout, format, fd, rows)
	},
}

func parseBoolFlag(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "on", "true", "1", "yes":
		return true, nil
	case "off", "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("expected on|off, got %q", s)
	}
}

func init() {
	feederConfigCmd.Flags().StringArrayVar(&feederConfigSet, "set", nil, "set config key=value (repeatable)")
	feederCmd.AddCommand(feederConfigCmd)
}
