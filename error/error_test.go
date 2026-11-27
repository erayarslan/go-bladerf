package error

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	if err := New(0); err != nil {
		t.Fatalf("expected nil for code 0, got %v", err)
	}

	tests := []struct {
		code Code
		want string
	}{
		{Unexpected, "An unexpected failure occurred"},
		{Range, "Provided parameter is out of range"},
		{Inval, "Invalid operation/parameter"},
		{Mem, "Memory allocation error"},
		{Io, "File/Device I/O error"},
		{Timeout, "Operation timed out"},
		{Nodev, "No device(s) available"},
		{Unsupported, "Operation not supported"},
		{Misaligned, "Misaligned flash access"},
		{Checksum, "Invalid checksum"},
		{NoFile, "File not found"},
		{UpdateFpga, "An FPGA update is required"},
		{UpdateFw, "A firmware update is required"},
		{TimePast, "Requested timestamp is in the past"},
		{QueueFull, "Could not enqueue data into full queue"},
		{FpgaOp, "An FPGA operation reported failure"},
		{Permission, "Insufficient permissions for the requested operation"},
		{WouldBlock, "Operation would block, but has been requested to be non-blocking. This indicates to a caller that it may need to retry the operation later."},
		{NotInit, "Device insufficiently initialized for operation"},
		{Code(-999), "InvalidError"},
	}

	for _, tc := range tests {
		err := New(int(tc.code))
		if err == nil {
			t.Fatalf("expected non-nil error for code %d", tc.code)
		}
		if err.Error() != tc.want {
			t.Errorf("code %d: got %q, want %q", tc.code, err.Error(), tc.want)
		}
		if tc.code.String() != tc.want {
			t.Errorf("code %d String(): got %q, want %q", tc.code, tc.code.String(), tc.want)
		}
		if !errors.Is(err, tc.code) {
			t.Errorf("errors.Is(%v, %v) = false, want true", err, tc.code)
		}
	}

	if errors.Is(New(int(Nodev)), Inval) {
		t.Error("errors.Is(Nodev, Inval) = true, want false")
	}
}
