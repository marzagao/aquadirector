package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// addWatchFlag adds --watch and --interval flags to a command.
func addWatchFlag(cmd *cobra.Command, interval *time.Duration) {
	cmd.Flags().DurationVar(interval, "watch", 0, "continuously refresh at this interval (e.g. 10s, 1m)")
}

// runWithWatch wraps a one-shot RunE function with watch mode.
// If watch interval is 0, runs once. Otherwise loops until interrupted.
func runWithWatch(ctx context.Context, interval time.Duration, fn func() error) error {
	if interval == 0 {
		return fn()
	}

	if interval < time.Second {
		interval = time.Second
	}

	fmt.Fprintf(os.Stderr, "Watching every %s (Ctrl+C to stop)\n\n", interval)

	for {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		now := time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stderr, "[%s] Refreshing...\n\n", now)

		if err := fn(); err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
		}
	}
}
