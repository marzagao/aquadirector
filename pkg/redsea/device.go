package redsea

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type DeviceInfo struct {
	HWModel    string `json:"hw_model"`
	HWRevision string `json:"hw_revision"`
	HWID       string `json:"hwid"`
	Name       string `json:"name"`
	Success    bool   `json:"success"`
}

type FirmwareInfo struct {
	Version          string `json:"version"`
	Framework        string `json:"framework"`
	Board            string `json:"board"`
	FrameworkVersion string `json:"framework_version"`
	ChipRevision     string `json:"chip_revision"`
	Success          bool   `json:"success"`
}

type ModeResponse struct {
	Mode string `json:"mode"`
}

type WifiInfo struct {
	IP        string `json:"ip"`
	Connected bool   `json:"connected"`
}

func (c *Client) DeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	var info DeviceInfo
	if err := c.Get(ctx, "/device-info", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) Firmware(ctx context.Context) (*FirmwareInfo, error) {
	var info FirmwareInfo
	if err := c.Get(ctx, "/firmware", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) Mode(ctx context.Context) (string, error) {
	var resp ModeResponse
	if err := c.Get(ctx, "/mode", &resp); err != nil {
		return "", err
	}
	return resp.Mode, nil
}

func (c *Client) SetMode(ctx context.Context, mode string) error {
	return c.Post(ctx, "/mode", ModeResponse{Mode: mode}, nil)
}

func (c *Client) Wifi(ctx context.Context) (*WifiInfo, error) {
	var info WifiInfo
	if err := c.Get(ctx, "/wifi", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// DescriptionUUID fetches /description.xml and extracts the UUID from the UDN element.
func (c *Client) DescriptionUUID(ctx context.Context) (string, error) {
	url := c.baseURL + "/description.xml"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("description.xml returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return parseUDN(body)
}

func parseUDN(xmlData []byte) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(xmlData)))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok {
			localName := se.Name.Local
			if localName == "UDN" {
				var udn string
				if err := decoder.DecodeElement(&udn, &se); err != nil {
					return "", err
				}
				udn = strings.TrimSpace(udn)
				udn = strings.TrimPrefix(udn, "uuid:")
				return udn, nil
			}
		}
	}
	return "", fmt.Errorf("UDN element not found in description.xml")
}
