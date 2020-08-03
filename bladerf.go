package bladerf

import "C"

// #cgo darwin CFLAGS: -I/usr/local/include
// #cgo darwin LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lbladeRF
// #include <libbladeRF.h>
//
// extern void* cbGo(struct bladerf *dev, struct bladerf_stream *stream, struct bladerf_metadata *md, void* samples, size_t num_samples, void* user_data);
import "C"
import (
	"bladerf/channel_layout"
	"bladerf/direction"
	exception "bladerf/error"
	"bladerf/format"
	"bladerf/gain_mode"
	"bladerf/loopback"
	"fmt"
	"github.com/mattn/go-pointer"
	"unsafe"
)

func GetError(code C.int) error {
	return exception.New(int(code))
}

var buff *C.int16_t

const size = unsafe.Sizeof(*buff)

//export cbGo
func cbGo(dev *C.struct_bladerf,
	stream *C.struct_bladerf_stream,
	metadata *C.struct_bladerf_metadata,
	samples unsafe.Pointer,
	numSamples C.size_t,
	userData unsafe.Pointer) unsafe.Pointer {
	cb := pointer.Restore(userData).(Callback)

	for i := 0; i < cb.bufferSize; i++ {
		cb.results[i] = int16(*((*C.int16_t)(unsafe.Pointer(uintptr(samples) + (size * uintptr(i))))))
	}

	cb.cb(cb.results)

	return samples // C.malloc(C.size_t(size * cb.bufferSize * 2 * 1)) allocate always fcjinf new
}

type Channel int

const (
	RX_V int = 0x0
	TX_V int = 0x1
)

const (
	BLADERF_VCTCXO_FREQUENCY = 38.4e6
	BLADERF_REFIN_DEFAULT    = 10.0e6
	BLADERF_SERIAL_LENGTH    = 33
)

func GetVersion() Version {
	var version C.struct_bladerf_version
	C.bladerf_version(&version)
	return Version{version: &version, major: int(version.major), minor: int(version.minor), patch: int(version.patch), describe: C.GoString(version.describe)}
}

func PrintVersion(version Version) {
	fmt.Printf("v%d.%d.%d (\"%s\")", version.major, version.minor, version.patch, version.describe)
}

func LoadFpga(bladeRF BladeRF, imagePath string) {
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))

	C.bladerf_load_fpga(bladeRF.bladeRF, path)
}

func FreeDeviceList(devInfo DevInfo) {
	C.bladerf_free_device_list(devInfo.devInfo)
}

func GetDeviceList() []DevInfo {
	var devInfo *C.struct_bladerf_devinfo
	var devices []DevInfo

	count := int(C.bladerf_get_device_list(&devInfo))

	if count < 1 {
		return devices
	}

	first := DevInfo{devInfo: devInfo}

	defer FreeDeviceList(first)

	devices = append(devices, first)

	for i := 0; i < count-1; i++ {
		size := unsafe.Sizeof(*devInfo)
		devInfo = (*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(devInfo)) + size))
		devices = append(devices, DevInfo{devInfo: devInfo})
	}

	return devices
}

func GetBootloaderList() []DevInfo {
	var devInfo *C.struct_bladerf_devinfo
	var devices []DevInfo

	count := int(C.bladerf_get_bootloader_list(&devInfo))

	if count < 1 {
		return devices
	}

	first := DevInfo{devInfo: devInfo}

	defer FreeDeviceList(first)

	devices = append(devices, first)

	for i := 0; i < count-1; i++ {
		size := unsafe.Sizeof(*devInfo)
		devInfo = (*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(devInfo)) + size))
		devices = append(devices, DevInfo{devInfo: devInfo})
	}

	return devices
}

func InitDevInfo() DevInfo {
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_init_devinfo(&devInfo)
	return DevInfo{devInfo: &devInfo}
}

func GetDevInfo(bladeRF *BladeRF) DevInfo {
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_get_devinfo((*bladeRF).bladeRF, &devInfo)

	var fs []rune
	for i := range devInfo.serial {
		if devInfo.serial[i] != 0 {
			fs = append(fs, rune(devInfo.serial[i]))
		}
	}

	info := DevInfo{devInfo: &devInfo, serial: string(fs)}
	return info
}

func DevInfoMatches(a DevInfo, b DevInfo) bool {
	return bool(C.bladerf_devinfo_matches(a.devInfo, b.devInfo))
}

func DevStrMatches(devstr string, info DevInfo) bool {
	val := C.CString(devstr)
	defer C.free(unsafe.Pointer(val))
	return bool(C.bladerf_devstr_matches(val, info.devInfo))
}

func GetDevInfoFromStr(devstr string) DevInfo {
	val := C.CString(devstr)
	defer C.free(unsafe.Pointer(val))
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_get_devinfo_from_str(val, &devInfo)
	return DevInfo{devInfo: &devInfo}
}

func OpenWithDevInfo(devInfo DevInfo) BladeRF {
	var bladeRF *C.struct_bladerf
	C.bladerf_open_with_devinfo(&bladeRF, devInfo.devInfo)
	return BladeRF{bladeRF: bladeRF}
}

