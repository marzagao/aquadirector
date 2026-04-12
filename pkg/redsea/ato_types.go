package redsea

type ATODashboard struct {
	VolumeLeft        int        `json:"volume_left" yaml:"volume_left"`
	DailyFillsAvg     *float64   `json:"daily_fills_average" yaml:"daily_fills_average"`
	DailyVolumeAvg    *float64   `json:"daily_volume_average" yaml:"daily_volume_average"`
	DaysTillEmpty     *float64   `json:"days_till_empty" yaml:"days_till_empty"`
	Mode              string     `json:"mode" yaml:"mode"`
	IsPumpOn          bool       `json:"is_pump_on" yaml:"is_pump_on"`
	PumpState         string     `json:"pump_state" yaml:"pump_state"`
	LastPumpOnCause   string     `json:"last_pump_on_cause" yaml:"last_pump_on_cause"`
	PumpSpeed         int        `json:"pump_speed" yaml:"pump_speed"`
	TotalFills        int        `json:"total_fills" yaml:"total_fills"`
	TodayFills        int        `json:"today_fills" yaml:"today_fills"`
	TodayVolumeUsage  int        `json:"today_volume_usage" yaml:"today_volume_usage"`
	FlowRate          float64    `json:"flow_rate" yaml:"flow_rate"`
	LastFillDate      *int64     `json:"last_fill_date" yaml:"last_fill_date"`
	WaterLevel        string     `json:"water_level" yaml:"water_level"`
	ATOSensor         ATOSensor  `json:"ato_sensor" yaml:"ato_sensor"`
	LeakSensor        LeakSensor `json:"leak_sensor" yaml:"leak_sensor"`
}

// Temperature returns the water temperature in Celsius from the ATO sensor probe.
// Returns 0 if the probe is not connected.
func (d *ATODashboard) Temperature() float64 {
	if d.ATOSensor.TemperatureProbeStatus != "connected" {
		return 0
	}
	return d.ATOSensor.CurrentRead
}

// HasTemperature reports whether a valid temperature reading is available.
func (d *ATODashboard) HasTemperature() bool {
	return d.ATOSensor.TemperatureProbeStatus == "connected"
}

type ATOSensor struct {
	Connected              bool    `json:"connected" yaml:"connected"`
	CurrentLevel           string  `json:"current_level" yaml:"current_level"`
	CurrentRead            float64 `json:"current_read" yaml:"current_read"`
	IsCalibrated           bool    `json:"is_calibrated" yaml:"is_calibrated"`
	IsTempEnabled          bool    `json:"is_temp_enabled" yaml:"is_temp_enabled"`
	TemperatureProbeStatus string  `json:"temperature_probe_status" yaml:"temperature_probe_status"`
	TemperatureLogEnabled  bool    `json:"temperature_log_enabled" yaml:"temperature_log_enabled"`
}

type LeakSensor struct {
	Connected     bool   `json:"connected" yaml:"connected"`
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	BuzzerEnabled bool   `json:"buzzer_enabled" yaml:"buzzer_enabled"`
	Status        string `json:"status" yaml:"status"`
	CurrentRead   int    `json:"current_read" yaml:"current_read"`
}

type ATOConfiguration struct {
	AutoFill    bool          `json:"auto_fill" yaml:"auto_fill"`
	AutoDelay   int           `json:"auto_delay" yaml:"auto_delay"`
	TempEnabled bool          `json:"temp_enabled" yaml:"temp_enabled"`
	Buzzer      ATOBuzzer     `json:"buzzer" yaml:"buzzer"`
	Leak        ATOLeakConfig `json:"leak" yaml:"leak"`
	Temperature ATOTempConfig `json:"temperature" yaml:"temperature"`
}

type ATOBuzzer struct {
	Enabled   bool `json:"enabled" yaml:"enabled"`
	Frequency int  `json:"frequency" yaml:"frequency"`
	DutyCycle int  `json:"duty_cycle" yaml:"duty_cycle"`
}

type ATOLeakConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	SensorEnabled     bool `json:"sensor_enabled" yaml:"sensor_enabled"`
	DryThreshold      int  `json:"dry_threshold" yaml:"dry_threshold"`
	RodiThreshold     int  `json:"rodi_threshold" yaml:"rodi_threshold"`
	EmergencyShutdown bool `json:"emergency_shutdown" yaml:"emergency_shutdown"`
}

type ATOTempConfig struct {
	Enabled             bool    `json:"enabled" yaml:"enabled"`
	LogEnabled          bool    `json:"log_enabled" yaml:"log_enabled"`
	AcceptableRangeLow  float64 `json:"acceptable_range_low" yaml:"acceptable_range_low"`
	AcceptableRangeHigh float64 `json:"acceptable_range_high" yaml:"acceptable_range_high"`
	DesiredRangeLow     float64 `json:"desired_range_low" yaml:"desired_range_low"`
	DesiredRangeHigh    float64 `json:"desired_range_high" yaml:"desired_range_high"`
	Offset              float64 `json:"offset" yaml:"offset"`
}
