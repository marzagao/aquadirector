package tuya

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	DefaultPort    = 6668
	DefaultTimeout = 5 * time.Second

	CmdDPQueryNew uint32 = 0x10
)

type Device struct {
	ID         string
	IP         string
	Port       int
	LocalKey   []byte
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

func NewDevice(id, ip string, port int, localKey string) *Device {
	if port == 0 {
		port = DefaultPort
	}
	return &Device{
		ID:         id,
		IP:         ip,
		Port:       port,
		LocalKey:   []byte(localKey),
		Timeout:    DefaultTimeout,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

type statusResponse struct {
	DPS map[string]any `json:"dps"`
	T   int64          `json:"t"`
}

// Status queries the device for its current data points.
// Retries up to MaxRetries times on transient failures (EOF, timeout).
func (d *Device) Status(ctx context.Context) (map[string]any, error) {
	var lastErr error
	for attempt := 0; attempt < d.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(d.RetryDelay):
			}
		}

		dps, err := d.statusOnce(ctx)
		if err == nil {
			return dps, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("after %d attempts: %w", d.MaxRetries, lastErr)
}

func (d *Device) statusOnce(ctx context.Context) (map[string]any, error) {
	addr := fmt.Sprintf("%s:%d", d.IP, d.Port)

	dialer := &net.Dialer{Timeout: d.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()

	sessionKey, err := d.negotiate(conn)
	if err != nil {
		return nil, err
	}

	return d.queryDPS(conn, sessionKey)
}

// negotiate performs the 3-way session key exchange.
// All handshake packets use 6699 format encrypted with the local key.
func (d *Device) negotiate(conn net.Conn) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(d.Timeout))

	// Step 1: Send client nonce encrypted with local key
	clientNonce, err := GenerateNonce()
	if err != nil {
		return nil, err
	}

	pkt, err := Encode6699(1, CmdSessKeyNegStart, clientNonce[:], d.LocalKey)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(pkt); err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("send nonce failed: %v", err)}
	}

	// Step 2: Read device response (encrypted with local key)
	respData, err := ReadPacket(conn)
	if err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("read negotiate response failed: %v", err)}
	}

	cmd, payload, err := Decode6699(respData, d.LocalKey)
	if err != nil {
		return nil, err
	}
	if cmd != CmdSessKeyNegResp {
		return nil, &ProtocolError{Op: "negotiate", Msg: fmt.Sprintf("expected cmd 0x%02X, got 0x%02X", CmdSessKeyNegResp, cmd)}
	}

	// The device response may include a 4-byte retcode prefix before the actual data.
	// Decrypted payload is either 48 bytes (nonce+HMAC) or 52 bytes (retcode+nonce+HMAC).
	if len(payload) >= 52 {
		// Skip 4-byte retcode prefix
		payload = payload[4:]
	}
	if len(payload) < 48 {
		return nil, &ProtocolError{Op: "negotiate", Msg: fmt.Sprintf("response payload too short: %d bytes, need 48", len(payload))}
	}

	deviceNonce := payload[:16]
	deviceHMAC := payload[16:48]

	// Verify device HMAC
	if !VerifyHMAC(d.LocalKey, clientNonce[:], deviceHMAC) {
		return nil, &CryptoError{Op: "verify_device_hmac", Err: fmt.Errorf("device HMAC verification failed")}
	}

	// Derive session key
	sessionKey, err := DeriveSessionKey(d.LocalKey, clientNonce[:], deviceNonce)
	if err != nil {
		return nil, err
	}

	// Step 3: Send HMAC of device nonce encrypted with local key
	finishHMAC := ComputeHMAC(d.LocalKey, deviceNonce)
	finishPkt, err := Encode6699(2, CmdSessKeyNegFinish, finishHMAC, d.LocalKey)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(finishPkt); err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("send finish failed: %v", err)}
	}

	return sessionKey, nil
}

func (d *Device) queryDPS(conn net.Conn, sessionKey []byte) (map[string]any, error) {
	conn.SetDeadline(time.Now().Add(d.Timeout))

	// v3.5 uses cmd 0x10 (DP_QUERY_NEW) with empty JSON payload
	pkt, err := Encode6699(3, CmdDPQueryNew, []byte("{}"), sessionKey)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(pkt); err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("send query failed: %v", err)}
	}

	// Read response
	respData, err := ReadPacket(conn)
	if err != nil {
		return nil, &DeviceError{IP: d.IP, Msg: fmt.Sprintf("read query response failed: %v", err)}
	}

	_, payload, err := Decode6699(respData, sessionKey)
	if err != nil {
		return nil, err
	}

	// Strip any leading non-JSON bytes (retcode, source header, version header)
	payload = stripToJSON(payload)

	var resp statusResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, &ProtocolError{Op: "parse_response", Msg: fmt.Sprintf("invalid JSON: %v (payload: %q)", err, truncateBytes(payload, 100))}
	}

	return resp.DPS, nil
}

// stripToJSON finds the first '{' in the payload and returns from there.
func stripToJSON(data []byte) []byte {
	for i, b := range data {
		if b == '{' {
			return data[i:]
		}
	}
	return data
}

func truncateBytes(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}
	return b[:max]
}
