package sensor

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// KactoilyMAC is the known MAC OUI prefix for Kactoily/Tuya devices.
const KactoilyMAC = "3c:0b:59"

// FindDeviceByMAC pings the subnet to populate the ARP table,
// then searches for a device matching the Kactoily MAC prefix.
// Returns the IP if found, or empty string.
func FindDeviceByMAC(subnet string) string {
	// Ping sweep to populate ARP cache
	cidr := subnet
	if cidr == "" {
		cidr = "192.168.50.0/24"
	}

	ips := subnetIPs(cidr)
	for _, ip := range ips {
		go func(ip string) {
			conn, err := net.DialTimeout("tcp", ip+":6668", 500*time.Millisecond)
			if err == nil {
				conn.Close()
			}
		}(ip)
	}
	// Also do a quick ICMP sweep
	for _, ip := range ips {
		go exec.Command("ping", "-c", "1", "-W", "1", ip).Run()
	}

	// Wait briefly for ARP table to populate
	time.Sleep(2 * time.Second)

	// Parse ARP table
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, KactoilyMAC) {
			// Extract IP from arp output: "? (192.168.x.x) at mac ..."
			start := strings.Index(line, "(")
			end := strings.Index(line, ")")
			if start >= 0 && end > start {
				return line[start+1 : end]
			}
		}
	}

	return ""
}

func subnetIPs(cidr string) []string {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}
	return ips
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ResolveIP tries the configured IP first by doing a quick TCP probe.
// If that fails, falls back to MAC-based ARP discovery.
// Returns the working IP and whether it was rediscovered.
func ResolveIP(configuredIP, subnet string) (ip string, rediscovered bool) {
	// Quick check: can we connect to the configured IP?
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", configuredIP, 6668), 2*time.Second)
	if err == nil {
		conn.Close()
		return configuredIP, false
	}

	// Configured IP failed — search by MAC
	found := FindDeviceByMAC(subnet)
	if found != "" {
		return found, true
	}

	// Fall back to configured IP (will likely fail but gives a clear error)
	return configuredIP, false
}
