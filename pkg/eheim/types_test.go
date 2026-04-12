package eheim

import (
	"testing"
)

func TestParseFeederData(t *testing.T) {
	raw := `{
		"title": "FEEDER_DATA",
		"from": "AA:BB:CC:DD:EE:FF",
		"weight": 45.3,
		"isSpinning": 0,
		"level": [75, 0],
		"configuration": [
			[[180, 1200], [1, 2]],
			[[], []],
			[[180, 1200], [1, 2]],
			[[180, 1200], [1, 2]],
			[[], []],
			[[180, 1200], [1, 2]],
			[[180, 1200], [1, 2]]
		],
		"overfeeding": 1,
		"sync": "",
		"partnerName": "",
		"sollRegulation": 0,
		"feedingBreak": 1,
		"breakDay": 0,
		"turnTimeFeeding": 5
	}`

	fd, err := ParseFeederData([]byte(raw))
	if err != nil {
		t.Fatalf("ParseFeederData error: %v", err)
	}

	if fd.Weight != 45.3 {
		t.Errorf("Weight = %v, want 45.3", fd.Weight)
	}
	if fd.IsSpinning {
		t.Error("IsSpinning = true, want false")
	}
	if fd.Level != 75 {
		t.Errorf("Level = %d, want 75", fd.Level)
	}
	if fd.DrumState != DrumGreen {
		t.Errorf("DrumState = %v, want GREEN", fd.DrumState)
	}
	if !fd.Overfeeding {
		t.Error("Overfeeding = false, want true")
	}
	if !fd.FeedingBreak {
		t.Error("FeedingBreak = false, want true")
	}
	if fd.IsBreakDay {
		t.Error("IsBreakDay = true, want false")
	}
	if fd.SyncFilter != "" {
		t.Errorf("SyncFilter = %q, want empty", fd.SyncFilter)
	}
	if fd.TurnTimeFeeding != 5 {
		t.Errorf("TurnTimeFeeding = %d, want 5", fd.TurnTimeFeeding)
	}
}

func TestParseFeederDataSchedule(t *testing.T) {
	raw := `{
		"title": "FEEDER_DATA",
		"from": "AA:BB:CC:DD:EE:FF",
		"weight": 10.0,
		"isSpinning": 0,
		"level": [50, 1],
		"configuration": [
			[[480, 1080], [2, 3]],
			[[], []],
			[[480], [1]],
			[[], []],
			[[], []],
			[[720, 1080], [1, 1]],
			[[], []]
		],
		"overfeeding": 0,
		"sync": "",
		"partnerName": "",
		"sollRegulation": 0,
		"feedingBreak": 0,
		"breakDay": 0,
		"turnTimeFeeding": 5
	}`

	fd, err := ParseFeederData([]byte(raw))
	if err != nil {
		t.Fatalf("ParseFeederData error: %v", err)
	}

	if fd.DrumState != DrumOrange {
		t.Errorf("DrumState = %v, want ORANGE", fd.DrumState)
	}

	// Monday: 08:00 (2 turns), 18:00 (3 turns)
	if len(fd.Schedule[0].Slots) != 2 {
		t.Fatalf("Monday slots = %d, want 2", len(fd.Schedule[0].Slots))
	}
	if fd.Schedule[0].Slots[0].TimeMinutes != 480 {
		t.Errorf("Mon slot 0 time = %d, want 480", fd.Schedule[0].Slots[0].TimeMinutes)
	}
	if fd.Schedule[0].Slots[0].Turns != 2 {
		t.Errorf("Mon slot 0 turns = %d, want 2", fd.Schedule[0].Slots[0].Turns)
	}
	if fd.Schedule[0].Slots[1].TimeMinutes != 1080 {
		t.Errorf("Mon slot 1 time = %d, want 1080", fd.Schedule[0].Slots[1].TimeMinutes)
	}
	if fd.Schedule[0].Slots[1].Turns != 3 {
		t.Errorf("Mon slot 1 turns = %d, want 3", fd.Schedule[0].Slots[1].Turns)
	}

	// Tuesday: no feeding
	if len(fd.Schedule[1].Slots) != 0 {
		t.Errorf("Tuesday slots = %d, want 0", len(fd.Schedule[1].Slots))
	}

	// Wednesday: 08:00 (1 turn)
	if len(fd.Schedule[2].Slots) != 1 {
		t.Fatalf("Wednesday slots = %d, want 1", len(fd.Schedule[2].Slots))
	}
	if fd.Schedule[2].Slots[0].TimeMinutes != 480 || fd.Schedule[2].Slots[0].Turns != 1 {
		t.Errorf("Wed slot = {%d, %d}, want {480, 1}", fd.Schedule[2].Slots[0].TimeMinutes, fd.Schedule[2].Slots[0].Turns)
	}

	// Saturday: 12:00 (1 turn), 18:00 (1 turn)
	if len(fd.Schedule[5].Slots) != 2 {
		t.Fatalf("Saturday slots = %d, want 2", len(fd.Schedule[5].Slots))
	}
	if fd.Schedule[5].Slots[0].TimeMinutes != 720 {
		t.Errorf("Sat slot 0 time = %d, want 720", fd.Schedule[5].Slots[0].TimeMinutes)
	}
}

