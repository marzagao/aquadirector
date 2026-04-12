package redsea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestATODashboard(t *testing.T) {
	expected := ATODashboard{
		VolumeLeft: 2500,
		Mode:       "auto",
		IsPumpOn:   false,
		PumpState:  "off",
		PumpSpeed:  100,
		TotalFills: 36,
		FlowRate:   1176,
		ATOSensor: ATOSensor{
			Connected:    true,
			CurrentLevel: "desired",
		},
		LeakSensor: LeakSensor{
			Connected: true,
			Enabled:   true,
			Status:    "dry",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dashboard":
			json.NewEncoder(w).Encode(expected)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewATOClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	dash, err := client.Dashboard(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dash.VolumeLeft != 2500 {
		t.Errorf("VolumeLeft = %d, want 2500", dash.VolumeLeft)
	}
	if dash.Mode != "auto" {
		t.Errorf("Mode = %q, want %q", dash.Mode, "auto")
	}
	if dash.LeakSensor.Status != "dry" {
		t.Errorf("LeakSensor.Status = %q, want %q", dash.LeakSensor.Status, "dry")
	}
}

func TestATOResume(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/resume" && r.Method == http.MethodPost {
			called = true
			w.WriteHeader(200)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewATOClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	err := client.Resume(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("resume endpoint was not called")
	}
}

func TestATOSetVolume(t *testing.T) {
	var receivedPayload map[string]int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/update-volume" && r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&receivedPayload)
			w.WriteHeader(200)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewATOClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	err := client.SetVolume(context.Background(), 3000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPayload["volume"] != 3000 {
		t.Errorf("volume = %d, want 3000", receivedPayload["volume"])
	}
}
