package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/pkg/eheim"
)

var (
	feederHost string
	feederMAC  string
)

var feederCmd = &cobra.Command{
	Use:   "feeder",
	Short: "Eheim autofeeder+ commands",
}

func init() {
	feederCmd.PersistentFlags().StringVar(&feederHost, "host", "", "Eheim hub host (default from config)")
	feederCmd.PersistentFlags().StringVar(&feederMAC, "mac", "", "feeder MAC address (default from config)")
	rootCmd.AddCommand(feederCmd)
}

func feederOpts() []eheim.Option {
	var opts []eheim.Option
	if appConfig.Feeder.Username != "" || appConfig.Feeder.Password != "" {
		opts = append(opts, eheim.WithCredentials(appConfig.Feeder.Username, appConfig.Feeder.Password))
	}
	return opts
}

func getFeederClient(ctx context.Context) (*eheim.FeederClient, error) {
	host := feederHost
	if host == "" {
		host = appConfig.Feeder.Host
	}
	if host == "" {
		host = "eheimdigital.local"
	}

	mac := feederMAC
	if mac == "" {
		mac = appConfig.Feeder.MAC
	}

	if mac == "" {
		hub := eheim.NewHubClient(host, feederOpts()...)
		found, err := hub.FindFeeder(ctx)
		if err != nil {
			return nil, fmt.Errorf("no feeder MAC configured and auto-discovery failed: %w", err)
		}
		mac = found
	}

	return eheim.NewFeederClient(host, mac, feederOpts()...), nil
}
