package redsea

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const cloudBaseURL = "https://cloud.thereefbeat.com"

// CloudClient authenticates against the Red Sea cloud API and fetches
// cloud-only resources such as in-app notifications.
type CloudClient struct {
	username     string
	password     string
	clientCreds  string // base64-encoded "client_id:client_secret"
	tokenFile    string
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

type cloudToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// NewCloudClient creates a CloudClient. tokenFile is a path where the OAuth2
// token is cached between runs; pass "" to disable caching.
func NewCloudClient(username, password, clientCredentials, tokenFile string) *CloudClient {
	return &CloudClient{
		username:    username,
		password:    password,
		clientCreds: clientCredentials,
		tokenFile:   tokenFile,
		httpClient:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *CloudClient) loadToken() {
	if c.tokenFile == "" {
		return
	}
	data, err := os.ReadFile(c.tokenFile)
	if err != nil {
		return
	}
	var t cloudToken
	if err := json.Unmarshal(data, &t); err != nil {
		return
	}
	c.accessToken = t.AccessToken
	c.refreshToken = t.RefreshToken
	c.tokenExpiry = t.Expiry
}

func (c *CloudClient) saveToken() {
	if c.tokenFile == "" {
		return
	}
	data, err := json.Marshal(cloudToken{
		AccessToken:  c.accessToken,
		RefreshToken: c.refreshToken,
		Expiry:       c.tokenExpiry,
	})
	if err != nil {
		return
	}
	os.WriteFile(c.tokenFile, data, 0600)
}

func (c *CloudClient) authenticate(ctx context.Context, grantType string, extra url.Values) error {
	form := url.Values{"grant_type": {grantType}}
	if grantType == "password" {
		form.Set("username", c.username)
		form.Set("password", c.password)
	}
	for k, vs := range extra {
		form[k] = vs
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cloudBaseURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+c.clientCreds)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloud auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("cloud auth: HTTP %d", resp.StatusCode)
	}

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("cloud auth: decoding token: %w", err)
	}

	c.accessToken = token.AccessToken
	c.refreshToken = token.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)
	c.saveToken()
	return nil
}

func (c *CloudClient) ensureToken(ctx context.Context) error {
	if c.accessToken == "" {
		c.loadToken()
	}
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}
	if c.refreshToken != "" {
		if err := c.authenticate(ctx, "refresh_token", url.Values{"refresh_token": {c.refreshToken}}); err == nil {
			return nil
		}
		// refresh token expired — fall through to password login
	}
	return c.authenticate(ctx, "password", nil)
}

func (c *CloudClient) get(ctx context.Context, path string, result any) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloud GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("cloud GET %s: HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// GetATOTemperatureLog returns the temperature log for the given ATO device
// over the specified ISO 8601 duration (e.g. "P7D").
func (c *CloudClient) GetATOTemperatureLog(ctx context.Context, hwid, duration string) (*ATOTemperatureLog, error) {
	var entries []ATOTempLogEntry
	path := fmt.Sprintf("/reef-ato/%s/temperature-log?duration=%s", hwid, url.QueryEscape(duration))
	if err := c.get(ctx, path, &entries); err != nil {
		return nil, err
	}
	return &ATOTemperatureLog{Entries: entries}, nil
}

// GetNotifications returns up to size notifications from the last days days,
// newest first.
func (c *CloudClient) GetNotifications(ctx context.Context, days, size int) ([]CloudNotification, error) {
	var page struct {
		Content []CloudNotification `json:"content"`
	}
	path := fmt.Sprintf("/notification/inapp?expirationDays=%d&page=0&size=%d&sortDirection=DESC", days, size)
	if err := c.get(ctx, path, &page); err != nil {
		return nil, err
	}
	return page.Content, nil
}
