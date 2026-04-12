package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadDefaults(t *testing.T) {
	v := viper.New()
	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Network.Subnet != "192.168.1.0/24" {
		t.Errorf("subnet = %q, want 192.168.1.0/24", cfg.Network.Subnet)
	}
	if cfg.Network.ScanThreads != 64 {
		t.Errorf("scan_threads = %d, want 64", cfg.Network.ScanThreads)
	}
	if cfg.Network.DefaultTimeout != 20*time.Second {
		t.Errorf("timeout = %v, want 20s", cfg.Network.DefaultTimeout)
	}
	if cfg.Network.RetryMax != 5 {
		t.Errorf("retry_max = %d, want 5", cfg.Network.RetryMax)
	}
	if cfg.Network.RetryDelay != 2*time.Second {
		t.Errorf("retry_delay = %v, want 2s", cfg.Network.RetryDelay)
	}
	if cfg.Sensor.IP != "192.168.1.15" {
		t.Errorf("sensor.ip = %q, want 192.168.1.15", cfg.Sensor.IP)
	}
	if !cfg.Alerts.Enabled {
		t.Error("alerts should be enabled by default")
	}
}

func TestLoadFromViper(t *testing.T) {
	v := viper.New()
	v.Set("network.subnet", "10.0.0.0/8")
	v.Set("sensor.device_id", "test123")
	v.Set("sensor.local_key", "abc")
	v.Set("sensor.version", "3.5")

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Network.Subnet != "10.0.0.0/8" {
		t.Errorf("subnet = %q, want 10.0.0.0/8", cfg.Network.Subnet)
	}
	if cfg.Sensor.DeviceID != "test123" {
		t.Errorf("device_id = %q, want test123", cfg.Sensor.DeviceID)
	}
	if cfg.Sensor.LocalKey != "abc" {
		t.Errorf("local_key = %q, want abc", cfg.Sensor.LocalKey)
	}
}

func TestDeviceLookup(t *testing.T) {
	cfg := &Config{
		Devices: []DeviceConfig{
			{Name: "MyATO", IP: "192.168.1.10", Type: "RSATO+"},
			{Name: "MyLED", IP: "192.168.1.20", Type: "RSLED60"},
		},
	}

	if d := cfg.DeviceByName("MyATO"); d == nil || d.IP != "192.168.1.10" {
		t.Error("DeviceByName(MyATO) failed")
	}
	if d := cfg.DeviceByName("nonexistent"); d != nil {
		t.Error("DeviceByName should return nil for unknown name")
	}
	if d := cfg.DeviceByIP("192.168.1.20"); d == nil || d.Name != "MyLED" {
		t.Error("DeviceByIP(192.168.1.20) failed")
	}
	if d := cfg.DeviceByIP("10.0.0.1"); d != nil {
		t.Error("DeviceByIP should return nil for unknown IP")
	}
	if d := cfg.DeviceByType("RSLED60"); d == nil || d.Name != "MyLED" {
		t.Error("DeviceByType(RSLED60) failed")
	}
	if d := cfg.DeviceByType("RSWAVE25"); d != nil {
		t.Error("DeviceByType should return nil for unknown type")
	}
}
