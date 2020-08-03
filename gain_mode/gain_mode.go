package gain_mode

// #include <libbladeRF.h>
import "C"

type GainMode int

const (
	Default        GainMode = C.BLADERF_GAIN_DEFAULT
	Manual         GainMode = C.BLADERF_GAIN_MGC
	FastAttack_AGC GainMode = C.BLADERF_GAIN_FASTATTACK_AGC
	SlowAttack_AGC GainMode = C.BLADERF_GAIN_SLOWATTACK_AGC
	Hybrid_AGC     GainMode = C.BLADERF_GAIN_HYBRID_AGC
)
