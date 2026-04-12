package redsea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLEDManualState(t *testing.T) {
	expected := LEDManualState{
		White:       100,
		Blue:        50,
		Moon:        10,
		Temperature: 28.5,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manual" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(expected)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewLEDClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	state, err := client.ManualState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.White != 100 {
		t.Errorf("White = %d, want 100", state.White)
	}
	if state.Blue != 50 {
		t.Errorf("Blue = %d, want 50", state.Blue)
	}
	if state.Moon != 10 {
		t.Errorf("Moon = %d, want 10", state.Moon)
	}
}

func TestLEDSetManual(t *testing.T) {
	var received LEDManualSet
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manual" && r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&received)
			w.WriteHeader(200)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewLEDClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	err := client.SetManual(context.Background(), LEDManualSet{White: 200, Blue: 100, Moon: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.White != 200 || received.Blue != 100 || received.Moon != 5 {
		t.Errorf("received = %+v, want white=200 blue=100 moon=5", received)
	}
}

func TestLEDSchedule(t *testing.T) {
	expected := LEDSchedule{
		White: LEDChannel{Rise: 660, Set: 1260, Points: []LEDSchedulePoint{{T: 120, I: 100}}},
		Blue:  LEDChannel{Rise: 660, Set: 1341, Points: []LEDSchedulePoint{{T: 60, I: 100}}},
		Moon:  LEDChannel{Rise: 1345, Set: 1523, Points: []LEDSchedulePoint{{T: 75, I: 10}}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auto/1" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(expected)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewLEDClient(srv.Listener.Addr().String(),
		WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	sched, err := client.Schedule(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sched.White.Rise != 660 {
		t.Errorf("White.Rise = %d, want 660", sched.White.Rise)
	}
	if len(sched.White.Points) != 1 || sched.White.Points[0].I != 100 {
		t.Errorf("unexpected white points: %+v", sched.White.Points)
	}
}

func TestLEDScheduleInvalidDay(t *testing.T) {
	client := NewLEDClient("127.0.0.1",
		WithTimeout(100*time.Millisecond), WithRetries(1, 10*time.Millisecond))

	_, err := client.Schedule(context.Background(), 0)
	if err == nil {
		t.Error("expected error for day 0")
	}

	_, err = client.Schedule(context.Background(), 8)
	if err == nil {
		t.Error("expected error for day 8")
	}
}
