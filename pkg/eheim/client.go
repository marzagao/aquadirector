package eheim

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the base HTTP transport for Eheim Digital devices.
// Uses Digest authentication (realm "asyncesp").
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration

	// Cached digest challenge to avoid the initial 401 on subsequent requests.
	mu           sync.Mutex
	cachedNonce  string
	cachedOpaque string
	cachedRealm  string
	cachedQOP    string
	nonceCount   int
}

// Option configures a Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetries sets the max retry count and delay between retries.
func WithRetries(max int, delay time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = max
		c.retryDelay = delay
	}
}

// WithCredentials sets the Digest auth username and password.
func WithCredentials(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// New creates a new Eheim HTTP client.
func New(host string, opts ...Option) *Client {
	resolved := resolveLocal(host)
	c := &Client{
		baseURL:  "http://" + resolved,
		username: "api",
		password: "admin",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Host returns the configured hub host.
func (c *Client) Host() string {
	return strings.TrimPrefix(c.baseURL, "http://")
}

// Get performs a GET request with Digest auth and decodes the JSON response.
func (c *Client) Get(ctx context.Context, path string, query map[string]string, result any) error {
	return c.doWithRetry(ctx, http.MethodGet, path, query, nil, result)
}

// Post performs a POST request with Digest auth and a JSON body.
func (c *Client) Post(ctx context.Context, path string, payload any) error {
	return c.doWithRetry(ctx, http.MethodPost, path, nil, payload, nil)
}

func (c *Client) doWithRetry(ctx context.Context, method, path string, query map[string]string, payload, result any) error {
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		err := c.doDigestRequest(ctx, method, path, query, payload, result)
		if err == nil {
			return nil
		}

		// Don't retry on 400/404
		if reqErr, ok := err.(*RequestFailedError); ok {
			if reqErr.StatusCode == 400 || reqErr.StatusCode == 404 {
				return err
			}
		}

		lastErr = err
	}
	return fmt.Errorf("after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) buildURL(path string, query map[string]string) string {
	url := c.baseURL + path
	if len(query) > 0 {
		params := make([]string, 0, len(query))
		for k, v := range query {
			params = append(params, k+"="+v)
		}
		url += "?" + strings.Join(params, "&")
	}
	return url
}

// doDigestRequest performs HTTP Digest authentication (RFC 2617).
// Uses a cached nonce from prior requests to skip the initial 401 when possible.
func (c *Client) doDigestRequest(ctx context.Context, method, path string, query map[string]string, payload, result any) error {
	url := c.buildURL(path, query)

	// Try with cached nonce first
	c.mu.Lock()
	hasCached := c.cachedNonce != ""
	c.mu.Unlock()

	if hasCached {
		resp, err := c.doRequestWithAuth(ctx, method, url, path, payload)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			if resp.StatusCode >= 400 {
				return &RequestFailedError{Method: method, Path: path, StatusCode: resp.StatusCode}
			}
			return decodeResponse(resp, result)
		}
		// Nonce was stale, fall through to get a fresh one
		io.Copy(io.Discard, resp.Body)
		c.updateChallenge(resp.Header.Get("WWW-Authenticate"))
	}

	// Step 1: Send request without auth to get the challenge
	body, err := marshalBody(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &ConnectionError{Host: c.Host(), Err: err}
	}
	// Drain body so connection can be reused
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		if resp.StatusCode >= 400 {
			return &RequestFailedError{Method: method, Path: path, StatusCode: resp.StatusCode}
		}
		return nil
	}

	// Step 2: Cache the challenge and retry with auth
	challenge := resp.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return &ProtocolError{Op: "auth", Msg: "401 without WWW-Authenticate header"}
	}
	c.updateChallenge(challenge)

	resp2, err := c.doRequestWithAuth(ctx, method, url, path, payload)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusUnauthorized {
		return &ProtocolError{Op: "auth", Msg: "digest authentication failed (check username/password)"}
	}
	if resp2.StatusCode >= 400 {
		return &RequestFailedError{Method: method, Path: path, StatusCode: resp2.StatusCode}
	}

	return decodeResponse(resp2, result)
}

func (c *Client) doRequestWithAuth(ctx context.Context, method, url, path string, payload any) (*http.Response, error) {
	body, err := marshalBody(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", c.buildAuthHeader(method, path))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ConnectionError{Host: c.Host(), Err: err}
	}
	return resp, nil
}

func (c *Client) updateChallenge(header string) {
	params := parseDigestChallenge(header)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedRealm = params["realm"]
	c.cachedNonce = params["nonce"]
	c.cachedOpaque = params["opaque"]
	c.cachedQOP = params["qop"]
	c.nonceCount = 0
}

func (c *Client) buildAuthHeader(method, uri string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nonceCount++

	ha1 := md5hex(c.username + ":" + c.cachedRealm + ":" + c.password)
	ha2 := md5hex(method + ":" + uri)

	cnonce := fmt.Sprintf("%08x", rand.Uint32())
	nc := fmt.Sprintf("%08x", c.nonceCount)

	var response string
	if c.cachedQOP == "auth" {
		response = md5hex(ha1 + ":" + c.cachedNonce + ":" + nc + ":" + cnonce + ":" + c.cachedQOP + ":" + ha2)
	} else {
		response = md5hex(ha1 + ":" + c.cachedNonce + ":" + ha2)
	}

	header := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		c.username, c.cachedRealm, c.cachedNonce, uri, response)
	if c.cachedQOP != "" {
		header += fmt.Sprintf(`, qop=%s, nc=%s, cnonce="%s"`, c.cachedQOP, nc, cnonce)
	}
	if c.cachedOpaque != "" {
		header += fmt.Sprintf(`, opaque="%s"`, c.cachedOpaque)
	}

	return header
}

func decodeResponse(resp *http.Response, result any) error {
	if result == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

func marshalBody(payload any) (io.Reader, error) {
	if payload == nil {
		return nil, nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}
	return bytes.NewReader(data), nil
}

func parseDigestChallenge(header string) map[string]string {
	params := make(map[string]string)
	header = strings.TrimPrefix(header, "Digest ")

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if idx := strings.IndexByte(part, '='); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			val := strings.TrimSpace(part[idx+1:])
			val = strings.Trim(val, `"`)
			params[key] = val
		}
	}
	return params
}

func md5hex(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
