package eheim

import (
	"bytes"
	"context"
	"net"
	"os/exec"
	"strings"
	"time"
)

// resolveLocal resolves a .local hostname to an IP address.
// Go's DNS resolver is slow with .local hostnames (~5-10s) because it tries
// regular DNS first before falling back to mDNS. This function resolves
// via the system's `dns-sd` command which uses Bonjour directly (~100ms).
// Returns the original host unchanged if resolution fails or host is not .local.
func resolveLocal(host string) string {
	if !strings.HasSuffix(host, ".local") {
		return host
	}
	if net.ParseIP(host) != nil {
		return host
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// dns-sd -Q runs indefinitely, so we run it with a timeout and kill it
	// once we get a result. Output goes to stderr.
	cmd := exec.CommandContext(ctx, "dns-sd", "-Q", host)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return host
	}

	// Poll the output buffer for the IP
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			cmd.Wait()
			return host
		case <-ticker.C:
			if ip := extractIP(buf.String()); ip != "" {
				cmd.Process.Kill()
				cmd.Wait()
				return ip
			}
		}
	}
}

// extractIP looks for an A record line in dns-sd output and extracts the IP.
// Line format: "  1:45:16.480  Add  40000002  16  eheimdigital.local.  Addr  IN  192.168.50.81"
func extractIP(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Addr") && strings.Contains(line, "Add") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				candidate := fields[len(fields)-1]
				if net.ParseIP(candidate) != nil {
					return candidate
				}
			}
		}
	}
	return ""
}
