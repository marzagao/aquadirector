package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marzagao/aquadirector/internal/color"
	"github.com/marzagao/aquadirector/internal/config"
	"github.com/marzagao/aquadirector/internal/output"
	"github.com/marzagao/aquadirector/internal/sensor"
	"github.com/marzagao/aquadirector/pkg/redsea"
	"github.com/spf13/cobra"
)

const notificationDays = 7

type dashboardData struct {
	WaterQuality  *waterQualityData          `json:"water_quality" yaml:"water_quality"`
	Equipment     *equipmentData             `json:"equipment" yaml:"equipment"`
	Notifications []redsea.CloudNotification `json:"notifications,omitempty" yaml:"notifications,omitempty"`
}

type waterQualityData struct {
	PH          float64 `json:"ph" yaml:"ph"`
	Temperature float64 `json:"temperature_c" yaml:"temperature_c"`
	TDS         int     `json:"tds" yaml:"tds"`
	EC          int     `json:"ec" yaml:"ec"`
	ORP         int     `json:"orp" yaml:"orp"`
	Salinity    float64 `json:"salinity" yaml:"salinity"`
	SG          float64 `json:"sg" yaml:"sg"`
}

type equipmentData struct {
	ATOMode              string   `json:"ato_mode,omitempty"          yaml:"ato_mode,omitempty"`
	WaterLevel           string   `json:"water_level,omitempty"       yaml:"water_level,omitempty"`
	VolumeLeft           int      `json:"volume_left,omitempty"       yaml:"volume_left,omitempty"`
	LeakSensor           string   `json:"leak_sensor,omitempty"       yaml:"leak_sensor,omitempty"`
	ATOTemp              float64  `json:"ato_temp_c,omitempty"        yaml:"ato_temp_c,omitempty"`
	TodayFills           int      `json:"today_fills,omitempty"       yaml:"today_fills,omitempty"`
	TodayVolume          int      `json:"today_volume,omitempty"      yaml:"today_volume,omitempty"`
	LastFill             *int64   `json:"last_fill,omitempty"         yaml:"last_fill,omitempty"`
	PumpOn               bool     `json:"pump_on,omitempty"           yaml:"pump_on,omitempty"`
	LastPumpCause        string   `json:"last_pump_cause,omitempty"   yaml:"last_pump_cause,omitempty"`
	DaysTillEmpty        *float64 `json:"days_till_empty,omitempty"   yaml:"days_till_empty,omitempty"`
	DailyVolumeAvg       *float64 `json:"daily_volume_avg_ml,omitempty" yaml:"daily_volume_avg_ml,omitempty"`
	ATOTempMin           float64  `json:"ato_temp_7d_min_c,omitempty" yaml:"ato_temp_7d_min_c,omitempty"`
	ATOTempMax           float64  `json:"ato_temp_7d_max_c,omitempty" yaml:"ato_temp_7d_max_c,omitempty"`
	ATOTempAvg           float64  `json:"ato_temp_7d_avg_c,omitempty" yaml:"ato_temp_7d_avg_c,omitempty"`
	LEDMode              string   `json:"led_mode,omitempty"          yaml:"led_mode,omitempty"`
	White                int      `json:"white,omitempty"             yaml:"white,omitempty"`
	Blue                 int      `json:"blue,omitempty"              yaml:"blue,omitempty"`
	Moon                 int      `json:"moon,omitempty"              yaml:"moon,omitempty"`
	LEDTemp              float64  `json:"led_temp_c,omitempty"        yaml:"led_temp_c,omitempty"`
	AcclimationActive    bool     `json:"acclimation_active,omitempty"    yaml:"acclimation_active,omitempty"`
	AcclimationDaysLeft  int      `json:"acclimation_days_left,omitempty" yaml:"acclimation_days_left,omitempty"`
	AcclimationIntensity int      `json:"acclimation_intensity,omitempty" yaml:"acclimation_intensity,omitempty"`
	FeederWeight         float64  `json:"feeder_weight,omitempty"     yaml:"feeder_weight,omitempty"`
	FeederDrum           string   `json:"feeder_drum,omitempty"       yaml:"feeder_drum,omitempty"`
	FeederLevel          int      `json:"feeder_level,omitempty"      yaml:"feeder_level,omitempty"`
	Battery              int      `json:"battery,omitempty"           yaml:"battery,omitempty"`
}

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash"},
	Short:   "Show consolidated aquarium dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		data := &dashboardData{
			WaterQuality: &waterQualityData{},
			Equipment:    &equipmentData{},
		}

		// Kactoily sensor (water quality)
		if appConfig.Sensor.DeviceID != "" {
			sc := sensor.NewClient(appConfig.Sensor.IP, appConfig.Sensor.DeviceID, appConfig.Sensor.LocalKey, appConfig.Sensor.Version, appConfig.Sensor.Calibration)
			wq, err := sc.ReadWaterQuality(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Sensor: %v\n", err)
			} else {
				data.WaterQuality.PH = wq.PH
				data.WaterQuality.Temperature = wq.Temperature
				data.WaterQuality.TDS = wq.TDS
				data.WaterQuality.EC = wq.EC
				data.WaterQuality.ORP = wq.ORP
				data.WaterQuality.Salinity = wq.Salinity
				data.WaterQuality.SG = wq.SG
				data.Equipment.Battery = wq.Battery
			}
		}

		// ATO
		if dev := appConfig.DeviceByType("RSATO+"); dev != nil {
			client := redsea.NewATOClient(dev.IP,
				redsea.WithTimeout(appConfig.Network.DefaultTimeout),
				redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay))
			dash, err := client.Dashboard(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "ATO: %v\n", err)
			} else {
				data.Equipment.ATOMode = dash.Mode
				data.Equipment.WaterLevel = dash.ATOSensor.CurrentLevel
				data.Equipment.VolumeLeft = dash.VolumeLeft
				data.Equipment.TodayFills = dash.TodayFills
				data.Equipment.TodayVolume = dash.TodayVolumeUsage
				data.Equipment.LastFill = dash.LastFillDate
				data.Equipment.LeakSensor = dash.LeakSensor.Status
				data.Equipment.PumpOn = dash.IsPumpOn
				data.Equipment.LastPumpCause = dash.LastPumpOnCause
				data.Equipment.DaysTillEmpty = dash.DaysTillEmpty
				data.Equipment.DailyVolumeAvg = dash.DailyVolumeAvg
				if dash.HasTemperature() {
					data.Equipment.ATOTemp = dash.Temperature()
				}
			}
		}

		// LED
		for _, model := range []string{"RSLED60", "RSLED50", "RSLED90", "RSLED115", "RSLED160", "RSLED170"} {
			if dev := appConfig.DeviceByType(model); dev != nil {
				client := redsea.NewLEDClient(dev.IP,
					redsea.WithTimeout(appConfig.Network.DefaultTimeout),
					redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay))
				manual, err := client.ManualState(cmd.Context())
				if err != nil {
					fmt.Fprintf(os.Stderr, "LED: %v\n", err)
				} else {
					data.Equipment.White = manual.White
					data.Equipment.Blue = manual.Blue
					data.Equipment.Moon = manual.Moon
					data.Equipment.LEDTemp = manual.Temperature
				}
				mode, err := client.Mode(cmd.Context())
				if err == nil {
					data.Equipment.LEDMode = mode
				}
				accl, err := client.Acclimation(cmd.Context())
				if err == nil && accl.Enabled {
					data.Equipment.AcclimationActive = true
					data.Equipment.AcclimationDaysLeft = accl.RemainingDays
					data.Equipment.AcclimationIntensity = accl.CurrentIntensityFactor
				}
				break
			}
		}

		// Red Sea cloud
		if cc := newCloudClient(); cc != nil {
			notifs, err := cc.GetNotifications(cmd.Context(), notificationDays, 50)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cloud notifications: %v\n", err)
			} else {
				data.Notifications = notifs
			}

			if dev := appConfig.DeviceByType("RSATO+"); dev != nil {
				hwid := resolveATOHWID(cmd.Context(), dev)
				if hwid != "" {
					tempLog, err := cc.GetATOTemperatureLog(cmd.Context(), hwid, "P7D")
					if err != nil {
						fmt.Fprintf(os.Stderr, "ATO temp log: %v\n", err)
					} else if len(tempLog.Entries) > 0 {
						var mn, mx, sum float64
						count := 0
						for _, entry := range tempLog.Entries {
							for _, v := range entry.Avg {
								if v == 0 {
									continue // skip missing readings
								}
								if count == 0 {
									mn, mx = v, v
								}
								if v < mn {
									mn = v
								}
								if v > mx {
									mx = v
								}
								sum += v
								count++
							}
						}
						if count > 0 {
							data.Equipment.ATOTempMin = mn
							data.Equipment.ATOTempMax = mx
							data.Equipment.ATOTempAvg = sum / float64(count)
						}
					}
				}
			}
		}

		// Eheim feeder
		if appConfig.Feeder.Host != "" {
			client, err := getFeederClient(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Feeder: %v\n", err)
			} else {
				fd, err := client.Status(cmd.Context())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Feeder: %v\n", err)
				} else {
					data.Equipment.FeederWeight = fd.Weight
					data.Equipment.FeederDrum = fd.DrumState.String()
					data.Equipment.FeederLevel = fd.Level
				}
			}
		}

		format := output.ParseFormat(outputFmt)
		if format != output.Table {
			return output.Print(os.Stdout, format, data, nil)
		}

		// Water Quality section
		fmt.Println(color.Bold("=== Water Quality ==="))
		wqRows := []output.TableRow{
			{Label: color.Dim("pH"), Value: fmt.Sprintf("%.2f (%s)", data.WaterQuality.PH, color.StatusLabel(phStatus(data.WaterQuality.PH)))},
			{Label: color.Dim("Temperature"), Value: formatDashTemp(data.WaterQuality.Temperature) + " (sensor)"},
			{Label: color.Dim("ORP"), Value: fmt.Sprintf("%d mV (%s)", data.WaterQuality.ORP, color.StatusLabel(orpStatus(data.WaterQuality.ORP)))},
			{Label: color.Dim("Salinity"), Value: fmt.Sprintf("%.2f%%", data.WaterQuality.Salinity)},
			{Label: color.Dim("SG"), Value: fmt.Sprintf("%.3f", data.WaterQuality.SG)},
		}
		output.Print(os.Stdout, output.Table, nil, wqRows)

		// ATO section
		var atoRows []output.TableRow

		if data.Equipment.ATOMode != "" {
			atoInfo := fmt.Sprintf("%s (level=%s, %.1f gal)", data.Equipment.ATOMode, data.Equipment.WaterLevel, float64(data.Equipment.VolumeLeft)/3785.41)
			if data.Equipment.ATOTemp > 0 {
				atoInfo += fmt.Sprintf(" temp=%.1fC (probe)", data.Equipment.ATOTemp)
			}
			atoRows = append(atoRows, output.TableRow{Label: color.Dim("Status"), Value: atoInfo})
		}

		if data.Equipment.LastPumpCause != "" || data.Equipment.PumpOn {
			pumpVal := "off"
			if data.Equipment.PumpOn {
				pumpVal = color.Blue("running")
			}
			if data.Equipment.LastPumpCause != "" {
				pumpVal += fmt.Sprintf(", last trigger: %s", formatPumpCause(data.Equipment.LastPumpCause))
			}
			atoRows = append(atoRows, output.TableRow{Label: color.Dim("Pump"), Value: pumpVal})
		}

		if data.Equipment.TodayFills > 0 || data.Equipment.LastFill != nil {
			atoRows = append(atoRows, output.TableRow{
				Label: color.Dim("Today Fills"),
				Value: fmt.Sprintf("%d (%.2f gal)", data.Equipment.TodayFills, float64(data.Equipment.TodayVolume)/3785.41),
			})
			lastFill := "never"
			if data.Equipment.LastFill != nil {
				lastFill = time.Unix(*data.Equipment.LastFill, 0).Format("2006-01-02 15:04")
			}
			atoRows = append(atoRows, output.TableRow{Label: color.Dim("Last Fill"), Value: lastFill})
		}

		if data.Equipment.DaysTillEmpty != nil {
			avgStr := ""
			if data.Equipment.DailyVolumeAvg != nil {
				avgStr = fmt.Sprintf(" (avg %.0fml/day)", *data.Equipment.DailyVolumeAvg)
			}
			days := *data.Equipment.DaysTillEmpty
			daysStr := fmt.Sprintf("~%.0f days till empty", days)
			switch {
			case days < 3:
				daysStr = color.Red(daysStr)
			case days < 7:
				daysStr = color.Yellow(daysStr)
			default:
				daysStr = color.Green(daysStr)
			}
			atoRows = append(atoRows, output.TableRow{
				Label: color.Dim("Reservoir"),
				Value: daysStr + avgStr,
			})
		}

		if data.Equipment.ATOTempAvg > 0 {
			atoRows = append(atoRows, output.TableRow{
				Label: color.Dim("Temp 7d (cloud)"),
				Value: fmt.Sprintf("min=%.1fC avg=%.1fC max=%.1fC",
					data.Equipment.ATOTempMin, data.Equipment.ATOTempAvg, data.Equipment.ATOTempMax),
			})
		}

		if data.Equipment.LeakSensor != "" {
			leakVal := data.Equipment.LeakSensor
			if leakVal == "dry" {
				leakVal = color.Green(leakVal)
			} else {
				leakVal = color.Red(leakVal)
			}
			atoRows = append(atoRows, output.TableRow{Label: color.Dim("Leak Sensor"), Value: leakVal})
		}

		if len(atoRows) > 0 {
			fmt.Println(color.Bold("\n=== ATO ==="))
			if err := output.Print(os.Stdout, output.Table, nil, atoRows); err != nil {
				return err
			}
		}

		// Lighting section
		if data.Equipment.LEDMode != "" || data.Equipment.AcclimationActive {
			fmt.Println(color.Bold("\n=== Lighting ==="))
			var lightRows []output.TableRow

			if data.Equipment.LEDMode != "" {
				lightRows = append(lightRows, output.TableRow{
					Label: color.Dim("Mode"),
					Value: fmt.Sprintf("%s (moon=%d, white=%d, blue=%d)", data.Equipment.LEDMode, data.Equipment.Moon, data.Equipment.White, data.Equipment.Blue),
				})
			}

			if data.Equipment.AcclimationActive {
				lightRows = append(lightRows, output.TableRow{
					Label: color.Dim("Acclimation"),
					Value: color.Blue(fmt.Sprintf("%d days remaining (intensity %d%%)",
						data.Equipment.AcclimationDaysLeft, data.Equipment.AcclimationIntensity)),
				})
			}

			if err := output.Print(os.Stdout, output.Table, nil, lightRows); err != nil {
				return err
			}
		}

		// Feeding section
		if data.Equipment.FeederDrum != "" {
			fmt.Println(color.Bold("\n=== Feeding ==="))
			weightStr := fmt.Sprintf("%.1fg", data.Equipment.FeederWeight)
			if data.Equipment.FeederWeight < 0 {
				weightStr = "~0g"
			}
			feederRows := []output.TableRow{
				{Label: color.Dim("Food"), Value: fmt.Sprintf("%s drum=%s (level=%d)", weightStr, colorizeDrumState(data.Equipment.FeederDrum), data.Equipment.FeederLevel)},
			}
			if err := output.Print(os.Stdout, output.Table, nil, feederRows); err != nil {
				return err
			}
		}

		// Notifications section (only when cloud is configured)
		if appConfig.Cloud.Username == "" || appConfig.Cloud.ClientCredentials == "" {
			return nil
		}
		fmt.Println(color.Bold(fmt.Sprintf("\n=== Notifications (last %d days) ===", notificationDays)))
		if len(data.Notifications) == 0 {
			fmt.Println("(none)")
		} else {
			var nRows []output.TableRow
			for _, n := range data.Notifications {
				ts := n.TimeSent.Local().Format("2006-01-02 15:04")
				unread := ""
				if !n.Read {
					unread = color.Yellow(" *")
				}
				nRows = append(nRows, output.TableRow{
					Label: color.Dim(ts) + unread,
					Value: formatNotificationText(n),
				})
			}
			output.Print(os.Stdout, output.Table, nil, nRows)
		}

		return nil
	},
}

