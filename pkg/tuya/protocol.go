package tuya

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	Prefix55AA uint32 = 0x000055AA
	Footer55AA uint32 = 0x0000AA55
	Prefix6699 uint32 = 0x00006699
	Footer6699 uint32 = 0x00009966

	CmdSessKeyNegStart  uint32 = 0x03
	CmdSessKeyNegResp   uint32 = 0x04
	CmdSessKeyNegFinish uint32 = 0x05
	CmdHeartbeat        uint32 = 0x09
	CmdDPQuery          uint32 = 0x0A

	gcmTagSize  = 16
	gcmNonceLen = 12
)

// Encode55AA builds a 55AA-format packet for client→device messages.
// Layout: [prefix:4][seq:4][cmd:4][length:4][payload:N][hmac:32][footer:4]
func Encode55AA(seq, cmd uint32, payload, hmacKey []byte) []byte {
	// length = payload + hmac(32) + footer(4)
	length := uint32(len(payload) + 32 + 4)

	buf := make([]byte, 0, 16+len(payload)+32+4)
	buf = binary.BigEndian.AppendUint32(buf, Prefix55AA)
	buf = binary.BigEndian.AppendUint32(buf, seq)
	buf = binary.BigEndian.AppendUint32(buf, cmd)
	buf = binary.BigEndian.AppendUint32(buf, length)
	buf = append(buf, payload...)

	// HMAC-SHA256 over everything before the HMAC field
	mac := ComputeHMAC(hmacKey, buf)
	buf = append(buf, mac...)
	buf = binary.BigEndian.AppendUint32(buf, Footer55AA)

	return buf
}

// Decode55AA parses a device→client 55AA packet.
// Device responses include a 4-byte retcode between length and payload.
// Returns seq, cmd, retcode, payload (after retcode, before HMAC).
func Decode55AA(data []byte) (seq, cmd, retcode uint32, payload []byte, err error) {
	if len(data) < 24 { // 16 header + 4 retcode + 4 footer minimum
		return 0, 0, 0, nil, &ProtocolError{Op: "decode_55aa", Msg: fmt.Sprintf("packet too short: %d bytes", len(data))}
	}

	prefix := binary.BigEndian.Uint32(data[0:4])
	if prefix != Prefix55AA {
		return 0, 0, 0, nil, &ProtocolError{Op: "decode_55aa", Msg: fmt.Sprintf("bad prefix: 0x%08X", prefix)}
	}

	seq = binary.BigEndian.Uint32(data[4:8])
	cmd = binary.BigEndian.Uint32(data[8:12])
	_ = binary.BigEndian.Uint32(data[12:16]) // length

	footer := binary.BigEndian.Uint32(data[len(data)-4:])
	if footer != Footer55AA {
		return 0, 0, 0, nil, &ProtocolError{Op: "decode_55aa", Msg: fmt.Sprintf("bad footer: 0x%08X", footer)}
	}

	// After header(16), before footer(4) and HMAC(32): retcode(4) + payload
	inner := data[16 : len(data)-4-32]
	if len(inner) < 4 {
		return seq, cmd, 0, nil, nil
	}

	retcode = binary.BigEndian.Uint32(inner[0:4])
	payload = inner[4:]
	return seq, cmd, retcode, payload, nil
}

// Encode6699 builds a 6699-format encrypted packet.
// Layout: [prefix:4][reserved:2][seq:4][cmd:4][length:4][iv:12][ciphertext:N][tag:16][footer:4]
func Encode6699(seq, cmd uint32, plaintext, sessionKey []byte) ([]byte, error) {
	var iv [gcmNonceLen]byte
	if _, err := rand.Read(iv[:]); err != nil {
		return nil, fmt.Errorf("generating IV: %w", err)
	}

	// Build header first to compute AAD
	ctLen := len(plaintext) // GCM ciphertext is same length as plaintext
	length := uint32(gcmNonceLen + ctLen + gcmTagSize)

	header := make([]byte, 18)
	binary.BigEndian.PutUint32(header[0:4], Prefix6699)
	// reserved = 0 at header[4:6]
	binary.BigEndian.PutUint32(header[6:10], seq)
	binary.BigEndian.PutUint32(header[10:14], cmd)
	binary.BigEndian.PutUint32(header[14:18], length)

	// AAD = header[4:18] (reserved + seq + cmd + length)
	aad := header[4:18]

	sealed, err := GCMEncrypt(sessionKey, iv[:], plaintext, aad)
	if err != nil {
		return nil, err
	}

	// sealed = ciphertext + tag (concatenated by Go's GCM)
	ciphertext := sealed[:ctLen]
	tag := sealed[ctLen:]

	buf := make([]byte, 0, 18+gcmNonceLen+ctLen+gcmTagSize+4)
	buf = append(buf, header...)
	buf = append(buf, iv[:]...)
	buf = append(buf, ciphertext...)
	buf = append(buf, tag...)
	buf = binary.BigEndian.AppendUint32(buf, Footer6699)

	return buf, nil
}

