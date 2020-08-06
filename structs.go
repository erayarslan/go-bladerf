package bladerf

// #include <libbladeRF.h>
import "C"

import (
	"bladerf/backend"
	"bladerf/gain_mode"
)

type DevInfo struct {
	ref          *C.struct_bladerf_devinfo
	backend      backend.Backend
	serial       string
	usbBus       int8
	usbAddr      int8
	instance     uint
	manufacturer string
	product      string
}

func NewDevInfo(ref *C.struct_bladerf_devinfo) DevInfo {
	devInfo := DevInfo{ref: ref}

	var serial []rune
	var manufacturer []rune
	var product []rune

	for i := range devInfo.ref.serial {
		if devInfo.ref.serial[i] != 0 {
			serial = append(serial, rune(devInfo.ref.serial[i]))
		}
	}

	for i := range devInfo.ref.manufacturer {
		if devInfo.ref.manufacturer[i] != 0 {
			manufacturer = append(manufacturer, rune(devInfo.ref.manufacturer[i]))
		}
	}

	for i := range devInfo.ref.product {
		if devInfo.ref.product[i] != 0 {
			product = append(product, rune(devInfo.ref.product[i]))
		}
	}

	devInfo.backend = backend.Backend(devInfo.ref.backend)
	devInfo.serial = string(serial)
	devInfo.usbBus = int8(devInfo.ref.usb_bus)
	devInfo.usbAddr = int8(devInfo.ref.usb_addr)
	devInfo.instance = uint(devInfo.ref.instance)
	devInfo.manufacturer = string(manufacturer)
	devInfo.product = string(product)

	return devInfo
}

type Version struct {
	ref      *C.struct_bladerf_version
	major    uint16
	minor    uint16
	patch    uint16
	describe string
}

type Range struct {
	ref   *C.struct_bladerf_range
	min   int64
	max   int64
	step  int64
	scale float64
}

type BladeRF struct {
	ref *C.struct_bladerf
}

type Module struct {
	ref *C.struct_bladerf_module
}

type Stream struct {
	ref *C.struct_bladerf_stream
}

type GainModes struct {
	ref  *C.struct_bladerf_gain_modes
	name string
	mode gain_mode.GainMode
}

func NewGainModes(ref *C.struct_bladerf_gain_modes) GainModes {
	gainModes := GainModes{ref: ref}

	gainModes.name = C.GoString(gainModes.ref.name)
	gainModes.mode = gain_mode.GainMode(gainModes.ref.mode)

	return gainModes
}

type UserData struct {
	callback   func(data []int16)
	results    []int16
	bufferSize int
}

func NewUserData(callback func(data []int16), bufferSize int) UserData {
	return UserData{callback: callback, results: make([]int16, bufferSize), bufferSize: bufferSize}
}
