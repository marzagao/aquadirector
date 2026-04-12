package eheim

import (
	"encoding/json"
	"fmt"
)

// DrumState represents the food drum fill level indicator.
type DrumState int

const (
	DrumGreen     DrumState = 0
	DrumOrange    DrumState = 1
	DrumRed       DrumState = 2
	DrumMeasuring DrumState = 5
)

func (d DrumState) String() string {
	switch d {
	case DrumGreen:
		return "GREEN"
	case DrumOrange:
		return "ORANGE"
	case DrumRed:
		return "RED"
	case DrumMeasuring:
		return "MEASURING"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(d))
	}
}

// DeviceType identifies Eheim device types by version number.
const (
	DeviceTypeFeeder = 6
)

// DeviceTypeName returns a human-readable name for a device version.
func DeviceTypeName(version int) string {
	switch version {
	case 3:
		return "LEDcontrol"
	case 4:
		return "professional 5e"
	case 5:
		return "thermocontrol+e"
	case DeviceTypeFeeder:
		return "autofeeder+"
	case 9:
		return "pHcontrol+e"
	case 17:
		return "classicLEDcontrol+e"
	case 18:
		return "classicVARIO+e"
	default:
		return fmt.Sprintf("unknown (v%d)", version)
	}
}

// FeedSlot is a single feeding event within a day.
type FeedSlot struct {
	TimeMinutes int `json:"time_minutes"` // minutes since midnight
	Turns       int `json:"turns"`        // drum turns (portions)
}

// TimeString returns HH:MM format.
func (f FeedSlot) TimeString() string {
	return fmt.Sprintf("%02d:%02d", f.TimeMinutes/60, f.TimeMinutes%60)
}

// DaySchedule holds up to 2 feeding slots for a single day.
type DaySchedule struct {
	Slots []FeedSlot `json:"slots"`
}

// WeekSchedule holds the feeding schedule for an entire week (Mon=0 ... Sun=6).
type WeekSchedule [7]DaySchedule

// DayNames maps index to day name.
var DayNames = [7]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

// ShortDayNames maps index to abbreviated day name.
var ShortDayNames = [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// FeederData is the parsed autofeeder status from a FEEDER_DATA message.
type FeederData struct {
	Weight          float64      `json:"weight"`
	IsSpinning      bool         `json:"is_spinning"`
	Level           int          `json:"level"`
	DrumState       DrumState    `json:"drum_state"`
	Schedule        WeekSchedule `json:"schedule"`
	Overfeeding     bool         `json:"overfeeding"`
	SyncFilter      string       `json:"sync_filter"` // MAC of paired filter (reduces flow before feeding)
	FeedingBreak    bool         `json:"feeding_break"`
	IsBreakDay      bool         `json:"is_break_day"`
	TurnTimeFeeding int          `json:"turn_time_feeding"`
}

// feederDataWire is the raw JSON structure from the Eheim protocol.
type feederDataWire struct {
	Title           string    `json:"title"`
	From            string    `json:"from"`
	Weight          float64   `json:"weight"`
	IsSpinning      int       `json:"isSpinning"`
	Level           []int     `json:"level"`
	Configuration   [][][]int `json:"configuration"`
	Overfeeding     int       `json:"overfeeding"`
	Sync            string    `json:"sync"`
	PartnerName     string    `json:"partnerName"`
	SollRegulation  int       `json:"sollRegulation"`
	FeedingBreak    int       `json:"feedingBreak"`
	BreakDay        int       `json:"breakDay"`
	TurnTimeFeeding int       `json:"turnTimeFeeding"`
}

// ParseFeederData converts a raw JSON message into FeederData.
func ParseFeederData(data []byte) (*FeederData, error) {
	var wire feederDataWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("parsing feeder data: %w", err)
	}
	return wireToFeederData(&wire)
}

// ParseFeederDataMap converts a decoded JSON map into FeederData.
func ParseFeederDataMap(msg map[string]any) (*FeederData, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("re-marshaling feeder data: %w", err)
	}
	return ParseFeederData(data)
}

func wireToFeederData(wire *feederDataWire) (*FeederData, error) {
	fd := &FeederData{
		Weight:          wire.Weight,
		IsSpinning:      wire.IsSpinning != 0,
		Overfeeding:     wire.Overfeeding != 0,
		SyncFilter:      wire.Sync,
		FeedingBreak:    wire.FeedingBreak != 0,
		IsBreakDay:      wire.BreakDay != 0,
		TurnTimeFeeding: wire.TurnTimeFeeding,
	}

	// Parse level: [level_value, drum_state]
	if len(wire.Level) >= 2 {
		fd.Level = wire.Level[0]
		fd.DrumState = DrumState(wire.Level[1])
	}

	// Parse configuration: 7-element array (Mon-Sun), each [[times], [turns]]
	fd.Schedule = parseSchedule(wire.Configuration)

	return fd, nil
}

func parseSchedule(config [][][]int) WeekSchedule {
	var week WeekSchedule
	for i := 0; i < 7 && i < len(config); i++ {
		day := config[i]
		if len(day) < 2 {
			continue
		}
		times := day[0]
		turns := day[1]
		for j := 0; j < len(times) && j < len(turns); j++ {
			week[i].Slots = append(week[i].Slots, FeedSlot{
				TimeMinutes: times[j],
				Turns:       turns[j],
			})
		}
	}
	return week
}

// scheduleToWire converts a WeekSchedule back to the wire format.
func scheduleToWire(sched WeekSchedule) [][][]int {
	config := make([][][]int, 7)
	for i := 0; i < 7; i++ {
		times := make([]int, 0, len(sched[i].Slots))
		turns := make([]int, 0, len(sched[i].Slots))
		for _, slot := range sched[i].Slots {
			times = append(times, slot.TimeMinutes)
			turns = append(turns, slot.Turns)
		}
		config[i] = [][]int{times, turns}
	}
	return config
}

// FeederConfigUpdate holds optional fields for POST /api/autofeeder/config.
// Nil fields are not changed.
type FeederConfigUpdate struct {
	Overfeeding *bool
	SyncFilter  *string
}

// MeshDevice represents a device discovered on the Eheim mesh network.
type MeshDevice struct {
	MAC      string `json:"mac" yaml:"mac"`
	IP       string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Name     string `json:"name" yaml:"name"`
	Version  int    `json:"version" yaml:"version"`
	Revision string `json:"revision" yaml:"revision"`
}