func OpenWithDeviceIdentifier(identify string) BladeRF {
	var bladeRF *C.struct_bladerf
	C.bladerf_open(&bladeRF, C.CString(identify))
	return BladeRF{bladeRF: bladeRF}
}

func Open() BladeRF {
	var bladeRF *C.struct_bladerf
	C.bladerf_open(&bladeRF, nil)
	return BladeRF{bladeRF: bladeRF}
}

func Close(bladeRF BladeRF) {
	C.bladerf_close(bladeRF.bladeRF)
}

func SetLoopback(bladeRF *BladeRF, loopback loopback.Loopback) {
	C.bladerf_set_loopback((*bladeRF).bladeRF, C.bladerf_loopback(loopback))
}

func SetFrequency(bladeRF *BladeRF, channel Channel, frequency int) error {
	return GetError(C.bladerf_set_frequency((*bladeRF).bladeRF, C.bladerf_channel(channel), C.ulonglong(frequency)))
}

func SetSampleRate(bladeRF *BladeRF, channel Channel, sampleRate int) error {
	var actual C.uint
	err := GetError(C.bladerf_set_sample_rate((*bladeRF).bladeRF, C.bladerf_channel(channel), C.uint(sampleRate), &actual))

	if err == nil {
		println(uint(actual))
	}

	return err
}

func GetSampleRateRange(bladeRF *BladeRF, channel Channel) (int, int, int, error) {
	var bfRange *C.struct_bladerf_range

	err := GetError(C.bladerf_get_sample_rate_range((*bladeRF).bladeRF, C.bladerf_channel(channel), &bfRange))

	if err != nil {
		return 0, 0, 0, err
	}

	return int(bfRange.min), int(bfRange.max), int(bfRange.step), nil
}

func SetBandwidth(bladeRF *BladeRF, channel Channel, bandwidth int) (int, error) {
	var actual C.bladerf_bandwidth
	return int(actual), GetError(C.bladerf_set_bandwidth((*bladeRF).bladeRF, C.bladerf_channel(channel), C.uint(bandwidth), &actual))
}

func SetGain(bladeRF *BladeRF, channel Channel, gain int) error {
	return GetError(C.bladerf_set_gain((*bladeRF).bladeRF, C.bladerf_channel(channel), C.int(gain)))
}

func GetGain(bladeRF *BladeRF, channel Channel) (int, error) {
	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain((*bladeRF).bladeRF, C.bladerf_channel(channel), &gain))
	if err == nil {
		return int(gain), nil
	}

	return int(gain), err
}

func GetGainStage(bladeRF *BladeRF, channel Channel, stage string) (int, error) {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))
	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain_stage((*bladeRF).bladeRF, C.bladerf_channel(channel), val, &gain))
	if err == nil {
		return int(gain), nil
	}

	return int(gain), err
}

func GetGainMode(bladeRF *BladeRF, channel Channel) (gain_mode.GainMode, error) {
	var mode C.bladerf_gain_mode

	err := GetError(C.bladerf_get_gain_mode((*bladeRF).bladeRF, C.bladerf_channel(channel), &mode))
	result := gain_mode.GainMode(int(mode))
	if err == nil {
		return result, nil
	}

	return result, err
}

func SetGainStage(bladeRF *BladeRF, channel Channel, stage string, gain int) error {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))
	return GetError(C.bladerf_set_gain_stage((*bladeRF).bladeRF, C.bladerf_channel(channel), val, C.int(gain)))
}

func GetGainStageRange(bladeRF *BladeRF, channel Channel, stage string) (Range, error) {
	var bfRange *C.struct_bladerf_range
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))

	err := GetError(C.bladerf_get_gain_stage_range((*bladeRF).bladeRF, C.bladerf_channel(channel), val, &bfRange))

	if err == nil {
		return Range{
			bfRange: bfRange,
			min:     int64(bfRange.min),
			max:     int64(bfRange.max),
			step:    int64(bfRange.step),
			scale:   float64(bfRange.scale),
		}, nil
	}

	return Range{}, err
}

func GetGainRange(bladeRF *BladeRF, channel Channel) (Range, error) {
	var bfRange *C.struct_bladerf_range

	err := GetError(C.bladerf_get_gain_range((*bladeRF).bladeRF, C.bladerf_channel(channel), &bfRange))

	if err == nil {
		return Range{
			bfRange: bfRange,
			min:     int64(bfRange.min),
			max:     int64(bfRange.max),
			step:    int64(bfRange.step),
			scale:   float64(bfRange.scale),
		}, nil
	}

	return Range{}, err
}

func GetNumberOfGainStages(bladeRF *BladeRF, channel Channel) int {
	count := int(C.bladerf_get_gain_stages((*bladeRF).bladeRF, C.bladerf_channel(channel), nil, 0))

	if count < 1 {
		return 0
	}

	return count
}

