package tuya

import "fmt"

type ProtocolError struct {
	Op  string
	Msg string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("tuya protocol error [%s]: %s", e.Op, e.Msg)
}

type CryptoError struct {
	Op  string
	Err error
}

func (e *CryptoError) Error() string {
	return fmt.Sprintf("tuya crypto error [%s]: %v", e.Op, e.Err)
}

func (e *CryptoError) Unwrap() error {
	return e.Err
}

type DeviceError struct {
	IP      string
	RetCode uint32
	Msg     string
}

func (e *DeviceError) Error() string {
	if e.RetCode != 0 {
		return fmt.Sprintf("tuya device %s error (retcode %d): %s", e.IP, e.RetCode, e.Msg)
	}
	return fmt.Sprintf("tuya device %s error: %s", e.IP, e.Msg)
}
