package cmd

import (
	"context"
	"fmt"

	"github.com/marzagao/aquadirector/internal/alerts"
	"github.com/marzagao/aquadirector/internal/config"
	"github.com/marzagao/aquadirector/internal/sensor"
	"github.com/marzagao/aquadirector/pkg/eheim"
	"github.com/marzagao/aquadirector/pkg/redsea"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Alerting system",
}

func init() {
	rootCmd.AddCommand(alertsCmd)
}

func buildRules(cfgRules []config.AlertRule) []alerts.Rule {
	var rules []alerts.Rule
	for _, r := range cfgRules {
		rules = append(rules, alerts.Rule{
			Name:      r.Name,
			Source:    r.Source,
			Metric:    r.Metric,
			Operator:  r.Operator,
			Threshold: r.Threshold,
			Severity:  alerts.ParseSeverity(r.Severity),
			Message:   r.Message,
		})
	}
	return rules
}

func buildNotifiers(cfgNotifs []config.NotificationConfig) []alerts.Notifier {
	var notifiers []alerts.Notifier
	for _, n := range cfgNotifs {
		sev := alerts.ParseSeverity(n.SeverityMin)
		switch n.Type {
		case "stdout":
			notifiers = append(notifiers, alerts.NewStdoutNotifier(sev))
		case "webhook":
			notifiers = append(notifiers, alerts.NewWebhookNotifier(sev, n.URL, n.Method, n.Headers, n.BodyTemplate))
		case "command":
			notifiers = append(notifiers, alerts.NewCommandNotifier(sev, n.Command, n.Args))
		}
	}
	if len(notifiers) == 0 {
		notifiers = append(notifiers, alerts.NewStdoutNotifier(alerts.SeverityInfo))
	}
	return notifiers
}

func buildFetcher() *alerts.DeviceFetcher {
	fetcher := &alerts.DeviceFetcher{}

	if dev := appConfig.DeviceByType("RSATO+"); dev != nil {
		fetcher.ATOClient = redsea.NewATOClient(dev.IP,
			redsea.WithTimeout(appConfig.Network.DefaultTimeout),
			redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay))
	}

	// Try all LED models
	for _, model := range []string{"RSLED60", "RSLED50", "RSLED90", "RSLED115", "RSLED160", "RSLED170"} {
		if dev := appConfig.DeviceByType(model); dev != nil {
			fetcher.LEDClient = redsea.NewLEDClient(dev.IP,
				redsea.WithTimeout(appConfig.Network.DefaultTimeout),
				redsea.WithRetries(appConfig.Network.RetryMax, appConfig.Network.RetryDelay))
			break
		}
	}

	// Wire sensor into fetcher
	if appConfig.Sensor.DeviceID != "" && appConfig.Sensor.LocalKey != "" {
		sc := sensor.NewClient(
			appConfig.Sensor.IP,
			appConfig.Sensor.DeviceID,
			appConfig.Sensor.LocalKey,
			appConfig.Sensor.Version,
			appConfig.Sensor.Calibration,
		)
		fetcher.SensorFetch = func(ctx context.Context, metric string) (any, error) {
			wq, err := sc.ReadWaterQuality(ctx)
			if err != nil {
				return nil, err
			}
			return extractSensorField(wq, metric)
		}
	}

	// Wire feeder into fetcher
	if appConfig.Feeder.Host != "" {
		fetcher.FeederFetch = func(ctx context.Context, metric string) (any, error) {
			client, err := getFeederClient(ctx)
			if err != nil {
				return nil, err
			}
			fd, err := client.Status(ctx)
			if err != nil {
				return nil, err
			}
			return extractFeederField(fd, metric)
		}
	}

	return fetcher
}

func extractFeederField(fd *eheim.FeederData, metric string) (any, error) {
	switch metric {
	case "weight":
		return fd.Weight, nil
	case "drum_state":
		return int(fd.DrumState), nil
	case "level":
		return fd.Level, nil
	case "is_spinning":
		return fd.IsSpinning, nil
	case "overfeeding":
		return fd.Overfeeding, nil
	case "feeding_break":
		return fd.FeedingBreak, nil
	case "is_break_day":
		return fd.IsBreakDay, nil
	default:
		return nil, fmt.Errorf("unknown feeder metric: %s", metric)
	}
}

func extractSensorField(wq *sensor.WaterQuality, metric string) (any, error) {
	switch metric {
	case "ph":
		return wq.PH, nil
	case "temperature":
		return wq.Temperature, nil
	case "tds":
		return wq.TDS, nil
	case "ec":
		return wq.EC, nil
	case "orp":
		return wq.ORP, nil
	case "salinity":
		return wq.Salinity, nil
	case "sg":
		return wq.SG, nil
	case "battery":
		return wq.Battery, nil
	default:
		return nil, fmt.Errorf("unknown sensor metric: %s", metric)
	}
}
