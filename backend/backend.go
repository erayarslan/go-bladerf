package backend

// #include <libbladeRF.h>
import "C"

type Backend int

const (
	Any     Backend = C.BLADERF_BACKEND_ANY
	Linux   Backend = C.BLADERF_BACKEND_LINUX
	LibUSB  Backend = C.BLADERF_BACKEND_LIBUSB
	Cypress Backend = C.BLADERF_BACKEND_CYPRESS
	Dummy   Backend = C.BLADERF_BACKEND_DUMMY
)
