package bladerf

// #include <libbladeRF.h>
import "C"

type Timestamp uint64

type DeviceInfo struct {
	ref          *C.struct_bladerf_devinfo
	Backend      Backend
	Serial       string
	UsbBus       int8
	UsbAddr      int8
	Instance     uint
	Manufacturer string
	Product      string
}

func NewDeviceInfo(ref *C.struct_bladerf_devinfo) DeviceInfo {
	deviceInfo := DeviceInfo{ref: ref}

	var serial []rune
	var manufacturer []rune
	var product []rune

	for i := range deviceInfo.ref.serial {
		if deviceInfo.ref.serial[i] != 0 {
			serial = append(serial, rune(deviceInfo.ref.serial[i]))
		}
	}

	for i := range deviceInfo.ref.manufacturer {
		if deviceInfo.ref.manufacturer[i] != 0 {
			manufacturer = append(manufacturer, rune(deviceInfo.ref.manufacturer[i]))
		}
	}

	for i := range deviceInfo.ref.product {
		if deviceInfo.ref.product[i] != 0 {
			product = append(product, rune(deviceInfo.ref.product[i]))
		}
	}

	deviceInfo.Backend = Backend(deviceInfo.ref.backend)
	deviceInfo.Serial = string(serial)
	deviceInfo.UsbBus = int8(deviceInfo.ref.usb_bus)
	deviceInfo.UsbAddr = int8(deviceInfo.ref.usb_addr)
	deviceInfo.Instance = uint(deviceInfo.ref.instance)
	deviceInfo.Manufacturer = string(manufacturer)
	deviceInfo.Product = string(product)

	return deviceInfo
}

type Version struct {
	ref      *C.struct_bladerf_version
	Major    uint16
	Minor    uint16
	Patch    uint16
	Describe string
}

func NewVersion(ref *C.struct_bladerf_version) Version {
	version := Version{ref: ref}

	version.Major = uint16((*ref).major)
	version.Minor = uint16((*ref).minor)
	version.Patch = uint16((*ref).patch)
	version.Describe = C.GoString((*ref).describe)

	return version
}

type RationalRate struct {
	ref     *C.struct_bladerf_rational_rate
	Integer uint64
	Num     uint64
	Den     uint64
}

func NewRationalRate(ref *C.struct_bladerf_rational_rate) RationalRate {
	return RationalRate{ref: ref, Integer: uint64((*ref).integer), Num: uint64((*ref).num), Den: uint64((*ref).den)}
}

type Range struct {
	ref   *C.struct_bladerf_range
	Min   int64
	Max   int64
	Step  int64
	Scale float64
}

func NewRange(ref *C.struct_bladerf_range) Range {
	return Range{ref: ref, Min: int64((*ref).min), Max: int64((*ref).max), Step: int64((*ref).step), Scale: float64((*ref).scale)}
}

type BladeRF struct {
	ref *C.struct_bladerf
}

type QuickTune struct {
	ref *C.struct_bladerf_quick_tune
}

type Serial struct {
	ref    *C.struct_bladerf_serial
	Serial string
}

func NewSerial(ref *C.struct_bladerf_serial) Serial {
	var serial []rune
	for i := range (*ref).serial {
		if (*ref).serial[i] != 0 {
			serial = append(serial, rune((*ref).serial[i]))
		}
	}

	return Serial{ref: ref, Serial: string(serial)}
}

type Stream struct {
	ref *C.struct_bladerf_stream
}

type Trigger struct {
	ref *C.struct_bladerf_trigger
}

type LoopbackModes struct {
	ref  *C.struct_bladerf_loopback_modes
	Name string
	Mode Loopback
}

func NewLoopbackModes(ref *C.struct_bladerf_loopback_modes) LoopbackModes {
	loopbackModes := LoopbackModes{ref: ref}

	loopbackModes.Name = C.GoString(loopbackModes.ref.name)
	loopbackModes.Mode = Loopback(loopbackModes.ref.mode)

	return loopbackModes
}

type GainModes struct {
	ref  *C.struct_bladerf_gain_modes
	Name string
	Mode GainMode
}

func NewGainModes(ref *C.struct_bladerf_gain_modes) GainModes {
	gainModes := GainModes{ref: ref}

	gainModes.Name = C.GoString(gainModes.ref.name)
	gainModes.Mode = GainMode(gainModes.ref.mode)

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
