package discovery

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/marzagao/aquadirector/pkg/eheim"
	"github.com/marzagao/aquadirector/pkg/redsea"
)

type DiscoveredDevice struct {
	IP       string `json:"ip" yaml:"ip"`
	HWModel  string `json:"hw_model" yaml:"hw_model"`
	Name     string `json:"name" yaml:"name"`
	UUID     string `json:"uuid" yaml:"uuid"`
	Firmware string `json:"firmware" yaml:"firmware"`
}

type DiscoveredEheimDevice struct {
	Host     string `json:"host" yaml:"host"`
	IP       string `json:"ip,omitempty" yaml:"ip,omitempty"`
	MAC      string `json:"mac" yaml:"mac"`
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	Version  int    `json:"version" yaml:"version"`
	Revision string `json:"revision" yaml:"revision"`
}

type ScanResult struct {
	Devices      []DiscoveredDevice      `json:"devices" yaml:"devices"`
	EheimDevices []DiscoveredEheimDevice `json:"eheim_devices,omitempty" yaml:"eheim_devices,omitempty"`
}

func Scan(ctx context.Context, subnet string, threads int) (*ScanResult, error) {
	ips, err := subnetIPs(subnet)
	if err != nil {
		return nil, fmt.Errorf("parsing subnet %s: %w", subnet, err)
	}

	if threads <= 0 {
		threads = 64
	}

	var (
		mu      sync.Mutex
		devices []DiscoveredDevice
		wg      sync.WaitGroup
		sem     = make(chan struct{}, threads)
	)

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return &ScanResult{Devices: devices}, ctx.Err()
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			dev := probeDevice(ctx, ip)
			if dev != nil {
				mu.Lock()
				devices = append(devices, *dev)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	return &ScanResult{Devices: devices}, nil
}

func probeDevice(ctx context.Context, ip string) *DiscoveredDevice {
	client := redsea.New(ip, redsea.WithTimeout(2e9)) // 2s timeout for discovery

	info, err := client.DeviceInfo(ctx)
	if err != nil || !info.Success {
		return nil
	}

	if !IsKnownModel(info.HWModel) {
		return nil
	}

	dev := &DiscoveredDevice{
		IP:      ip,
		HWModel: info.HWModel,
		Name:    info.Name,
	}

	if uuid, err := client.DescriptionUUID(ctx); err == nil {
		dev.UUID = uuid
	}

	fw, err := client.Firmware(ctx)
	if err == nil {
		dev.Firmware = fw.Version
	}

	return dev
}

func subnetIPs(cidr string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast addresses for /24 and smaller
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ScanEheim connects to an Eheim hub and returns all devices on the mesh.
func ScanEheim(ctx context.Context, host string) ([]DiscoveredEheimDevice, error) {
	hub := eheim.NewHubClient(host, eheim.WithTimeout(5e9))
	meshDevices, err := hub.MeshDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanning Eheim mesh at %s: %w", host, err)
	}

	var devices []DiscoveredEheimDevice
	for _, d := range meshDevices {
		devices = append(devices, DiscoveredEheimDevice{
			Host:     host,
			IP:       d.IP,
			MAC:      d.MAC,
			Name:     d.Name,
			Type:     eheim.DeviceTypeName(d.Version),
			Version:  d.Version,
			Revision: d.Revision,
		})
	}
	return devices, nil
}
