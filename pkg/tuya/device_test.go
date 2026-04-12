package tuya

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

// mockTuyaDevice simulates a Tuya v3.5 device for testing.
// All packets use 6699 format. Handshake encrypted with localKey, data with sessionKey.
func mockTuyaDevice(t *testing.T, localKey []byte, dpsFixture map[string]any) (addr string, stop func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		serveMockTuyaSession(t, conn, localKey, dpsFixture)
	}()

	return ln.Addr().String(), func() {
		ln.Close()
		<-done
	}
}

// serveMockTuyaSession runs the v3.5 handshake + single DP query exchange.
func serveMockTuyaSession(t *testing.T, conn net.Conn, localKey []byte, dpsFixture map[string]any) {
	t.Helper()

	// Handshake start: client sends SESS_KEY_NEG_START with its nonce.
	clientNonce, err := readExpectedCmd(conn, localKey, CmdSessKeyNegStart)
	if err != nil {
		t.Errorf("mock: start: %v", err)
		return
	}

	// Handshake response: device replies with its nonce + HMAC of client nonce.
	deviceNonce, _ := GenerateNonce()
	respPayload := append(deviceNonce[:], ComputeHMAC(localKey, clientNonce)...)
	if err := writeEncoded(conn, CmdSessKeyNegResp, respPayload, localKey); err != nil {
		t.Errorf("mock: resp: %v", err)
		return
	}

	// Handshake finish: client confirms.
	if _, err := readExpectedCmd(conn, localKey, CmdSessKeyNegFinish); err != nil {
		t.Errorf("mock: finish: %v", err)
		return
	}

	sessionKey, err := DeriveSessionKey(localKey, clientNonce, deviceNonce[:])
	if err != nil {
		t.Errorf("mock: derive key: %v", err)
		return
	}

	// Data phase: client queries DPs, device responds with the fixture.
	if _, err := readExpectedCmd(conn, sessionKey, CmdDPQueryNew); err != nil {
		t.Errorf("mock: query: %v", err)
		return
	}
	respJSON, _ := json.Marshal(map[string]any{
		"dps": dpsFixture,
		"t":   time.Now().Unix(),
	})
	if err := writeEncoded(conn, CmdDPQueryNew, respJSON, sessionKey); err != nil {
		t.Errorf("mock: data resp: %v", err)
	}
}

// readExpectedCmd reads a 6699 packet, decodes it, and asserts the command matches.
func readExpectedCmd(conn net.Conn, key []byte, want uint32) ([]byte, error) {
	pktData, err := ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	cmd, payload, err := Decode6699(pktData, key)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if cmd != want {
		return nil, fmt.Errorf("cmd=0x%02X want=0x%02X", cmd, want)
	}
	return payload, nil
}

// writeEncoded builds a 6699 packet and writes it to the connection.
func writeEncoded(conn net.Conn, cmd uint32, payload, key []byte) error {
	pkt, err := Encode6699(1, cmd, payload, key)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if _, err := conn.Write(pkt); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func TestDeviceStatus_Success(t *testing.T) {
	localKey := []byte("0123456789abcdef")
	fixture := map[string]any{
		"1":  float64(9150),
		"2":  float64(263),
		"7":  float64(100),
		"10": float64(800),
	}

	addr, stop := mockTuyaDevice(t, localKey, fixture)
	defer stop()

	host, port, _ := net.SplitHostPort(addr)
	var portInt int
	fmt.Sscanf(port, "%d", &portInt)

	dev := NewDevice("test-device-id", host, portInt, string(localKey))
	dev.Timeout = 5 * time.Second

	dps, err := dev.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if dps["1"] != float64(9150) {
		t.Errorf("dps[1] = %v, want 9150", dps["1"])
	}
	if dps["2"] != float64(263) {
		t.Errorf("dps[2] = %v, want 263", dps["2"])
	}
	if dps["10"] != float64(800) {
		t.Errorf("dps[10] = %v, want 800", dps["10"])
	}
}

func TestDeviceStatus_ConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	var portInt int
	fmt.Sscanf(port, "%d", &portInt)

	dev := NewDevice("test", host, portInt, "0123456789abcdef")
	dev.Timeout = 500 * time.Millisecond

	_, err = dev.Status(context.Background())
	if err == nil {
		t.Error("expected error for closed listener")
	}
}
