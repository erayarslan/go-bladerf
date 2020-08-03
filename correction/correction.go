package channel_layout

// #include <libbladeRF.h>
import "C"

type Correction int

const (
	DCOFF_I Correction = C.BLADERF_CORR_DCOFF_I
	DCOFF_Q Correction = C.BLADERF_CORR_DCOFF_Q
	PHASE   Correction = C.BLADERF_CORR_PHASE
	GAIN    Correction = C.BLADERF_CORR_GAIN
)
