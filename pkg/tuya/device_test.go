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

		// Step 1: Read SESS_KEY_NEG_START (6699, encrypted with localKey)
		pktData, err := ReadPacket(conn)
		if err != nil {
			t.Errorf("mock: read start: %v", err)
			return
		}
		cmd, clientNonce, err := Decode6699(pktData, localKey)
		if err != nil || cmd != CmdSessKeyNegStart {
			t.Errorf("mock: decode start: cmd=0x%02X err=%v", cmd, err)
			return
		}

		// Step 2: Send SESS_KEY_NEG_RESP with deviceNonce + HMAC, encrypted with localKey
		deviceNonce, _ := GenerateNonce()
		hmacVal := ComputeHMAC(localKey, clientNonce)
		respPayload := append(deviceNonce[:], hmacVal...)

		respPkt, err := Encode6699(1, CmdSessKeyNegResp, respPayload, localKey)
		if err != nil {
			t.Errorf("mock: encode resp: %v", err)
			return
		}
		if _, err := conn.Write(respPkt); err != nil {
			t.Errorf("mock: write resp: %v", err)
			return
		}

		// Step 3: Read SESS_KEY_NEG_FINISH (6699, encrypted with localKey)
		pktData, err = ReadPacket(conn)
		if err != nil {
			t.Errorf("mock: read finish: %v", err)
			return
		}
		cmd, _, err = Decode6699(pktData, localKey)
		if err != nil || cmd != CmdSessKeyNegFinish {
			t.Errorf("mock: decode finish: cmd=0x%02X err=%v", cmd, err)
			return
		}

		// Derive session key
		sessionKey, err := DeriveSessionKey(localKey, clientNonce, deviceNonce[:])
		if err != nil {
			t.Errorf("mock: derive key: %v", err)
			return
		}

		// Step 4: Read DP_QUERY_NEW (6699, encrypted with sessionKey)
		pktData, err = ReadPacket(conn)
		if err != nil {
			t.Errorf("mock: read query: %v", err)
			return
		}
		cmd, _, err = Decode6699(pktData, sessionKey)
		if err != nil || cmd != CmdDPQueryNew {
			t.Errorf("mock: decode query: cmd=0x%02X err=%v", cmd, err)
			return
		}

		// Step 5: Send DPS response (6699, encrypted with sessionKey)
		response := map[string]any{
			"dps": dpsFixture,
			"t":   time.Now().Unix(),
		}
		respJSON, _ := json.Marshal(response)

		dataPkt, err := Encode6699(1, CmdDPQueryNew, respJSON, sessionKey)
		if err != nil {
			t.Errorf("mock: encode data resp: %v", err)
			return
		}
		conn.Write(dataPkt)
	}()

	return ln.Addr().String(), func() {
		ln.Close()
		<-done
	}
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