// Decode6699 parses and decrypts a 6699-format packet.
// Returns the command and decrypted payload.
func Decode6699(data, sessionKey []byte) (cmd uint32, payload []byte, err error) {
	if len(data) < 18+gcmNonceLen+gcmTagSize+4 {
		return 0, nil, &ProtocolError{Op: "decode_6699", Msg: fmt.Sprintf("packet too short: %d bytes", len(data))}
	}

	prefix := binary.BigEndian.Uint32(data[0:4])
	if prefix != Prefix6699 {
		return 0, nil, &ProtocolError{Op: "decode_6699", Msg: fmt.Sprintf("bad prefix: 0x%08X", prefix)}
	}

	cmd = binary.BigEndian.Uint32(data[10:14])
	length := binary.BigEndian.Uint32(data[14:18])

	footer := binary.BigEndian.Uint32(data[len(data)-4:])
	if footer != Footer6699 {
		return 0, nil, &ProtocolError{Op: "decode_6699", Msg: fmt.Sprintf("bad footer: 0x%08X", footer)}
	}

	// Verify we have enough data: header(18) + length + footer(4)
	if uint32(len(data)) < 18+length+4 {
		return 0, nil, &ProtocolError{Op: "decode_6699", Msg: "length field exceeds packet size"}
	}

	// AAD = data[4:18]
	aad := data[4:18]

	// After header: iv(12), ciphertext(length-12-16), tag(16)
	encStart := uint32(18)
	iv := data[encStart : encStart+gcmNonceLen]
	ctAndTag := data[encStart+gcmNonceLen : encStart+length]

	payload, err = GCMDecrypt(sessionKey, iv, ctAndTag, aad)
	if err != nil {
		return cmd, nil, err
	}

	return cmd, payload, nil
}

// ReadPacket reads one complete Tuya packet from a TCP stream.
func ReadPacket(r io.Reader) ([]byte, error) {
	// Read the first 4 bytes to identify the format
	prefixBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, prefixBuf); err != nil {
		return nil, fmt.Errorf("reading prefix: %w", err)
	}

	prefix := binary.BigEndian.Uint32(prefixBuf)

	switch prefix {
	case Prefix55AA:
		return readPacket55AA(r, prefixBuf)
	case Prefix6699:
		return readPacket6699(r, prefixBuf)
	default:
		return nil, &ProtocolError{Op: "read_packet", Msg: fmt.Sprintf("unknown prefix: 0x%08X", prefix)}
	}
}

func readPacket55AA(r io.Reader, prefixBuf []byte) ([]byte, error) {
	// Read rest of header: seq(4) + cmd(4) + length(4) = 12 bytes
	headerRest := make([]byte, 12)
	if _, err := io.ReadFull(r, headerRest); err != nil {
		return nil, fmt.Errorf("reading 55AA header: %w", err)
	}

	length := binary.BigEndian.Uint32(headerRest[8:12])
	if length > 65536 {
		return nil, &ProtocolError{Op: "read_55aa", Msg: fmt.Sprintf("length too large: %d", length)}
	}

	// Read the remaining `length` bytes (includes retcode/payload + hmac + footer)
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("reading 55AA body: %w", err)
	}

	// Assemble full packet
	pkt := make([]byte, 0, 16+length)
	pkt = append(pkt, prefixBuf...)
	pkt = append(pkt, headerRest...)
	pkt = append(pkt, body...)
	return pkt, nil
}

func readPacket6699(r io.Reader, prefixBuf []byte) ([]byte, error) {
	// Read rest of header: reserved(2) + seq(4) + cmd(4) + length(4) = 14 bytes
	headerRest := make([]byte, 14)
	if _, err := io.ReadFull(r, headerRest); err != nil {
		return nil, fmt.Errorf("reading 6699 header: %w", err)
	}

	length := binary.BigEndian.Uint32(headerRest[10:14])
	if length > 65536 {
		return nil, &ProtocolError{Op: "read_6699", Msg: fmt.Sprintf("length too large: %d", length)}
	}

	// Read: encrypted data (length bytes) + footer (4 bytes)
	body := make([]byte, length+4)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("reading 6699 body: %w", err)
	}

	pkt := make([]byte, 0, 18+length+4)
	pkt = append(pkt, prefixBuf...)
	pkt = append(pkt, headerRest...)
	pkt = append(pkt, body...)
	return pkt, nil
}
