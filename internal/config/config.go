package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Network NetworkConfig  `mapstructure:"network"`
	Devices []DeviceConfig `mapstructure:"devices"`
	Sensor  SensorConfig   `mapstructure:"sensor"`
	Feeder  FeederConfig   `mapstructure:"feeder"`
	Alerts  AlertsConfig   `mapstructure:"alerts"`
	Cloud   CloudConfig    `mapstructure:"cloud"`
}

type CloudConfig struct {
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	ClientCredentials string `mapstructure:"client_credentials"`
}

type NetworkConfig struct {
	Subnet         string        `mapstructure:"subnet"`
	ScanThreads    int           `mapstructure:"scan_threads"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	RetryMax       int           `mapstructure:"retry_max"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
}

type DeviceConfig struct {
	Name string `mapstructure:"name"`
	IP   string `mapstructure:"ip"`
	Type string `mapstructure:"type"`
	UUID string `mapstructure:"uuid"`
	HWID string `mapstructure:"hwid"` // optional; auto-discovered from local /device-info
}

type SensorConfig struct {
	IP           string             `mapstructure:"ip"`
	Protocol     string             `mapstructure:"protocol"`
	Port         int                `mapstructure:"port"`
	PollInterval time.Duration      `mapstructure:"poll_interval"`
	DeviceID     string             `mapstructure:"device_id"`
	LocalKey     string             `mapstructure:"local_key"`
	Version      string             `mapstructure:"version"`
	Calibration  map[string]float64 `mapstructure:"calibration"`
}

type FeederConfig struct {
	Host     string `mapstructure:"host"`
	MAC      string `mapstructure:"mac"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type AlertsConfig struct {
	Enabled       bool                 `mapstructure:"enabled"`
	Rules         []AlertRule          `mapstructure:"rules"`
	Notifications []NotificationConfig `mapstructure:"notifications"`
}

type AlertRule struct {
	Name      string `mapstructure:"name"`
	Source    string `mapstructure:"source"`
	Metric    string `mapstructure:"metric"`
	Operator  string `mapstructure:"operator"`
	Threshold any    `mapstructure:"threshold"`
	Severity  string `mapstructure:"severity"`
	Message   string `mapstructure:"message"`
}

type NotificationConfig struct {
	Type         string            `mapstructure:"type"`
	SeverityMin  string            `mapstructure:"severity_min"`
	URL          string            `mapstructure:"url"`
	Method       string            `mapstructure:"method"`
	Headers      map[string]string `mapstructure:"headers"`
	BodyTemplate string            `mapstructure:"body_template"`
	Command      string            `mapstructure:"command"`
	Args         []string          `mapstructure:"args"`
}

func Load(v *viper.Viper) (*Config, error) {
	setDefaults(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("network.subnet", "192.168.50.0/24")
	v.SetDefault("network.scan_threads", 64)
	v.SetDefault("network.default_timeout", 20*time.Second)
	v.SetDefault("network.retry_max", 5)
	v.SetDefault("network.retry_delay", 2*time.Second)
	v.SetDefault("sensor.ip", "192.168.50.15")
	v.SetDefault("sensor.poll_interval", 60*time.Second)
	v.SetDefault("feeder.host", "eheimdigital.local")
	v.SetDefault("feeder.username", "api")
	v.SetDefault("feeder.password", "admin")
	v.SetDefault("alerts.enabled", true)
}

func (c *Config) DeviceByName(name string) *DeviceConfig {
	for i := range c.Devices {
		if c.Devices[i].Name == name {
			return &c.Devices[i]
		}
	}
	return nil
}

func (c *Config) DeviceByIP(ip string) *DeviceConfig {
	for i := range c.Devices {
		if c.Devices[i].IP == ip {
			return &c.Devices[i]
		}
	}
	return nil
}

func (c *Config) DeviceByType(deviceType string) *DeviceConfig {
	for i := range c.Devices {
		if c.Devices[i].Type == deviceType {
			return &c.Devices[i]
		}
	}
	return nil
}
