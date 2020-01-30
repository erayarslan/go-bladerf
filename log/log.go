package log

// #cgo LDFLAGS: -lbladeRF
// #include <libbladeRF.h>
import "C"
import "errors"

type Level int

const (
	Verbose  Level = 0
	Debug    Level = 1
	Info     Level = 2
	Warning  Level = 3
	Error    Level = 4
	Critical Level = 5
	Silent   Level = 6
)

func SetVerbosity(level Level) {
	switch level {
	case Verbose:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_VERBOSE)
	case Debug:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_DEBUG)
	case Info:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_INFO)
	case Warning:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_WARNING)
	case Error:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_ERROR)
	case Critical:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_CRITICAL)
	case Silent:
		C.bladerf_log_set_verbosity(C.BLADERF_LOG_LEVEL_VERBOSE)
	default:
		panic(errors.New("invalid libbladerf_verbosity"))
	}
}
