package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var feederFeedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Trigger a single manual feeding",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getFeederClient(cmd.Context())
		if err != nil {
			return err
		}

		if err := client.Feed(cmd.Context()); err != nil {
			return fmt.Errorf("triggering feed: %w", err)
		}

		fmt.Println("Manual feed triggered.")
		return nil
	},
}

func init() {
	feederCmd.AddCommand(feederFeedCmd)
}
