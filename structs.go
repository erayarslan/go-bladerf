package bladerf

// #include <libbladeRF.h>
import "C"
import "unsafe"

type Timestamp uint64

type DeviceInfo struct {
	raw          C.struct_bladerf_devinfo
	hasRaw       bool
	listRef      *C.struct_bladerf_devinfo
	Backend      Backend
	Serial       string
	UsbBus       int8
	UsbAddr      int8
	Instance     uint
	Manufacturer string
	Product      string
	Addr         string
}

func NewDeviceInfo(ref *C.struct_bladerf_devinfo) DeviceInfo {
	if ref == nil {
		return DeviceInfo{}
	}

	return DeviceInfo{
		raw:          *ref,
		hasRaw:       true,
		Backend:      Backend(ref.backend),
		Serial:       C.GoString(&ref.serial[0]),
		UsbBus:       int8(ref.usb_bus),
		UsbAddr:      int8(ref.usb_addr),
		Instance:     uint(ref.instance),
		Manufacturer: C.GoString(&ref.manufacturer[0]),
		Product:      C.GoString(&ref.product[0]),
		Addr:         C.GoString(&ref.addr[0]),
	}
}

func (d *DeviceInfo) cStruct() C.struct_bladerf_devinfo {
	var out C.struct_bladerf_devinfo
	if d == nil {
		return out
	}
	if d.hasRaw {
		out = d.raw
	}

	out.backend = C.bladerf_backend(d.Backend)
	copyGoStringToCChars(out.serial[:], d.Serial)
	out.usb_bus = C.uint8_t(uint8(d.UsbBus))
	out.usb_addr = C.uint8_t(uint8(d.UsbAddr))
	out.instance = C.uint(d.Instance)
	copyGoStringToCChars(out.manufacturer[:], d.Manufacturer)
	copyGoStringToCChars(out.product[:], d.Product)
	copyGoStringToCChars(out.addr[:], d.Addr)
	return out
}

func copyGoStringToCChars(dst []C.char, s string) {
	if len(dst) == 0 {
		return
	}

	limit := len(dst) - 1
	if len(s) < limit {
		limit = len(s)
	}

	for i := 0; i < limit; i++ {
		dst[i] = C.char(s[i])
	}
	dst[limit] = 0
}

type Version struct {
	Major    uint16
	Minor    uint16
	Patch    uint16
	Describe string
}

func NewVersion(ref *C.struct_bladerf_version) Version {
	if ref == nil {
		return Version{}
	}

	return Version{
		Major:    uint16(ref.major),
		Minor:    uint16(ref.minor),
		Patch:    uint16(ref.patch),
		Describe: C.GoString(ref.describe),
	}
}

type RationalRate struct {
	Integer uint64
	Num     uint64
	Den     uint64
}

func NewRationalRate(ref *C.struct_bladerf_rational_rate) RationalRate {
	if ref == nil {
		return RationalRate{}
	}

	return RationalRate{
		Integer: uint64(ref.integer),
		Num:     uint64(ref.num),
		Den:     uint64(ref.den),
	}
}

func (r RationalRate) toC() C.struct_bladerf_rational_rate {
	return C.struct_bladerf_rational_rate{
		integer: C.uint64_t(r.Integer),
		num:     C.uint64_t(r.Num),
		den:     C.uint64_t(r.Den),
	}
}

type Range struct {
	Min   int64
	Max   int64
	Step  int64
	Scale float64
}

func NewRange(ref *C.struct_bladerf_range) Range {
	if ref == nil {
		return Range{}
	}

	return Range{
		Min:   int64(ref.min),
		Max:   int64(ref.max),
		Step:  int64(ref.step),
		Scale: float64(ref.scale),
	}
}

type BladeRF struct {
	ref *C.struct_bladerf
}

type QuickTune struct {
	ref *C.struct_bladerf_quick_tune
}

type Serial struct {
	Serial string
}

func NewSerial(ref *C.struct_bladerf_serial) Serial {
	if ref == nil {
		return Serial{}
	}

	return Serial{Serial: C.GoString(&ref.serial[0])}
}

type Stream struct {
	ref      *C.struct_bladerf_stream
	userData unsafe.Pointer
}

type Trigger struct {
	ref *C.struct_bladerf_trigger
}

type LoopbackModes struct {
	Name string
	Mode Loopback
}

func NewLoopbackModes(ref *C.struct_bladerf_loopback_modes) LoopbackModes {
	if ref == nil {
		return LoopbackModes{}
	}

	return LoopbackModes{
		Name: C.GoString(ref.name),
		Mode: Loopback(ref.mode),
	}
}

type GainModes struct {
	Name string
	Mode GainMode
}

func NewGainModes(ref *C.struct_bladerf_gain_modes) GainModes {
	if ref == nil {
		return GainModes{}
	}

	return GainModes{
		Name: C.GoString(ref.name),
		Mode: GainMode(ref.mode),
	}
}

type Metadata struct {
	Timestamp   Timestamp
	Flags       uint32
	Status      uint32
	ActualCount uint
}

func LoadMetadata(ref *C.struct_bladerf_metadata) Metadata {
	if ref == nil {
		return Metadata{}
	}

	return Metadata{
		Timestamp:   Timestamp(ref.timestamp),
		Flags:       uint32(ref.flags),
		Status:      uint32(ref.status),
		ActualCount: uint(ref.actual_count),
	}
}

func NewMetadata(timestamp Timestamp, flags uint32) Metadata {
	return Metadata{
		Timestamp: timestamp,
		Flags:     flags,
	}
}

func (m Metadata) toC() C.struct_bladerf_metadata {
	return C.struct_bladerf_metadata{
		timestamp:    C.bladerf_timestamp(m.Timestamp),
		flags:        C.uint32_t(m.Flags),
		status:       C.uint32_t(m.Status),
		actual_count: C.uint(m.ActualCount),
	}
}

type UserData struct {
	callback   func(data []int16) GoStream
	results    []int16
	bufferSize int
}

func NewUserData(callback func(data []int16) GoStream, bufferSize int) UserData {
	return UserData{
		callback:   callback,
		results:    make([]int16, bufferSize),
		bufferSize: bufferSize,
	}
}
