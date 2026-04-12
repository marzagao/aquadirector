// Package cmd implements the aquadirector CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marzagao/aquadirector/internal/color"
	"github.com/marzagao/aquadirector/internal/config"
	"github.com/marzagao/aquadirector/pkg/redsea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	outputFmt string
	colorMode string
	verbose   bool
	appConfig *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "aquadirector",
	Short: "Home aquarium automation CLI",
	Long:  "Monitor and control Red Sea and Kactoily aquarium devices from the command line.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(viper.GetViper())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		appConfig = cfg
		color.Init(colorMode)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if outputFmt == "" || outputFmt == "table" {
			fmt.Printf("\nRan at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/aquadirector/aquadirector.yaml)")
	rootCmd.PersistentFlags().StringVar(&outputFmt, "output", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&colorMode, "color", "auto", "colorize output: auto, always, never (honors NO_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug logging")
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aquadirector")
}

func newCloudClient() *redsea.CloudClient {
	if appConfig.Cloud.Username == "" || appConfig.Cloud.ClientCredentials == "" {
		return nil
	}
	tokenFile := filepath.Join(configDir(), "cloud_token.json")
	return redsea.NewCloudClient(
		appConfig.Cloud.Username,
		appConfig.Cloud.Password,
		appConfig.Cloud.ClientCredentials,
		tokenFile,
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		dir := configDir()
		if dir == "" {
			fmt.Fprintln(os.Stderr, "Error finding home directory")
			os.Exit(1)
		}
		viper.AddConfigPath(dir)
		viper.SetConfigName("aquadirector")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config:", err)
		}
	}
}
