package discovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsKnownModel(t *testing.T) {
	known := []string{"RSATO+", "RSLED60", "RSLED50", "RSLED90", "RSLED160", "RSLED115", "RSLED170", "RSDOSE2", "RSDOSE4", "RSMAT", "RSRUN", "RSWAVE25", "RSWAVE45"}
	for _, m := range known {
		if !IsKnownModel(m) {
			t.Errorf("IsKnownModel(%q) = false, want true", m)
		}
	}

	if IsKnownModel("UNKNOWN") {
		t.Error("IsKnownModel(UNKNOWN) = true, want false")
	}
}

func TestIsG2LED(t *testing.T) {
	if !IsG2LED("RSLED60") {
		t.Error("IsG2LED(RSLED60) = false, want true")
	}
	if !IsG2LED("RSLED115") {
		t.Error("IsG2LED(RSLED115) = false, want true")
	}
	if IsG2LED("RSLED50") {
		t.Error("IsG2LED(RSLED50) = true, want false (G1)")
	}
	if IsG2LED("RSATO+") {
		t.Error("IsG2LED(RSATO+) = true, want false")
	}
}

func TestScan_MockDevice(t *testing.T) {
	deviceInfo := map[string]any{
		"hw_model": "RSATO+",
		"hwid":     "abc123",
		"name":     "TestATO",
		"success":  true,
	}
	firmware := map[string]any{
		"version": "1.0.0",
		"success": true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device-info":
			json.NewEncoder(w).Encode(deviceInfo)
		case "/firmware":
			json.NewEncoder(w).Encode(firmware)
		case "/description.xml":
			w.Write([]byte(`<?xml version="1.0"?><root><device><UDN>uuid:test-uuid-123</UDN></device></root>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// Extract IP:port — we can't scan a subnet for a test server,
	// but we can test probeDevice directly
	addr := srv.Listener.Addr().String()
	dev := probeDevice(context.Background(), addr)
	if dev == nil {
		t.Fatal("probeDevice returned nil for mock device")
	}
	if dev.HWModel != "RSATO+" {
		t.Errorf("HWModel = %q, want RSATO+", dev.HWModel)
	}
	if dev.Name != "TestATO" {
		t.Errorf("Name = %q, want TestATO", dev.Name)
	}
	if dev.UUID != "test-uuid-123" {
		t.Errorf("UUID = %q, want test-uuid-123", dev.UUID)
	}
	if dev.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", dev.Firmware)
	}
}

func TestScan_NonDevice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	dev := probeDevice(context.Background(), addr)
	if dev != nil {
		t.Error("probeDevice should return nil for non-device")
	}
}
