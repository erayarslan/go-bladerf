package error

// #include <libbladeRF.h>
import "C"

type Code int

const (
	Unexpected  Code = -1
	Range       Code = -2
	Inval       Code = -3
	Mem         Code = -4
	Io          Code = -5
	Timeout     Code = -6
	Nodev       Code = -7
	Unsupported Code = -8
	Misaligned  Code = -9
	Checksum    Code = -10
	NoFile      Code = -11
	UpdateFpga  Code = -12
	UpdateFw    Code = -13
	TimePast    Code = -14
	QueueFull   Code = -15
	FpgaOp      Code = -16
	Permission  Code = -17
	WouldBlock  Code = -18
	NotInit     Code = -19
)

func codeToString(code Code) string {
	switch code {
	case Unexpected:
		return "An unexpected failure occurred"
	case Range:
		return "Provided parameter is out of range"
	case Inval:
		return "Invalid operation/parameter"
	case Mem:
		return "Memory allocation error"
	case Io:
		return "File/Device I/O error"
	case Timeout:
		return "Operation timed out"
	case Nodev:
		return "No device(s) available"
	case Unsupported:
		return "Operation not supported"
	case Misaligned:
		return "Misaligned flash access"
	case Checksum:
		return "Invalid checksum"
	case NoFile:
		return "File not found"
	case UpdateFpga:
		return "An FPGA update is required"
	case UpdateFw:
		return "A firmware update is required"
	case TimePast:
		return "Requested timestamp is in the past"
	case QueueFull:
		return "Could not enqueue data into full queue"
	case FpgaOp:
		return "An FPGA operation reported failure"
	case Permission:
		return "Insufficient permissions for the requested operation"
	case WouldBlock:
		return "Operation would block, but has been requested to be non-blocking. This indicates to a caller that it may need to retry the operation later."
	case NotInit:
		return "Device insufficiently initialized for operation"
	}

	return "InvalidError"
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func New(code int) error {
	if code == 0 {
		return nil
	}

	return &errorString{s: codeToString(Code(code))}
}
