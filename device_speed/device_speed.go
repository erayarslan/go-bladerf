package device_speed

// #include <libbladeRF.h>
import "C"

type DeviceSpeed int

const (
	SpeedUnknown DeviceSpeed = C.BLADERF_DEVICE_SPEED_UNKNOWN
	SpeedHigh    DeviceSpeed = C.BLADERF_DEVICE_SPEED_HIGH
	SpeedSuper   DeviceSpeed = C.BLADERF_DEVICE_SPEED_SUPER
)
