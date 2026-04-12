package tuya

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncode55AA_Decode55AA_Roundtrip(t *testing.T) {
	hmacKey := []byte("0123456789abcdef")
	payload := []byte("test payload data")

	encoded := Encode55AA(1, CmdSessKeyNegStart, payload, hmacKey)

	// Verify prefix
	prefix := binary.BigEndian.Uint32(encoded[0:4])
	if prefix != Prefix55AA {
		t.Errorf("prefix = 0x%08X, want 0x%08X", prefix, Prefix55AA)
	}

	// Verify footer
	footer := binary.BigEndian.Uint32(encoded[len(encoded)-4:])
	if footer != Footer55AA {
		t.Errorf("footer = 0x%08X, want 0x%08X", footer, Footer55AA)
	}
}

func TestDecode55AA_DeviceResponse(t *testing.T) {
	// Simulate a device response with retcode
	hmacKey := []byte("0123456789abcdef")
	payload := []byte("device response data")
	retcode := uint32(0)

	// Build: prefix + seq + cmd + length + retcode + payload + hmac + footer
	var buf []byte
	buf = binary.BigEndian.AppendUint32(buf, Prefix55AA)
	buf = binary.BigEndian.AppendUint32(buf, 1)                           // seq
	buf = binary.BigEndian.AppendUint32(buf, CmdSessKeyNegResp)           // cmd
	buf = binary.BigEndian.AppendUint32(buf, uint32(4+len(payload)+32+4)) // length
	buf = binary.BigEndian.AppendUint32(buf, retcode)
	buf = append(buf, payload...)
	mac := ComputeHMAC(hmacKey, buf)
	buf = append(buf, mac...)
	buf = binary.BigEndian.AppendUint32(buf, Footer55AA)

	seq, cmd, rc, p, err := Decode55AA(buf)
	if err != nil {
		t.Fatalf("Decode55AA: %v", err)
	}
	if seq != 1 {
		t.Errorf("seq = %d, want 1", seq)
	}
	if cmd != CmdSessKeyNegResp {
		t.Errorf("cmd = 0x%02X, want 0x%02X", cmd, CmdSessKeyNegResp)
	}
	if rc != 0 {
		t.Errorf("retcode = %d, want 0", rc)
	}
	if !bytes.Equal(p, payload) {
		t.Errorf("payload = %q, want %q", p, payload)
	}
}

func TestDecode55AA_TooShort(t *testing.T) {
	_, _, _, _, err := Decode55AA([]byte{0x00, 0x00, 0x55, 0xAA})
	if err == nil {
		t.Error("expected error for too-short packet")
	}
}

func TestDecode55AA_BadPrefix(t *testing.T) {
	data := make([]byte, 60)
	binary.BigEndian.PutUint32(data[0:4], 0xDEADBEEF)
	binary.BigEndian.PutUint32(data[56:60], Footer55AA)
	_, _, _, _, err := Decode55AA(data)
	if err == nil {
		t.Error("expected error for bad prefix")
	}
}

func TestEncode6699_Decode6699_Roundtrip(t *testing.T) {
	sessionKey := []byte("0123456789abcdef")
	plaintext := []byte(`{"dps":{"1":100}}`)

	encoded, err := Encode6699(1, CmdDPQuery, plaintext, sessionKey)
	if err != nil {
		t.Fatalf("Encode6699: %v", err)
	}

	// Verify prefix and footer
	prefix := binary.BigEndian.Uint32(encoded[0:4])
	if prefix != Prefix6699 {
		t.Errorf("prefix = 0x%08X, want 0x%08X", prefix, Prefix6699)
	}
	footer := binary.BigEndian.Uint32(encoded[len(encoded)-4:])
	if footer != Footer6699 {
		t.Errorf("footer = 0x%08X, want 0x%08X", footer, Footer6699)
	}

	// Decode
	cmd, decrypted, err := Decode6699(encoded, sessionKey)
	if err != nil {
		t.Fatalf("Decode6699: %v", err)
	}
	if cmd != CmdDPQuery {
		t.Errorf("cmd = 0x%02X, want 0x%02X", cmd, CmdDPQuery)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestDecode6699_WrongKey(t *testing.T) {
	sessionKey := []byte("0123456789abcdef")
	wrongKey := []byte("fedcba9876543210")
	plaintext := []byte(`{"test":true}`)

	encoded, err := Encode6699(1, CmdDPQuery, plaintext, sessionKey)
	if err != nil {
		t.Fatalf("Encode6699: %v", err)
	}

	_, _, err = Decode6699(encoded, wrongKey)
	if err == nil {
		t.Error("expected error when decoding with wrong key")
	}
}

func TestDecode6699_TooShort(t *testing.T) {
	_, _, err := Decode6699([]byte{0x00, 0x00, 0x66, 0x99}, []byte("0123456789abcdef"))
	if err == nil {
		t.Error("expected error for too-short packet")
	}
}

func TestReadPacket_55AA(t *testing.T) {
	hmacKey := []byte("0123456789abcdef")
	payload := []byte("hello")
	encoded := Encode55AA(1, CmdSessKeyNegStart, payload, hmacKey)

	r := bytes.NewReader(encoded)
	pkt, err := ReadPacket(r)
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if !bytes.Equal(pkt, encoded) {
		t.Error("read packet does not match encoded packet")
	}
}

func TestReadPacket_6699(t *testing.T) {
	sessionKey := []byte("0123456789abcdef")
	plaintext := []byte(`{"dps":{"1":1}}`)

	encoded, err := Encode6699(1, CmdDPQuery, plaintext, sessionKey)
	if err != nil {
		t.Fatalf("Encode6699: %v", err)
	}

	r := bytes.NewReader(encoded)
	pkt, err := ReadPacket(r)
	if err != nil {
		t.Fatalf("ReadPacket: %v", err)
	}
	if !bytes.Equal(pkt, encoded) {
		t.Error("read packet does not match encoded packet")
	}
}
