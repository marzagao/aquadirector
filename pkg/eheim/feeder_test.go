package eheim

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

var sampleFeederDataWire = feederDataWire{
	Title:           "FEEDER_DATA",
	From:            "AA:BB:CC:DD:EE:FF",
	Weight:          45.3,
	IsSpinning:      0,
	Level:           []int{75, 0},
	Configuration:   [][][]int{{{480, 1080}, {2, 3}}, {{}, {}}, {{480}, {1}}, {{}, {}}, {{}, {}}, {{720, 1080}, {1, 1}}, {{}, {}}},
	Overfeeding:     1,
	Sync:            "",
	PartnerName:     "",
	SollRegulation:  0,
	FeedingBreak:    1,
	BreakDay:        0,
	TurnTimeFeeding: 5,
}

func feederTestServer(t *testing.T) (string, func()) {
	t.Helper()
	srv := digestTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/autofeeder":
			json.NewEncoder(w).Encode(sampleFeederDataWire)
		case r.Method == http.MethodPost && r.URL.Path == "/api/autofeeder/feed":
			w.Write([]byte("success"))
		case r.Method == http.MethodPost && r.URL.Path == "/api/autofeeder/full":
			w.Write([]byte("success"))
		case r.Method == http.MethodPost && r.URL.Path == "/api/autofeeder/bio":
			w.Write([]byte("success"))
		case r.Method == http.MethodPost && r.URL.Path == "/api/autofeeder/config":
			w.Write([]byte("success"))
		case r.Method == http.MethodGet && r.URL.Path == "/api/devicelist":
			json.NewEncoder(w).Encode(map[string]any{
				"clientList":   []string{"AA:BB:CC:DD:EE:FF"},
				"clientIPList": []string{"192.168.1.81"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/userdata":
			json.NewEncoder(w).Encode(map[string]any{
				"name":     "My Feeder",
				"version":  6,
				"revision": []int{2050},
			})
		default:
			http.NotFound(w, r)
		}
	})

	host := strings.TrimPrefix(srv.URL, "http://")
	return host, srv.Close
}

func TestFeederClientStatus(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	client := NewFeederClient(host, "AA:BB:CC:DD:EE:FF",
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	fd, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}

	if fd.Weight != 45.3 {
		t.Errorf("Weight = %v, want 45.3", fd.Weight)
	}
	if fd.DrumState != DrumGreen {
		t.Errorf("DrumState = %v, want GREEN", fd.DrumState)
	}
	if fd.Level != 75 {
		t.Errorf("Level = %d, want 75", fd.Level)
	}
	if !fd.Overfeeding {
		t.Error("Overfeeding = false, want true")
	}
	if !fd.FeedingBreak {
		t.Error("FeedingBreak = false, want true")
	}

	// Check schedule
	if len(fd.Schedule[0].Slots) != 2 {
		t.Fatalf("Monday slots = %d, want 2", len(fd.Schedule[0].Slots))
	}
	if fd.Schedule[0].Slots[0].TimeMinutes != 480 {
		t.Errorf("Mon slot 0 time = %d, want 480", fd.Schedule[0].Slots[0].TimeMinutes)
	}
}

func TestFeederClientFeed(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	client := NewFeederClient(host, "AA:BB:CC:DD:EE:FF",
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	if err := client.Feed(context.Background()); err != nil {
		t.Fatalf("Feed error: %v", err)
	}
}

func TestFeederClientMarkDrumFull(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	client := NewFeederClient(host, "AA:BB:CC:DD:EE:FF",
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	if err := client.MarkDrumFull(context.Background()); err != nil {
		t.Fatalf("MarkDrumFull error: %v", err)
	}
}

func TestFeederClientSetSchedule(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	client := NewFeederClient(host, "AA:BB:CC:DD:EE:FF",
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	sched := WeekSchedule{}
	sched[0] = DaySchedule{Slots: []FeedSlot{{TimeMinutes: 480, Turns: 2}}}

	if err := client.SetSchedule(context.Background(), sched, true); err != nil {
		t.Fatalf("SetSchedule error: %v", err)
	}
}

func TestFeederClientSetConfig(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	client := NewFeederClient(host, "AA:BB:CC:DD:EE:FF",
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	overfeeding := false
	if err := client.SetConfig(context.Background(), FeederConfigUpdate{
		Overfeeding: &overfeeding,
	}); err != nil {
		t.Fatalf("SetConfig error: %v", err)
	}
}

func TestFeederClientMAC(t *testing.T) {
	client := NewFeederClient("host", "AA:BB:CC:DD:EE:FF")
	if client.MAC() != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", client.MAC())
	}
}

func TestHubFindFeeder(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	hub := NewHubClient(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	mac, err := hub.FindFeeder(context.Background())
	if err != nil {
		t.Fatalf("FindFeeder error: %v", err)
	}
	if mac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", mac)
	}
}

func TestHubMeshDevices(t *testing.T) {
	host, cleanup := feederTestServer(t)
	defer cleanup()

	hub := NewHubClient(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	devices, err := hub.MeshDevices(context.Background())
	if err != nil {
		t.Fatalf("MeshDevices error: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("got %d devices, want 1", len(devices))
	}
	if devices[0].MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", devices[0].MAC)
	}
	if devices[0].Name != "My Feeder" {
		t.Errorf("Name = %q, want My Feeder", devices[0].Name)
	}
	if devices[0].Version != 6 {
		t.Errorf("Version = %d, want 6", devices[0].Version)
	}
	if devices[0].IP != "192.168.1.81" {
		t.Errorf("IP = %q, want 192.168.1.81", devices[0].IP)
	}
}

func TestHubFindFeederNotFound(t *testing.T) {
	srv := digestTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/devicelist":
			json.NewEncoder(w).Encode(map[string]any{
				"clientList":   []string{"AA:BB:CC:DD:EE:01"},
				"clientIPList": []string{"192.168.1.82"},
			})
		case "/api/userdata":
			json.NewEncoder(w).Encode(map[string]any{
				"name":     "heater",
				"version":  5,
				"revision": []int{1000},
			})
		}
	})
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	hub := NewHubClient(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	_, err := hub.FindFeeder(context.Background())
	if err == nil {
		t.Fatal("expected error for no feeder found")
	}
	if _, ok := err.(*DeviceNotFoundError); !ok {
		t.Errorf("expected DeviceNotFoundError, got %T: %v", err, err)
	}
}
