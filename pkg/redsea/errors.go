package redsea

import "fmt"

type DeviceUnreachableError struct {
	IP  string
	Err error
}

func (e *DeviceUnreachableError) Error() string {
	return fmt.Sprintf("device at %s unreachable: %v", e.IP, e.Err)
}

func (e *DeviceUnreachableError) Unwrap() error {
	return e.Err
}

type RequestFailedError struct {
	Method     string
	Path       string
	StatusCode int
}

func (e *RequestFailedError) Error() string {
	return fmt.Sprintf("%s %s failed with status %d", e.Method, e.Path, e.StatusCode)
}
