package eheim

import (
	"context"
	"fmt"
)

// HubClient provides hub-level operations (discovery, mesh enumeration).
type HubClient struct {
	*Client
}

// NewHubClient creates a new Eheim hub client.
func NewHubClient(host string, opts ...Option) *HubClient {
	return &HubClient{Client: New(host, opts...)}
}

// MeshDevices returns all devices on the Eheim mesh network.
func (h *HubClient) MeshDevices(ctx context.Context) ([]MeshDevice, error) {
	// GET /api/devicelist returns MESH_NETWORK with clientList
	var mesh struct {
		ClientList   []string `json:"clientList"`
		ClientIPList []string `json:"clientIPList"`
	}
	if err := h.Get(ctx, "/api/devicelist", nil, &mesh); err != nil {
		return nil, fmt.Errorf("fetching device list: %w", err)
	}

	var devices []MeshDevice
	for i, mac := range mesh.ClientList {
		// GET /api/userdata?to=MAC returns USRDTA for that device
		var usrdta struct {
			Name     string `json:"name"`
			Version  int    `json:"version"`
			Revision []int  `json:"revision"`
		}
		query := map[string]string{"to": mac}
		if err := h.Get(ctx, "/api/userdata", query, &usrdta); err != nil {
			continue
		}

		dev := MeshDevice{
			MAC:     mac,
			Name:    usrdta.Name,
			Version: usrdta.Version,
		}
		if i < len(mesh.ClientIPList) {
			dev.IP = mesh.ClientIPList[i]
		}
		if len(usrdta.Revision) > 0 {
			r := usrdta.Revision[0]
			dev.Revision = fmt.Sprintf("%d.%02d.%d", r/1000, (r%1000)/10, r%10)
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

// FindFeeder returns the MAC of the first autofeeder+ on the mesh.
// Returns MultipleDevicesError if more than one feeder is found.
func (h *HubClient) FindFeeder(ctx context.Context) (string, error) {
	devices, err := h.MeshDevices(ctx)
	if err != nil {
		return "", err
	}

	var feeders []string
	for _, d := range devices {
		if d.Version == DeviceTypeFeeder {
			feeders = append(feeders, d.MAC)
		}
	}

	switch len(feeders) {
	case 0:
		return "", &DeviceNotFoundError{}
	case 1:
		return feeders[0], nil
	default:
		return "", &MultipleDevicesError{MACs: feeders}
	}
}
