package eheim

import "fmt"

// ConnectionError indicates failure to connect to the Eheim hub.
type ConnectionError struct {
	Host string
	Err  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("cannot connect to Eheim hub at %s: %v", e.Host, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// ProtocolError indicates an unexpected message or protocol violation.
type ProtocolError struct {
	Op  string
	Msg string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("eheim protocol error during %s: %s", e.Op, e.Msg)
}

// DeviceNotFoundError indicates the target device was not found on the mesh.
type DeviceNotFoundError struct {
	MAC string
}

func (e *DeviceNotFoundError) Error() string {
	if e.MAC != "" {
		return fmt.Sprintf("device %s not found on Eheim mesh", e.MAC)
	}
	return "no feeder found on Eheim mesh"
}

// RequestFailedError indicates an HTTP request returned an error status.
type RequestFailedError struct {
	Method     string
	Path       string
	StatusCode int
}

func (e *RequestFailedError) Error() string {
	return fmt.Sprintf("%s %s failed with status %d", e.Method, e.Path, e.StatusCode)
}

// MultipleDevicesError indicates multiple feeders exist and no MAC was specified.
type MultipleDevicesError struct {
	MACs []string
}

func (e *MultipleDevicesError) Error() string {
	return fmt.Sprintf("multiple feeders found on mesh, specify --mac: %v", e.MACs)
}
