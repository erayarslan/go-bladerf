package error

// #include <libbladeRF.h>
import "C"

type Error int

const (
	Unexpected  Error = -1
	Range       Error = -2
	Inval       Error = -3
	Mem         Error = -4
	Io          Error = -5
	Timeout     Error = -6
	Nodev       Error = -7
	Unsupported Error = -8
	Misaligned  Error = -9
	Checksum    Error = -10
	NoFile      Error = -11
	UpdateFpga  Error = -12
	UpdateFw    Error = -13
	TimePast    Error = -14
	QueueFull   Error = -15
	FpgaOp      Error = -16
	Permission  Error = -17
	WouldBlock  Error = -18
	NotInit     Error = -19
)

func (err Error) Error() string {
	return err.String()
}

func (err Error) String() string {
	switch err {
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
