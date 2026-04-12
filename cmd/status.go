package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/internal/sensor"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

var statusDevice string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all configured devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		format := output.ParseFormat(outputFmt)

		type deviceStatus struct {
			Name   string `json:"name" yaml:"name"`
			IP     string `json:"ip" yaml:"ip"`
			Type   string `json:"type" yaml:"type"`
			Status string `json:"status" yaml:"status"`
			Info   string `json:"info" yaml:"info"`
		}

		var statuses []deviceStatus

		// Red Sea devices
		for _, dev := range appConfig.Devices {
			if statusDevice != "" && dev.Name != statusDevice && dev.IP != statusDevice {
				continue
			}

			ds := deviceStatus{Name: dev.Name, IP: dev.IP, Type: dev.Type}

			client := redsea.New(dev.IP,
				redsea.WithTimeout(appConfig.Network.DefaultTimeout),
				redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay))

			info, err := client.DeviceInfo(cmd.Context())
			if err != nil {
				ds.Status = "offline"
				ds.Info = err.Error()
			} else {
				ds.Status = "online"
				ds.Info = fmt.Sprintf("model=%s fw=%s", info.HWModel, info.HWRevision)
			}

			statuses = append(statuses, ds)
		}

		// Kactoily sensor
		if appConfig.Sensor.DeviceID != "" {
			if statusDevice == "" || statusDevice == "sensor" || statusDevice == appConfig.Sensor.IP {
				ds := deviceStatus{
					Name: "Kactoily Sensor",
					IP:   appConfig.Sensor.IP,
					Type: "Tuya",
				}

				sc := sensor.NewClient(appConfig.Sensor.IP, appConfig.Sensor.DeviceID, appConfig.Sensor.LocalKey, appConfig.Sensor.Version, appConfig.Sensor.Calibration)
				wq, err := sc.ReadWaterQuality(cmd.Context())
				if err != nil {
					ds.Status = "offline"
					ds.Info = err.Error()
				} else {
					ds.Status = "online"
					ds.Info = fmt.Sprintf("pH=%.2f temp=%.1fC TDS=%d EC=%d", wq.PH, wq.Temperature, wq.TDS, wq.EC)
				}

				statuses = append(statuses, ds)
			}
		}

		// Eheim feeder
		if appConfig.Feeder.Host != "" {
			if statusDevice == "" || statusDevice == "feeder" || statusDevice == appConfig.Feeder.Host {
				ds := deviceStatus{
					Name: "Eheim Autofeeder+",
					IP:   appConfig.Feeder.Host,
					Type: "Eheim",
				}

				client, err := getFeederClient(cmd.Context())
				if err != nil {
					ds.Status = "error"
					ds.Info = err.Error()
				} else {
					fd, err := client.Status(cmd.Context())
					if err != nil {
						ds.Status = "offline"
						ds.Info = err.Error()
					} else {
						ds.Status = "online"
						ds.Info = fmt.Sprintf("weight=%.1fg drum=%s", fd.Weight, fd.DrumState)
					}
				}

				statuses = append(statuses, ds)
			}
		}

		if len(statuses) == 0 {
			fmt.Fprintln(os.Stderr, "No devices configured. Run 'aquadirector discover' or add sensor credentials to config.")
			return nil
		}

		if format != output.Table {
			return output.Print(os.Stdout, format, statuses, nil)
		}

		headers := []string{"NAME", "IP", "TYPE", "STATUS", "INFO"}
		var rows [][]string
		for _, s := range statuses {
			rows = append(rows, []string{s.Name, s.IP, s.Type, s.Status, s.Info})
		}

		return output.PrintList(os.Stdout, format, statuses, headers, rows)
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusDevice, "device", "", "filter by device name or IP")
	rootCmd.AddCommand(statusCmd)
}
