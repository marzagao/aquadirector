package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/pkg/eheim"
)

var feederWatchInterval time.Duration

var feederStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show autofeeder status and schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithWatch(cmd.Context(), feederWatchInterval, func() error {
			client, err := getFeederClient(cmd.Context())
			if err != nil {
				return err
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
				{Label: "Drum", Value: fmt.Sprintf("%s (level=%d%%)", fd.DrumState, fd.Level)},
				{Label: "Weight (scale)", Value: fmt.Sprintf("%.1fg", fd.Weight)},
				{Label: "Overfeeding Protection", Value: onOff(fd.Overfeeding)},
				{Label: "Feeding Break", Value: fmt.Sprintf("%s (today: %s)", onOff(fd.FeedingBreak), breakDayStatus(fd.IsBreakDay))},
			}
			if fd.SyncFilter != "" {
				rows = append(rows, output.TableRow{Label: "Filter Sync", Value: fd.SyncFilter})
			}

			output.Print(os.Stdout, format, fd, rows)

			// Schedule table
			fmt.Println("\n=== Feeding Schedule ===")
			printFeederSchedule(fd.Schedule)

			return nil
		})
	},
}

func printFeederSchedule(sched eheim.WeekSchedule) {
	for i := 0; i < 7; i++ {
		day := eheim.ShortDayNames[i]
		slots := sched[i].Slots
		if len(slots) == 0 {
			fmt.Printf("  %-3s  (no feeding)\n", day)
			continue
		}
		for j, slot := range slots {
			prefix := fmt.Sprintf("  %-3s", day)
			if j > 0 {
				prefix = "     "
			}
			fmt.Printf("%s  %s  %d turn(s)\n", prefix, slot.TimeString(), slot.Turns)
		}
	}
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func breakDayStatus(b bool) string {
	if b {
		return "active"
	}
	return "inactive"
}

func init() {
	addWatchFlag(feederStatusCmd, &feederWatchInterval)
	feederCmd.AddCommand(feederStatusCmd)
}
