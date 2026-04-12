package redsea

import "time"

// ATOTempLogEntry is one day's temperature data from the ATO cloud log.
// Avg contains one reading per Interval minutes (e.g. 96 readings for 15-min intervals).
type ATOTempLogEntry struct {
	Date     time.Time `json:"date"     yaml:"date"`
	Interval int       `json:"interval" yaml:"interval"`
	Avg      []float64 `json:"avg"      yaml:"avg"`
}

// ATOTemperatureLog is the response from the cloud temperature-log endpoint.
type ATOTemperatureLog struct {
	Entries []ATOTempLogEntry `json:"entries" yaml:"entries"`
}

// CloudNotification is a single in-app notification from the Red Sea cloud API.
type CloudNotification struct {
	ID          int       `json:"id"           yaml:"id"`
	Subject     string    `json:"subject"      yaml:"subject"`
	Text        string    `json:"text"         yaml:"text"`
	AquariumUID string    `json:"aquarium_uid" yaml:"aquarium_uid"`
	HWID        string    `json:"hwid"         yaml:"hwid"`
	DeviceType  string    `json:"device_type"  yaml:"device_type"`
	Type        string    `json:"type"         yaml:"type"`
	TimeSent    time.Time `json:"time_sent"    yaml:"time_sent"`
	Channel     string    `json:"channel"      yaml:"channel"`
	Read        bool      `json:"read"         yaml:"read"`
}
