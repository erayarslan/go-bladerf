package clock_select

// #include <libbladeRF.h>
import "C"

type ClockSelect int

const (
	ClockSelectUnknown  ClockSelect = -99
	ClockSelectVCTCXO   ClockSelect = C.CLOCK_SELECT_ONBOARD
	ClockSelectExternal ClockSelect = C.CLOCK_SELECT_EXTERNAL
)
