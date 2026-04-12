package sensor

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var commonPorts = []struct {
	port    int
	service string
}{
	{80, "HTTP"},
	{443, "HTTPS"},
	{8080, "HTTP-alt"},
	{3000, "HTTP-alt"},
	{9090, "HTTP-alt"},
	{1883, "MQTT"},
	{8883, "MQTT-TLS"},
	{6668, "Tuya"},
}

var httpProbePaths = []string{
	"/",
	"/api/status",
	"/status",
	"/data",
	"/sensor",
	"/device-info",
	"/api/v1/data",
}

func Probe(ctx context.Context, ip string, extraPorts []int) (*ProbeResult, error) {
	result := &ProbeResult{IP: ip}

	// Stage 1: Port scan
	ports := scanPorts(ctx, ip)
	result.OpenPorts = ports

	if len(ports) == 0 {
		result.Details = "No open ports found. Device may be offline or firewalled."
		return result, nil
	}

	// Stage 2: HTTP probing on open ports
	var httpDetails []string
	for _, p := range ports {
		if !isHTTPPort(p.Port) {
			continue
		}
		for _, path := range httpProbePaths {
			resp := probeHTTP(ctx, ip, p.Port, path)
			if resp != "" {
				httpDetails = append(httpDetails, fmt.Sprintf("  %s:%d%s -> %s", ip, p.Port, path, truncate(resp, 200)))
			}
		}
	}

	// Check for MQTT
	for _, p := range ports {
		if p.Port == 1883 || p.Port == 8883 {
			result.Protocol = "mqtt"
			break
		}
	}

	// Check for Tuya
	for _, p := range ports {
		if p.Port == 6668 {
			result.Protocol = "tuya"
			break
		}
	}

	// If HTTP responses found, prefer HTTP protocol identification
	if len(httpDetails) > 0 {
		if result.Protocol == "" {
			result.Protocol = "http"
		}
		result.Details = "HTTP responses:\n" + strings.Join(httpDetails, "\n")
	} else if result.Protocol != "" {
		result.Details = fmt.Sprintf("Detected %s protocol on open ports", result.Protocol)
	} else {
		var portStrs []string
		for _, p := range ports {
			portStrs = append(portStrs, fmt.Sprintf("%d(%s)", p.Port, p.Service))
		}
		result.Details = "Open ports: " + strings.Join(portStrs, ", ") + " but no HTTP responses"
	}

	return result, nil
}

func scanPorts(ctx context.Context, ip string) []PortInfo {
	var (
		mu    sync.Mutex
		ports []PortInfo
		wg    sync.WaitGroup
	)

	for _, p := range commonPorts {
		wg.Add(1)
		go func(port int, service string) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			conn.Close()

			mu.Lock()
			ports = append(ports, PortInfo{Port: port, Open: true, Service: service})
			mu.Unlock()
		}(p.port, p.service)
	}

	wg.Wait()
	return ports
}

func probeHTTP(ctx context.Context, ip string, port int, path string) string {
	url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
	client := &http.Client{Timeout: 3 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return ""
	}

	return fmt.Sprintf("[%d] %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func isHTTPPort(port int) bool {
	return port == 80 || port == 443 || port == 8080 || port == 3000 || port == 9090
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
