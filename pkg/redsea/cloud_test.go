package redsea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// newTestCloudClient returns a CloudClient pointed at srv with a pre-set
// access token so tests don't need to exercise the auth flow.
func newTestCloudClient(srv *httptest.Server) *CloudClient {
	c := NewCloudClient("user", "pass", "creds", "")
	c.accessToken = "test-token"
	c.tokenExpiry = time.Now().Add(time.Hour)
	// Point the HTTP client at the test server by overriding the base URL via
	// a transport that rewrites the host. Simpler: just replace httpClient with
	// one whose transport redirects all requests to the test server.
	c.httpClient = srv.Client()
	// Replace the hard-coded base URL by wrapping requests through the test
	// server. We do this by storing the server URL and patching get() — but
	// since cloudBaseURL is a package-level const we use a different approach:
	// set a custom RoundTripper that replaces the host.
	base, _ := url.Parse(srv.URL)
	c.httpClient.Transport = &rewriteTransport{base: base, inner: http.DefaultTransport}
	return c
}

// rewriteTransport replaces the scheme+host of every request with those of base.
type rewriteTransport struct {
	base  *url.URL
	inner http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.base.Scheme
	clone.URL.Host = t.base.Host
	return t.inner.RoundTrip(clone)
}

func TestGetNotifications(t *testing.T) {
	sent := time.Date(2026, 4, 11, 8, 51, 0, 0, time.UTC)
	notifications := []CloudNotification{
		{
			ID:         1,
			Subject:    "Low reservoir",
			Text:       "ATO: Reservoir is running low.",
			DeviceType: "reef-ato",
			TimeSent:   sent,
			Read:       false,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/notification/inapp" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"content": notifications})
	}))
	defer srv.Close()

	c := newTestCloudClient(srv)
	got, err := c.GetNotifications(context.Background(), 7, 50)
	if err != nil {
		t.Fatalf("GetNotifications: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != 1 {
		t.Errorf("ID = %d, want 1", got[0].ID)
	}
	if got[0].DeviceType != "reef-ato" {
		t.Errorf("DeviceType = %q, want %q", got[0].DeviceType, "reef-ato")
	}
	if !got[0].TimeSent.Equal(sent) {
		t.Errorf("TimeSent = %v, want %v", got[0].TimeSent, sent)
	}
}

func TestGetNotifications_QueryParams(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	}))
	defer srv.Close()

	c := newTestCloudClient(srv)
	_, err := c.GetNotifications(context.Background(), 7, 50)
	if err != nil {
		t.Fatalf("GetNotifications: %v", err)
	}
	if gotQuery != "expirationDays=7&page=0&size=50&sortDirection=DESC" {
		t.Errorf("query = %q, want expirationDays=7&page=0&size=50&sortDirection=DESC", gotQuery)
	}
}

func TestGetATOTemperatureLog(t *testing.T) {
	date := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	entries := []ATOTempLogEntry{
		{Date: date, Interval: 15, Avg: []float64{25.0, 25.2, 0, 25.4}}, // 0 = missing reading
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reef-ato/hwid-123/temperature-log" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(entries)
	}))
	defer srv.Close()

	c := newTestCloudClient(srv)
	log, err := c.GetATOTemperatureLog(context.Background(), "hwid-123", "P7D")
	if err != nil {
		t.Fatalf("GetATOTemperatureLog: %v", err)
	}
	if len(log.Entries) != 1 {
		t.Fatalf("len = %d, want 1", len(log.Entries))
	}
	if log.Entries[0].Interval != 15 {
		t.Errorf("Interval = %d, want 15", log.Entries[0].Interval)
	}
	if len(log.Entries[0].Avg) != 4 {
		t.Errorf("len(Avg) = %d, want 4", len(log.Entries[0].Avg))
	}
}

func TestGetATOTemperatureLog_QueryParam(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode([]ATOTempLogEntry{})
	}))
	defer srv.Close()

	c := newTestCloudClient(srv)
	_, err := c.GetATOTemperatureLog(context.Background(), "hwid-123", "P7D")
	if err != nil {
		t.Fatalf("GetATOTemperatureLog: %v", err)
	}
	if gotQuery != "duration=P7D" {
		t.Errorf("query = %q, want duration=P7D", gotQuery)
	}
}

func TestGetNotifications_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestCloudClient(srv)
	_, err := c.GetNotifications(context.Background(), 7, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
