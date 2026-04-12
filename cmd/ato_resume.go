package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var atoResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Clear empty state and resume ATO operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getATOClient()
		if err != nil {
			return err
		}

		if err := client.Resume(cmd.Context()); err != nil {
			return fmt.Errorf("resume failed: %w", err)
		}

		fmt.Println("ATO resumed.")
		return nil
	},
}

func init() {
	atoCmd.AddCommand(atoResumeCmd)
}
