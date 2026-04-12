package cmd

import (
	"fmt"
	"os"

	"github.com/marzagao/aquadirector/internal/alerts"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/spf13/cobra"
)

var alertsNotify bool

var alertsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Evaluate alert rules against current device state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !appConfig.Alerts.Enabled {
			fmt.Fprintln(os.Stderr, "Alerts are disabled in config.")
			return nil
		}

		rules := buildRules(appConfig.Alerts.Rules)
		notifiers := buildNotifiers(appConfig.Alerts.Notifications)
		fetcher := buildFetcher()

		engine := alerts.NewEngine(rules, fetcher, notifiers)
		results, err := engine.Check(cmd.Context())
		if err != nil {
			return fmt.Errorf("checking alerts: %w", err)
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, results, nil)
		}

		headers := []string{"RULE", "SEVERITY", "STATUS", "VALUE", "MESSAGE"}
		var rows [][]string
		triggered := 0
		for _, r := range results {
			status := "ok"
			if r.Triggered {
				status = "TRIGGERED"
				triggered++
			}
			valueStr := "N/A"
			if r.Value != nil {
				valueStr = fmt.Sprintf("%v", r.Value)
			}
			rows = append(rows, []string{
				r.Rule.Name,
				r.Rule.Severity.String(),
				status,
				valueStr,
				r.Message,
			})
		}

		if err := output.PrintList(os.Stdout, format, results, headers, rows); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "\n%d/%d rules triggered.\n", triggered, len(results))

		if alertsNotify && triggered > 0 {
			if err := engine.Notify(cmd.Context(), results); err != nil {
				return fmt.Errorf("sending notifications: %w", err)
			}
		}

		return nil
	},
}

func init() {
	alertsCheckCmd.Flags().BoolVar(&alertsNotify, "notify", false, "send notifications for triggered alerts")
	alertsCmd.AddCommand(alertsCheckCmd)
}
