package sensor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marzagao/aquadirector/pkg/tuya"
)

// Client communicates with the Kactoily water sensor via the Tuya local protocol.
type Client struct {
	IP          string
	DeviceID    string
	LocalKey    string
	Version     string
	Calibration map[string]float64
}

func NewClient(ip, deviceID, localKey, version string, calibration map[string]float64) *Client {
	if version == "" {
		version = "3.5"
	}
	return &Client{IP: ip, DeviceID: deviceID, LocalKey: localKey, Version: version, Calibration: calibration}
}

func (c *Client) ReadWaterQuality(ctx context.Context) (*WaterQuality, error) {
	if c.DeviceID == "" || c.LocalKey == "" {
		return nil, fmt.Errorf("sensor device_id and local_key are required in config")
	}
	return c.readTuya(ctx)
}

func (c *Client) readTuya(ctx context.Context) (*WaterQuality, error) {
	device := tuya.NewDevice(c.DeviceID, c.IP, 0, c.LocalKey)
	dps, err := device.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("tuya query failed: %w", err)
	}
	wq, err := parseDPS(dps)
	if err != nil {
		return nil, err
	}
	applyCalibration(wq, c.Calibration)
	return wq, nil
}

func applyCalibration(wq *WaterQuality, cal map[string]float64) {
	if len(cal) == 0 {
		return
	}
	wq.PH += cal["ph"]
	wq.Temperature += cal["temperature"]
	wq.TDS += int(cal["tds"])
	wq.EC += int(cal["ec"])
	wq.ORP += int(cal["orp"])
	wq.Salinity += cal["salinity"]
	wq.SG += cal["sg"]
}

func parseDPS(dps map[string]any) (*WaterQuality, error) {
	wq := &WaterQuality{}

	if v, ok := dps["1"]; ok {
		wq.TDS = toInt(v)
	}
	if v, ok := dps["2"]; ok {
		wq.Temperature = float64(toInt(v)) / 10.0
	}
	if v, ok := dps["7"]; ok {
		wq.Battery = toInt(v)
	}
	if v, ok := dps["10"]; ok {
		wq.PH = float64(toInt(v)) / 100.0
	}
	if v, ok := dps["11"]; ok {
		wq.EC = toInt(v)
	}
	if v, ok := dps["12"]; ok {
		wq.ORP = toInt(v)
	}
	if v, ok := dps["102"]; ok {
		wq.Salinity = float64(toInt(v)) / 100.0
	}
	if v, ok := dps["103"]; ok {
		wq.SG = float64(toInt(v)) / 1000.0
	}

	return wq, nil
}

func toInt(v any) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}
