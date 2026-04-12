package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/pkg/redsea"
	"github.com/spf13/cobra"
)

var (
	ledSchedDay     int
	ledSchedSetFile string
)

var ledScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "View or set LED schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getLEDClient()
		if err != nil {
			return err
		}

		if ledSchedSetFile != "" {
			data, err := os.ReadFile(ledSchedSetFile)
			if err != nil {
				return fmt.Errorf("reading schedule file: %w", err)
			}
			var sched redsea.LEDSchedule
			if err := json.Unmarshal(data, &sched); err != nil {
				return fmt.Errorf("parsing schedule JSON: %w", err)
			}
			if err := client.SetSchedule(cmd.Context(), ledSchedDay, &sched); err != nil {
				return fmt.Errorf("setting schedule: %w", err)
			}
			fmt.Printf("Schedule for day %d updated.\n", ledSchedDay)
			return nil
		}

		if ledSchedDay < 1 || ledSchedDay > 7 {
			// Show all days
			for day := 1; day <= 7; day++ {
				sched, err := client.Schedule(cmd.Context(), day)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Day %d: error: %v\n", day, err)
					continue
				}
				fmt.Printf("--- Day %d ---\n", day)
				printSchedule(sched)
			}
			return nil
		}

		sched, err := client.Schedule(cmd.Context(), ledSchedDay)
		if err != nil {
			return fmt.Errorf("fetching schedule: %w", err)
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, sched, nil)
		}

		printSchedule(sched)
		return nil
	},
}

func printSchedule(s *redsea.LEDSchedule) {
	printChannel("White", s.White)
	printChannel("Blue", s.Blue)
	printChannel("Moon", s.Moon)
}

func printChannel(name string, ch redsea.LEDChannel) {
	fmt.Printf("  %s: rise=%d set=%d points=", name, ch.Rise, ch.Set)
	for i, p := range ch.Points {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf("{t:%d,i:%d}", p.T, p.I)
	}
	fmt.Println()
}

func init() {
	ledScheduleCmd.Flags().IntVar(&ledSchedDay, "day", 0, "day of week (1-7, 0=all)")
	ledScheduleCmd.Flags().StringVar(&ledSchedSetFile, "set", "", "JSON file to upload as schedule")
	ledCmd.AddCommand(ledScheduleCmd)
}
