package alerts

import (
	"context"
	"fmt"
	"strings"

	"github.com/marzagao/aquadirector/pkg/redsea"
)

// DeviceFetcher fetches metrics from configured aquarium devices.
type DeviceFetcher struct {
	ATOClient   *redsea.ATOClient
	LEDClient   *redsea.LEDClient
	SensorFetch func(ctx context.Context, metric string) (any, error)
	FeederFetch func(ctx context.Context, metric string) (any, error)
}

func (f *DeviceFetcher) Fetch(ctx context.Context, source, metric string) (any, error) {
	switch source {
	case "ato":
		return f.fetchATO(ctx, metric)
	case "led":
		return f.fetchLED(ctx, metric)
	case "sensor":
		if f.SensorFetch != nil {
			return f.SensorFetch(ctx, metric)
		}
		return nil, fmt.Errorf("sensor not configured")
	case "feeder":
		if f.FeederFetch != nil {
			return f.FeederFetch(ctx, metric)
		}
		return nil, fmt.Errorf("feeder not configured")
	default:
		return nil, fmt.Errorf("unknown source: %s", source)
	}
}

func (f *DeviceFetcher) fetchATO(ctx context.Context, metric string) (any, error) {
	if f.ATOClient == nil {
		return nil, fmt.Errorf("ATO device not configured")
	}

	dash, err := f.ATOClient.Dashboard(ctx)
	if err != nil {
		return nil, err
	}

	return extractField(dash, metric)
}

func (f *DeviceFetcher) fetchLED(ctx context.Context, metric string) (any, error) {
	if f.LEDClient == nil {
		return nil, fmt.Errorf("LED device not configured")
	}

	state, err := f.LEDClient.ManualState(ctx)
	if err != nil {
		return nil, err
	}

	return extractField(state, metric)
}

func extractField(obj any, path string) (any, error) {
	parts := strings.Split(path, ".")

	switch v := obj.(type) {
	case *redsea.ATODashboard:
		return extractATOField(v, parts)
	case *redsea.LEDManualState:
		return extractLEDField(v, parts)
	default:
		return nil, fmt.Errorf("unsupported object type for field extraction")
	}
}

func extractATOField(d *redsea.ATODashboard, parts []string) (any, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty metric path")
	}

	switch parts[0] {
	case "volume_left":
		return d.VolumeLeft, nil
	case "mode":
		return d.Mode, nil
	case "is_pump_on":
		return d.IsPumpOn, nil
	case "pump_state":
		return d.PumpState, nil
	case "pump_speed":
		return d.PumpSpeed, nil
	case "total_fills":
		return d.TotalFills, nil
	case "flow_rate":
		return d.FlowRate, nil
	case "temperature":
		if !d.HasTemperature() {
			return nil, nil
		}
		return d.Temperature(), nil
	case "ato_sensor":
		return extractATOSensorField(d, parts)
	case "leak_sensor":
		return extractLeakSensorField(d, parts)
	default:
		return nil, fmt.Errorf("unknown ATO metric: %s", parts[0])
	}
}

func extractATOSensorField(d *redsea.ATODashboard, parts []string) (any, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("ato_sensor requires a sub-field")
	}
	switch parts[1] {
	case "connected":
		return d.ATOSensor.Connected, nil
	case "current_level":
		return d.ATOSensor.CurrentLevel, nil
	case "temperature_probe_status":
		return d.ATOSensor.TemperatureProbeStatus, nil
	default:
		return nil, fmt.Errorf("unknown ato_sensor field: %s", parts[1])
	}
}

func extractLeakSensorField(d *redsea.ATODashboard, parts []string) (any, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("leak_sensor requires a sub-field")
	}
	switch parts[1] {
	case "status":
		return d.LeakSensor.Status, nil
	case "enabled":
		return d.LeakSensor.Enabled, nil
	case "connected":
		return d.LeakSensor.Connected, nil
	default:
		return nil, fmt.Errorf("unknown leak_sensor field: %s", parts[1])
	}
}

func extractLEDField(s *redsea.LEDManualState, parts []string) (any, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty metric path")
	}

	switch parts[0] {
	case "white":
		return s.White, nil
	case "blue":
		return s.Blue, nil
	case "moon":
		return s.Moon, nil
	case "temperature":
		return s.Temperature, nil
	default:
		return nil, fmt.Errorf("unknown LED metric: %s", parts[0])
	}
}
