package redsea

type LEDManualState struct {
	White       int     `json:"white" yaml:"white"`
	Blue        int     `json:"blue" yaml:"blue"`
	Moon        int     `json:"moon" yaml:"moon"`
	WhitePWM    int     `json:"white_pwm" yaml:"white_pwm"`
	BluePWM     int     `json:"blue_pwm" yaml:"blue_pwm"`
	MoonPWM     int     `json:"moon_pwm" yaml:"moon_pwm"`
	Fan         int     `json:"fan" yaml:"fan"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
}

type LEDManualSet struct {
	White int `json:"white"`
	Blue  int `json:"blue"`
	Moon  int `json:"moon"`
}

type LEDTimerSet struct {
	White    int `json:"white"`
	Blue     int `json:"blue"`
	Moon     int `json:"moon"`
	Duration int `json:"duration"` // milliseconds
}

type LEDSchedule struct {
	White LEDChannel `json:"white" yaml:"white"`
	Blue  LEDChannel `json:"blue" yaml:"blue"`
	Moon  LEDChannel `json:"moon" yaml:"moon"`
}

type LEDChannel struct {
	Rise   int              `json:"rise" yaml:"rise"`     // minute of day
	Set    int              `json:"set" yaml:"set"`       // minute of day
	Points []LEDSchedulePoint `json:"points" yaml:"points"`
}

type LEDSchedulePoint struct {
	T int `json:"t" yaml:"t"` // minutes from rise
	I int `json:"i" yaml:"i"` // intensity 0-100
}

type LEDAcclimation struct {
	Enabled                bool   `json:"enabled" yaml:"enabled"`
	Duration               int    `json:"duration" yaml:"duration"`
	StartIntensityFactor   int    `json:"start_intensity_factor" yaml:"start_intensity_factor"`
	CurrentIntensityFactor int    `json:"current_intensity_factor" yaml:"current_intensity_factor"`
	RemainingDays          int    `json:"remaining_days" yaml:"remaining_days"`
	StartedOn              string `json:"started_on" yaml:"started_on"`
}

type LEDMoonPhase struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	Name         string `json:"name" yaml:"name"`
	Intensity    int    `json:"intensity" yaml:"intensity"`
	TodaysMoonDay int   `json:"todays_moon_day" yaml:"todays_moon_day"`
	NextFullMoon int    `json:"next_full_moon" yaml:"next_full_moon"`
	NextNewMoon  int    `json:"next_new_moon" yaml:"next_new_moon"`
}

type LEDDashboard struct {
	White       int     `json:"white" yaml:"white"`
	Blue        int     `json:"blue" yaml:"blue"`
	Moon        int     `json:"moon" yaml:"moon"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	Mode        string  `json:"mode" yaml:"mode"`
}
