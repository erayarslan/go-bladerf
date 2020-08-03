package io_module

// #include <libbladeRF.h>
import "C"

type IOModule int

const (
	IOTX IOModule = C.BLADERF_MODULE_TX
	IORX IOModule = C.BLADERF_MODULE_RX
)