func TestFeedSlotTimeString(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{0, "00:00"},
		{180, "03:00"},
		{480, "08:00"},
		{720, "12:00"},
		{1080, "18:00"},
		{1200, "20:00"},
		{1439, "23:59"},
	}

	for _, tt := range tests {
		slot := FeedSlot{TimeMinutes: tt.minutes}
		if got := slot.TimeString(); got != tt.want {
			t.Errorf("FeedSlot{%d}.TimeString() = %q, want %q", tt.minutes, got, tt.want)
		}
	}
}

func TestDrumStateString(t *testing.T) {
	tests := []struct {
		state DrumState
		want  string
	}{
		{DrumGreen, "GREEN"},
		{DrumOrange, "ORANGE"},
		{DrumRed, "RED"},
		{DrumMeasuring, "MEASURING"},
		{DrumState(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("DrumState(%d).String() = %q, want %q", int(tt.state), got, tt.want)
		}
	}
}

func TestScheduleRoundTrip(t *testing.T) {
	original := WeekSchedule{
		{Slots: []FeedSlot{{TimeMinutes: 480, Turns: 2}, {TimeMinutes: 1080, Turns: 3}}},
		{Slots: nil},
		{Slots: []FeedSlot{{TimeMinutes: 480, Turns: 1}}},
		{Slots: nil},
		{Slots: nil},
		{Slots: []FeedSlot{{TimeMinutes: 720, Turns: 1}, {TimeMinutes: 1080, Turns: 1}}},
		{Slots: nil},
	}

	wire := scheduleToWire(original)
	parsed := parseSchedule(wire)

	for i := 0; i < 7; i++ {
		if len(parsed[i].Slots) != len(original[i].Slots) {
			t.Errorf("day %d: slot count = %d, want %d", i, len(parsed[i].Slots), len(original[i].Slots))
			continue
		}
		for j := range original[i].Slots {
			if parsed[i].Slots[j] != original[i].Slots[j] {
				t.Errorf("day %d slot %d = %+v, want %+v", i, j, parsed[i].Slots[j], original[i].Slots[j])
			}
		}
	}
}

func TestParseFeederDataSpinning(t *testing.T) {
	raw := `{
		"title": "FEEDER_DATA",
		"from": "AA:BB:CC:DD:EE:FF",
		"weight": 0.0,
		"isSpinning": 1,
		"level": [0, 2],
		"configuration": [[[], []], [[], []], [[], []], [[], []], [[], []], [[], []], [[], []]],
		"overfeeding": 0,
		"sync": "11:22:33:44:55:66",
		"partnerName": "MyFilter",
		"sollRegulation": 1,
		"feedingBreak": 0,
		"breakDay": 1,
		"turnTimeFeeding": 3
	}`

	fd, err := ParseFeederData([]byte(raw))
	if err != nil {
		t.Fatalf("ParseFeederData error: %v", err)
	}

	if !fd.IsSpinning {
		t.Error("IsSpinning = false, want true")
	}
	if fd.DrumState != DrumRed {
		t.Errorf("DrumState = %v, want RED", fd.DrumState)
	}
	if fd.SyncFilter != "11:22:33:44:55:66" {
		t.Errorf("SyncFilter = %q, want 11:22:33:44:55:66", fd.SyncFilter)
	}
	if !fd.IsBreakDay {
		t.Error("IsBreakDay = false, want true")
	}
}

func TestDeviceTypeName(t *testing.T) {
	if got := DeviceTypeName(6); got != "autofeeder+" {
		t.Errorf("DeviceTypeName(6) = %q, want autofeeder+", got)
	}
	if got := DeviceTypeName(5); got != "thermocontrol+e" {
		t.Errorf("DeviceTypeName(5) = %q, want thermocontrol+e", got)
	}
}
