package cmd

import (
	"fmt"
	"os"

	"github.com/marzagao/aquadirector/internal/output"
	"github.com/spf13/cobra"
)

var alertsConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current alert configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		format := output.ParseFormat(outputFmt)

		if format != output.Table {
			return output.Print(os.Stdout, format, appConfig.Alerts, nil)
		}

		fmt.Printf("Alerts enabled: %v\n\n", appConfig.Alerts.Enabled)

		headers := []string{"RULE", "SOURCE", "METRIC", "OPERATOR", "THRESHOLD", "SEVERITY"}
		var rows [][]string
		for _, r := range appConfig.Alerts.Rules {
			rows = append(rows, []string{
				r.Name,
				r.Source,
				r.Metric,
				r.Operator,
				fmt.Sprintf("%v", r.Threshold),
				r.Severity,
			})
		}

		fmt.Println("Rules:")
		if err := output.PrintList(os.Stdout, format, nil, headers, rows); err != nil {
			return err
		}

		fmt.Printf("\nNotification channels: %d\n", len(appConfig.Alerts.Notifications))
		for _, n := range appConfig.Alerts.Notifications {
			fmt.Printf("  - %s (min severity: %s)\n", n.Type, n.SeverityMin)
		}

		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsConfigCmd)
}
