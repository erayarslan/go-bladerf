package device_speed

// #include <libbladeRF.h>
import "C"

type DeviceSpeed int

const (
	Unknown DeviceSpeed = C.BLADERF_DEVICE_SPEED_UNKNOWN
	High    DeviceSpeed = C.BLADERF_DEVICE_SPEED_HIGH
	Super   DeviceSpeed = C.BLADERF_DEVICE_SPEED_SUPER
)
