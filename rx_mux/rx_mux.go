package rx_mux

// #include <libbladeRF.h>
import "C"

type RXMux int

const (
	Invalid          RXMux = C.BLADERF_RX_MUX_INVALID
	Baseband         RXMux = C.BLADERF_RX_MUX_BASEBAND
	Counter_12bit    RXMux = C.BLADERF_RX_MUX_12BIT_COUNTER
	Counter_32bit    RXMux = C.BLADERF_RX_MUX_32BIT_COUNTER
	Digital_Loopback RXMux = C.BLADERF_RX_MUX_DIGITAL_LOOPBACK
)
