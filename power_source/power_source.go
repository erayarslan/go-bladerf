package power_source

// #include <libbladeRF.h>
import "C"

type PowerSource int

const (
	PowerSourceUnknown   PowerSource = C.BLADERF_UNKNOWN
	PowerSourceDC_Barrel PowerSource = C.BLADERF_PS_DC
	PowerSourceUSB_VBUS  PowerSource = C.BLADERF_PS_USB_VBUS
)