func GetGainStages(bladeRF *BladeRF, channel Channel) []string {
	var stage *C.char
	var stages []string

	count := int(C.bladerf_get_gain_stages(
		(*bladeRF).bladeRF,
		C.bladerf_channel(channel),
		&stage,
		C.ulong(GetNumberOfGainStages(bladeRF, channel))),
	)

	if count < 1 {
		return stages
	}

	first := C.GoString(stage)
	stages = append(stages, first)

	for i := 0; i < count-1; i++ {
		size := unsafe.Sizeof(*stage)
		stage = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(stage)) + size))
		stages = append(stages, C.GoString(stage))
	}

	return stages
}

func GetGainModes(bladeRF *BladeRF, channel Channel) []GainModes {
	var gainMode *C.struct_bladerf_gain_modes
	var gainModes []GainModes

	count := int(C.bladerf_get_gain_modes((*bladeRF).bladeRF, C.bladerf_channel(channel), &gainMode))

	if count < 1 {
		return gainModes
	}

	first := GainModes{
		gainModes: gainMode,
		name:      C.GoString(gainMode.name),
		mode:      gain_mode.GainMode(gainMode.mode),
	}

	gainModes = append(gainModes, first)

	for i := 0; i < count-1; i++ {
		size := unsafe.Sizeof(*gainMode)
		gainMode = (*C.struct_bladerf_gain_modes)(unsafe.Pointer(uintptr(unsafe.Pointer(gainMode)) + size))
		gainModes = append(gainModes, GainModes{
			gainModes: gainMode,
			name:      C.GoString(gainMode.name),
			mode:      gain_mode.GainMode(gainMode.mode),
		})
	}

	return gainModes
}

func SetGainMode(bladeRF *BladeRF, channel Channel, mode gain_mode.GainMode) error {
	return GetError(C.bladerf_set_gain_mode((*bladeRF).bladeRF, C.bladerf_channel(channel), C.bladerf_gain_mode(mode)))
}

func EnableModule(bladeRF *BladeRF, channel Channel) error {
	return GetError(C.bladerf_enable_module((*bladeRF).bladeRF, C.bladerf_channel(channel), true))
}

func DisableModule(bladeRF *BladeRF, channel Channel) error {
	return GetError(C.bladerf_enable_module((*bladeRF).bladeRF, C.bladerf_channel(channel), false))
}

func SyncRX(bladeRF *BladeRF, bufferSize uintptr) []int16 {
	var metadata C.struct_bladerf_metadata

	start := C.malloc(C.size_t(size * bufferSize * 2 * 1))

	var err error
	var results []int16

	err = GetError(C.bladerf_sync_rx((*bladeRF).bladeRF, start, C.uint(bufferSize), &metadata, 32))
	if err == nil {
		for i := 0; i < (int(metadata.actual_count)); i++ {
			n := (*C.int16_t)(unsafe.Pointer(uintptr(start) + (size * uintptr(i))))
			results = append(results, int16(*n))
		}
	} else {
		fmt.Printf("Failed to RX samples: %s", err)
	}

	return results
}

type Callback struct {
	cb         func(data []int16)
	results    []int16
	bufferSize int
}

func InitStream(bladeRF *BladeRF, format format.Format, numBuffers int, samplesPerBuffer int, numTransfers int, callback func(data []int16)) *Stream {
	var buffers *unsafe.Pointer
	var rxStream *C.struct_bladerf_stream

	stream := Stream{stream: rxStream}

	results := make([]int16, samplesPerBuffer)

	cb := Callback{
		cb:         callback,
		results:    results,
		bufferSize: samplesPerBuffer,
	}

	C.bladerf_init_stream(
		&((stream).stream),
		(*bladeRF).bladeRF,
		(*[0]byte)((C.cbGo)),
		&buffers,
		C.ulong(numBuffers),
		C.bladerf_format(format),
		C.ulong(samplesPerBuffer),
		C.ulong(numTransfers),
		pointer.Save(cb),
	)

	return &stream
}

func DeInitStream(stream *Stream) {
	C.bladerf_deinit_stream(stream.stream)
}

func GetStreamTimeout(bladeRF *BladeRF, direction direction.Direction) (int, error) {
	var timeout C.uint
	err := GetError(C.bladerf_get_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), &timeout))
	return int(timeout), err
}

func SetStreamTimeout(bladeRF *BladeRF, direction direction.Direction, timeout int) error {
	return GetError(C.bladerf_set_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), C.uint(timeout)))
}

func SyncConfig(bladeRF *BladeRF, layout channel_layout.ChannelLayout, format format.Format, numBuffers int, bufferSize int, numTransfers int, timeout int) error {
	return GetError(C.bladerf_sync_config((*bladeRF).bladeRF, C.bladerf_channel_layout(layout), C.bladerf_format(format), C.uint(numBuffers), C.uint(bufferSize), C.uint(numTransfers), C.uint(timeout)))
}

func StartStream(stream *Stream, layout channel_layout.ChannelLayout) error {
	return GetError(C.bladerf_stream(stream.stream, C.bladerf_channel_layout(layout)))
}
