package direction

// #include <libbladeRF.h>
import "C"

type Direction int

const (
	TX Direction = C.BLADERF_TX
	RX Direction = C.BLADERF_RX
)