func phStatus(ph float64) string {
	switch {
	case ph < 7.8:
		return "critical"
	case ph < 8.1:
		return "low"
	case ph <= 8.3:
		return "ok"
	default:
		return "high"
	}
}

func orpStatus(orp int) string {
	switch {
	case orp < 100:
		return "critical"
	case orp < 200:
		return "low"
	case orp <= 450:
		return "ok"
	default:
		return "high"
	}
}

func colorizeDrumState(state string) string {
	switch state {
	case "GREEN":
		return color.Green(state)
	case "ORANGE":
		return color.Yellow(state)
	case "RED":
		return color.Red(state)
	case "MEASURING":
		return color.Blue(state)
	default:
		return state
	}
}

func formatNotificationText(n redsea.CloudNotification) string {
	deviceLabels := map[string]string{
		"reef-ato":  "ATO",
		"reef-mat":  "ReefMat",
		"reef-run":  "ReefRun",
		"reef-dose": "ReefDose",
		"reef-wave": "ReefWave",
		"reef-led":  "LED",
	}
	text := n.Text
	if i := strings.Index(text, ": "); i >= 0 {
		text = text[i+2:]
	}
	if label, ok := deviceLabels[n.DeviceType]; ok {
		return label + ": " + text
	}
	return text
}

func formatPumpCause(cause string) string {
	labels := map[string]string{
		"ec_sensor_s1": "EC sensor",
		"ec_sensor_s2": "EC sensor 2",
		"schedule":     "schedule",
		"manual":       "manual",
		"timer":        "timer",
	}
	if label, ok := labels[cause]; ok {
		return label
	}
	if cause == "" {
		return "unknown"
	}
	return cause
}

func resolveATOHWID(ctx context.Context, dev *config.DeviceConfig) string {
	if dev.HWID != "" {
		return dev.HWID
	}
	client := redsea.NewATOClient(dev.IP,
		redsea.WithTimeout(appConfig.Network.DefaultTimeout),
		redsea.WithRetries(1, appConfig.Network.RetryDelay))
	info, err := client.DeviceInfo(ctx)
	if err != nil || info.HWID == "" {
		return ""
	}
	return info.HWID
}

func formatDashTemp(c float64) string {
	if c == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.1fC / %.1fF", c, c*9/5+32)
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
