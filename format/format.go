package format

// #include <libbladeRF.h>
import "C"

type Format int

const (
	SC16_Q11      Format = C.BLADERF_FORMAT_SC16_Q11
	SC16_Q11_META Format = C.BLADERF_FORMAT_SC16_Q11_META
)
