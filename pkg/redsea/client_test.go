package redsea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGet(t *testing.T) {
	expected := DeviceInfo{
		HWModel: "RSATO+",
		HWID:    "abc123",
		Name:    "MyATO",
		Success: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device-info" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	// Extract host:port from test server URL
	ip := srv.Listener.Addr().String()
	client := New(ip, WithTimeout(2*time.Second), WithRetries(1, 100*time.Millisecond))

	var got DeviceInfo
	err := client.Get(context.Background(), "/device-info", &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.HWModel != expected.HWModel {
		t.Errorf("HWModel = %q, want %q", got.HWModel, expected.HWModel)
	}
	if got.HWID != expected.HWID {
		t.Errorf("HWID = %q, want %q", got.HWID, expected.HWID)
	}
	if got.Name != expected.Name {
		t.Errorf("Name = %q, want %q", got.Name, expected.Name)
	}
}

func TestClientRetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	ip := srv.Listener.Addr().String()
	client := New(ip, WithTimeout(2*time.Second), WithRetries(5, 10*time.Millisecond))

	var result map[string]string
	err := client.Get(context.Background(), "/", &result)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientNoRetryOn404(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.NotFound(w, r)
	}))
	defer srv.Close()

	ip := srv.Listener.Addr().String()
	client := New(ip, WithTimeout(2*time.Second), WithRetries(5, 10*time.Millisecond))

	var result map[string]string
	err := client.Get(context.Background(), "/missing", &result)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry on 404), got %d", attempts)
	}
}

func TestClientPost(t *testing.T) {
	var receivedPayload map[string]int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	ip := srv.Listener.Addr().String()
	client := New(ip, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	payload := map[string]int{"volume": 2500}
	err := client.Post(context.Background(), "/update-volume", payload, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPayload["volume"] != 2500 {
		t.Errorf("payload volume = %d, want 2500", receivedPayload["volume"])
	}
}
