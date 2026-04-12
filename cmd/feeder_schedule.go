package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/pkg/eheim"
)

var (
	scheduleDay      string
	scheduleClearAll bool
	feedingBreakFlag string
)

var feederScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "View feeding schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return output.Print(os.Stdout, format, fd.Schedule, nil)
		}

		printFeederSchedule(fd.Schedule)
		if fd.FeedingBreak {
			fmt.Println("\n  Feeding break: on (random fasting day)")
		}
		return nil
	},
}

var feederScheduleSetCmd = &cobra.Command{
	Use:   "set [slots...]",
	Short: "Set feeding schedule for a day",
	Long: `Set the feeding schedule for a specific day.

Each slot is specified as HH:MM/TURNS (e.g. "08:00/2" for 2 turns at 8 AM).
Up to 2 slots per day.

Examples:
  feeder schedule set --day mon "08:00/2" "20:00/1"
  feeder schedule set --day fri "12:00/3"
  feeder schedule set --day mon "08:00/2" --feeding-break on`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dayIdx, err := parseDayName(scheduleDay)
		if err != nil {
			return err
		}

		if len(args) > 2 {
			return fmt.Errorf("maximum 2 feeding slots per day")
		}

		slots, err := parseSlots(args)
		if err != nil {
			return err
		}

		client, err := getFeederClient(cmd.Context())
		if err != nil {
			return err
		}

		fd, err := client.Status(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching current schedule: %w", err)
		}

		sched := fd.Schedule
		sched[dayIdx] = eheim.DaySchedule{Slots: slots}

		feedingBreak := fd.FeedingBreak
		if feedingBreakFlag != "" {
			feedingBreak, err = parseBoolFlag(feedingBreakFlag)
			if err != nil {
				return fmt.Errorf("invalid --feeding-break value: %w", err)
			}
		}

		if err := client.SetSchedule(cmd.Context(), sched, feedingBreak); err != nil {
			return fmt.Errorf("updating schedule: %w", err)
		}

		fmt.Printf("Schedule updated for %s.\n", eheim.DayNames[dayIdx])
		return nil
	},
}

var feederScheduleClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear feeding schedule (makes fasting day)",
	Long: `Clear the feeding schedule for a day or the entire week.

Examples:
  feeder schedule clear --day tue     Clear Tuesday (fasting day)
  feeder schedule clear --all         Clear entire week`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getFeederClient(cmd.Context())
		if err != nil {
			return err
		}

		fd, err := client.Status(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching current schedule: %w", err)
		}

		sched := fd.Schedule

		if scheduleClearAll {
			for i := range sched {
				sched[i] = eheim.DaySchedule{}
			}
		} else {
			dayIdx, err := parseDayName(scheduleDay)
			if err != nil {
				return err
			}
			sched[dayIdx] = eheim.DaySchedule{}
		}

		if err := client.SetSchedule(cmd.Context(), sched, fd.FeedingBreak); err != nil {
			return fmt.Errorf("clearing schedule: %w", err)
		}

		if scheduleClearAll {
			fmt.Println("Entire weekly schedule cleared.")
		} else {
			dayIdx, _ := parseDayName(scheduleDay)
			fmt.Printf("Schedule cleared for %s.\n", eheim.DayNames[dayIdx])
		}
		return nil
	},
}

func parseDayName(s string) (int, error) {
	switch strings.ToLower(s) {
	case "mon", "monday":
		return 0, nil
	case "tue", "tuesday":
		return 1, nil
	case "wed", "wednesday":
		return 2, nil
	case "thu", "thursday":
		return 3, nil
	case "fri", "friday":
		return 4, nil
	case "sat", "saturday":
		return 5, nil
	case "sun", "sunday":
		return 6, nil
	default:
		return 0, fmt.Errorf("invalid day: %q (use mon/tue/wed/thu/fri/sat/sun)", s)
	}
}

func parseSlots(args []string) ([]eheim.FeedSlot, error) {
	var slots []eheim.FeedSlot
	for _, arg := range args {
		parts := strings.SplitN(arg, "/", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid slot format %q (expected HH:MM/TURNS)", arg)
		}

		timeParts := strings.SplitN(parts[0], ":", 2)
		if len(timeParts) != 2 {
			return nil, fmt.Errorf("invalid time format %q (expected HH:MM)", parts[0])
		}

		hour, err := strconv.Atoi(timeParts[0])
		if err != nil || hour < 0 || hour > 23 {
			return nil, fmt.Errorf("invalid hour %q", timeParts[0])
		}
		minute, err := strconv.Atoi(timeParts[1])
		if err != nil || minute < 0 || minute > 59 {
			return nil, fmt.Errorf("invalid minute %q", timeParts[1])
		}

		turns, err := strconv.Atoi(parts[1])
		if err != nil || turns < 1 {
			return nil, fmt.Errorf("invalid turns %q (must be >= 1)", parts[1])
		}

		slots = append(slots, eheim.FeedSlot{
			TimeMinutes: hour*60 + minute,
			Turns:       turns,
		})
	}
	return slots, nil
}

func init() {
	feederScheduleSetCmd.Flags().StringVar(&scheduleDay, "day", "", "day of week (mon/tue/wed/thu/fri/sat/sun)")
	feederScheduleSetCmd.MarkFlagRequired("day")
	feederScheduleSetCmd.Flags().StringVar(&feedingBreakFlag, "feeding-break", "", "enable random fasting day (on|off)")

	feederScheduleClearCmd.Flags().StringVar(&scheduleDay, "day", "", "day of week to clear")
	feederScheduleClearCmd.Flags().BoolVar(&scheduleClearAll, "all", false, "clear entire week")

	feederScheduleCmd.AddCommand(feederScheduleSetCmd)
	feederScheduleCmd.AddCommand(feederScheduleClearCmd)
	feederCmd.AddCommand(feederScheduleCmd)
}
