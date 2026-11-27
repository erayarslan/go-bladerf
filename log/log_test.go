package log

import "testing"

func TestSetVerbosity(t *testing.T) {
	levels := []Level{
		Verbose,
		Debug,
		Info,
		Warning,
		Error,
		Critical,
		Silent,
		Info,
	}

	for _, level := range levels {
		SetVerbosity(level)
	}
}

func TestSetVerbosityInvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid log level")
		}
	}()

	SetVerbosity(Level(99))
}
