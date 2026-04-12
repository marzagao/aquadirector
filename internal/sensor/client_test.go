package sensor

import (
	"testing"
)

func TestParseDPS(t *testing.T) {
	dps := map[string]any{
		"1":   float64(9150),
		"2":   float64(263),
		"7":   float64(100),
		"10":  float64(800),
		"11":  float64(18300),
		"12":  float64(165),
		"102": float64(104),
		"103": float64(1005),
	}

	wq, err := parseDPS(dps)
	if err != nil {
		t.Fatalf("parseDPS: %v", err)
	}

	if wq.TDS != 9150 {
		t.Errorf("TDS = %d, want 9150", wq.TDS)
	}
	if wq.Temperature != 26.3 {
		t.Errorf("Temperature = %f, want 26.3", wq.Temperature)
	}
	if wq.Battery != 100 {
		t.Errorf("Battery = %d, want 100", wq.Battery)
	}
	if wq.PH != 8.00 {
		t.Errorf("PH = %f, want 8.00", wq.PH)
	}
	if wq.EC != 18300 {
		t.Errorf("EC = %d, want 18300", wq.EC)
	}
	if wq.ORP != 165 {
		t.Errorf("ORP = %d, want 165", wq.ORP)
	}
	if wq.Salinity != 1.04 {
		t.Errorf("Salinity = %f, want 1.04", wq.Salinity)
	}
	if wq.SG != 1.005 {
		t.Errorf("SG = %f, want 1.005", wq.SG)
	}
}

func TestParseDPS_Empty(t *testing.T) {
	wq, err := parseDPS(map[string]any{})
	if err != nil {
		t.Fatalf("parseDPS: %v", err)
	}
	if wq.PH != 0 || wq.Temperature != 0 || wq.TDS != 0 {
		t.Error("empty DPS should result in zero values")
	}
}

func TestParseDPS_PartialData(t *testing.T) {
	dps := map[string]any{
		"1": float64(5000),
		"2": float64(250),
	}

	wq, err := parseDPS(dps)
	if err != nil {
		t.Fatalf("parseDPS: %v", err)
	}
	if wq.TDS != 5000 {
		t.Errorf("TDS = %d, want 5000", wq.TDS)
	}
	if wq.Temperature != 25.0 {
		t.Errorf("Temperature = %f, want 25.0", wq.Temperature)
	}
	if wq.PH != 0 {
		t.Errorf("PH = %f, want 0 (not present)", wq.PH)
	}
}

func TestApplyCalibration(t *testing.T) {
	wq := &WaterQuality{
		PH: 8.00, Temperature: 26.0, TDS: 9000, EC: 18000,
		ORP: 165, Salinity: 1.04, SG: 1.005, Battery: 100,
	}

	cal := map[string]float64{"sg": 0.020, "ph": -0.05}
	applyCalibration(wq, cal)

	if wq.SG != 1.025 {
		t.Errorf("SG = %f, want 1.025", wq.SG)
	}
	if wq.PH != 7.95 {
		t.Errorf("PH = %f, want 7.95", wq.PH)
	}
	// Uncalibrated fields unchanged
	if wq.TDS != 9000 {
		t.Errorf("TDS = %d, want 9000 (unchanged)", wq.TDS)
	}
}

func TestApplyCalibrationNil(t *testing.T) {
	wq := &WaterQuality{SG: 1.005}
	applyCalibration(wq, nil)
	if wq.SG != 1.005 {
		t.Errorf("SG = %f, want 1.005 (nil calibration should be no-op)", wq.SG)
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input any
		want  int
	}{
		{float64(42), 42},
		{float64(3.7), 3},
		{int(10), 10},
		{"123", 123},
		{"abc", 0},
		{true, 0},
		{nil, 0},
	}

	for _, tt := range tests {
		got := toInt(tt.input)
		if got != tt.want {
			t.Errorf("toInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
