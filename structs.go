package bladerf

// #include <libbladeRF.h>
import "C"

import "bladerf/gain_mode"

type Version struct {
	version  *C.struct_bladerf_version
	major    int
	minor    int
	patch    int
	describe string
}

type DevInfo struct {
	devInfo *C.struct_bladerf_devinfo
	serial  string
}

type Range struct {
	bfRange *C.struct_bladerf_range
	min     int64
	max     int64
	step    int64
	scale   float64
}

type BladeRF struct {
	bladeRF *C.struct_bladerf
}

type Module struct {
	module *C.struct_bladerf_module
}

type Stream struct {
	stream *C.struct_bladerf_stream
}

type GainModes struct {
	gainModes *C.struct_bladerf_gain_modes
	name      string
	mode      gain_mode.GainMode
}
