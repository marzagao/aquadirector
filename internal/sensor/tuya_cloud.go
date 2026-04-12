package sensor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const tuyaCloudBaseURL = "https://openapi.tuyaus.com"

type TuyaCloud struct {
	ClientID     string
	ClientSecret string
	accessToken  string
	httpClient   *http.Client
}

func NewTuyaCloud(clientID, clientSecret string) *TuyaCloud {
	return &TuyaCloud{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

type tuyaTokenResponse struct {
	Result struct {
		AccessToken string `json:"access_token"`
		ExpireTime  int    `json:"expire_time"`
	} `json:"result"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

type TuyaDeviceInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LocalKey string `json:"local_key"`
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
}

type tuyaDeviceListResponse struct {
	Result []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Key  string `json:"key"`
		IP   string `json:"ip"`
		MAC  string `json:"mac"`
	} `json:"result"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

func (c *TuyaCloud) authenticate() error {
	path := "/v1.0/token?grant_type=1"
	body, err := c.signedRequest("GET", path, "")
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}

	var resp tuyaTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("tuya auth failed: %s", resp.Msg)
	}

	c.accessToken = resp.Result.AccessToken
	return nil
}

// GetDevices returns all linked devices with their local keys.
func (c *TuyaCloud) GetDevices() ([]TuyaDeviceInfo, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	path := "/v1.0/users/" + c.ClientID + "/devices"
	body, err := c.signedRequest("GET", path, "")
	if err != nil {
		return nil, fmt.Errorf("device list request failed: %w", err)
	}

	// Try the user devices endpoint first
	var resp tuyaDeviceListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing device list: %w", err)
	}

	// If user endpoint returned nothing, try the iot-03 endpoint
	if !resp.Success || len(resp.Result) == 0 {
		path = "/v1.0/iot-03/devices"
		body, err = c.signedRequest("GET", path, "")
		if err != nil {
			return nil, fmt.Errorf("device list request failed: %w", err)
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing device list: %w", err)
		}
	}

	if !resp.Success {
		return nil, fmt.Errorf("tuya get devices failed: %s", resp.Msg)
	}

	var devices []TuyaDeviceInfo
	for _, d := range resp.Result {
		devices = append(devices, TuyaDeviceInfo{
			ID:       d.ID,
			Name:     d.Name,
			LocalKey: d.Key,
			IP:       d.IP,
			MAC:      d.MAC,
		})
	}
	return devices, nil
}

func (c *TuyaCloud) signedRequest(method, path, body string) ([]byte, error) {
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	url := tuyaCloudBaseURL + path

	// Build string to sign
	var sb strings.Builder
	sb.WriteString(c.ClientID)
	if c.accessToken != "" {
		sb.WriteString(c.accessToken)
	}
	sb.WriteString(ts)

	// For GET with no body, the content hash is the sha256 of empty string
	contentHash := sha256Hex([]byte(body))
	stringToSign := method + "\n" + contentHash + "\n\n" + path
	sb.WriteString(stringToSign)

	sign := hmacSHA256([]byte(c.ClientSecret), []byte(sb.String()))

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("client_id", c.ClientID)
	req.Header.Set("sign", strings.ToUpper(sign))
	req.Header.Set("t", ts)
	req.Header.Set("sign_method", "HMAC-SHA256")
	if c.accessToken != "" {
		req.Header.Set("access_token", c.accessToken)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
