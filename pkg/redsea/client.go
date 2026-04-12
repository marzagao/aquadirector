package redsea

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func WithRetries(max int, delay time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = max
		c.retryDelay = delay
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

func New(ip string, opts ...Option) *Client {
	c := &Client{
		baseURL:    "http://" + ip,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		maxRetries: 5,
		retryDelay: 2 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) IP() string {
	// Extract IP from baseURL "http://IP"
	return c.baseURL[len("http://"):]
}

func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doWithRetry(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, payload, result any) error {
	return c.doWithRetry(ctx, http.MethodPost, path, payload, result)
}

func (c *Client) Put(ctx context.Context, path string, payload, result any) error {
	return c.doWithRetry(ctx, http.MethodPut, path, payload, result)
}

func (c *Client) doWithRetry(ctx context.Context, method, path string, payload, result any) error {
	url := c.baseURL + path
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		err := c.doRequest(ctx, method, url, path, payload, result)
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

func (c *Client) doRequest(ctx context.Context, method, url, path string, payload, result any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshaling payload: %w", err)
		}
		body = bytes.NewReader(data)
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
		return &DeviceUnreachableError{IP: c.IP(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &RequestFailedError{Method: method, Path: path, StatusCode: resp.StatusCode}
	}

	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}
	}

	return nil
}
