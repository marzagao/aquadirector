package eheim

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// digestTestServer creates a test server that requires Digest auth.
// handler is called after successful authentication.
func digestTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Digest ") {
			w.Header().Set("WWW-Authenticate", `Digest realm="asyncesp", qop="auth", nonce="testnonce123", opaque="testopaque456"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}))
}

func TestClientGet(t *testing.T) {
	expected := map[string]any{"title": "FEEDER_DATA", "weight": 45.3}

	srv := digestTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/autofeeder" {
			t.Errorf("path = %q, want /api/autofeeder", r.URL.Path)
		}
		if r.URL.Query().Get("to") != "AA:BB:CC" {
			t.Errorf("to = %q, want AA:BB:CC", r.URL.Query().Get("to"))
		}
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := New(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	var result map[string]any
	err := client.Get(context.Background(), "/api/autofeeder", map[string]string{"to": "AA:BB:CC"}, &result)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result["weight"] != 45.3 {
		t.Errorf("weight = %v, want 45.3", result["weight"])
	}
}

func TestClientPost(t *testing.T) {
	var receivedBody map[string]any
	srv := digestTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Write([]byte("success"))
	})
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := New(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	err := client.Post(context.Background(), "/api/autofeeder/feed", map[string]string{"to": "AA:BB:CC"})
	if err != nil {
		t.Fatalf("Post error: %v", err)
	}
	if receivedBody["to"] != "AA:BB:CC" {
		t.Errorf("to = %v, want AA:BB:CC", receivedBody["to"])
	}
}

func TestClientAuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Digest realm="asyncesp", qop="auth", nonce="testnonce", opaque="testopaque"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := New(host, WithTimeout(2*time.Second), WithRetries(1, 10*time.Millisecond))

	err := client.Get(context.Background(), "/api/autofeeder", nil, nil)
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("error = %v, want auth failure message", err)
	}
}

func TestClientConnectionFailure(t *testing.T) {
	client := New("127.0.0.1:1", WithTimeout(500*time.Millisecond), WithRetries(1, 10*time.Millisecond))

	err := client.Get(context.Background(), "/api/autofeeder", nil, nil)
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestClientRetryOnError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// No auth required for simplicity in retry test
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := New(host, WithTimeout(2*time.Second), WithRetries(5, 10*time.Millisecond))

	var result map[string]string
	err := client.Get(context.Background(), "/", nil, &result)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	// 2 failed + 1 success = 3 attempts minimum (Digest doubles requests,
	// but server doesn't require auth, so no 401s)
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestDigestAuthHeader(t *testing.T) {
	client := New("test", WithCredentials("api", "admin"))
	client.updateChallenge(`Digest realm="asyncesp", qop="auth", nonce="abc123", opaque="def456"`)

	header := client.buildAuthHeader("GET", "/api/autofeeder")

	if !strings.HasPrefix(header, "Digest ") {
		t.Errorf("header should start with 'Digest ', got: %s", header)
	}
	if !strings.Contains(header, `username="api"`) {
		t.Errorf("header missing username: %s", header)
	}
	if !strings.Contains(header, `realm="asyncesp"`) {
		t.Errorf("header missing realm: %s", header)
	}
	if !strings.Contains(header, `nonce="abc123"`) {
		t.Errorf("header missing nonce: %s", header)
	}
	if !strings.Contains(header, `opaque="def456"`) {
		t.Errorf("header missing opaque: %s", header)
	}
	if !strings.Contains(header, "qop=auth") {
		t.Errorf("header missing qop: %s", header)
	}
}

func TestParseDigestChallenge(t *testing.T) {
	challenge := `Digest realm="asyncesp", qop="auth", nonce="abc123", opaque="def456"`
	params := parseDigestChallenge(challenge)

	if params["realm"] != "asyncesp" {
		t.Errorf("realm = %q, want asyncesp", params["realm"])
	}
	if params["qop"] != "auth" {
		t.Errorf("qop = %q, want auth", params["qop"])
	}
	if params["nonce"] != "abc123" {
		t.Errorf("nonce = %q, want abc123", params["nonce"])
	}
	if params["opaque"] != "def456" {
		t.Errorf("opaque = %q, want def456", params["opaque"])
	}
}
