package bladerf

import "C"

// #cgo darwin CFLAGS: -I/usr/local/include
// #cgo darwin LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lbladeRF
// #include <libbladeRF.h>
//
// extern void* StreamCallback(struct bladerf *dev, struct bladerf_stream *stream, struct bladerf_metadata *md, void* samples, size_t num_samples, void* user_data);
import "C"
import (
	"fmt"
	"github.com/erayarslan/go-bladerf/channel"
	"github.com/erayarslan/go-bladerf/channel_layout"
	"github.com/erayarslan/go-bladerf/direction"
	exception "github.com/erayarslan/go-bladerf/error"
	"github.com/erayarslan/go-bladerf/format"
	"github.com/erayarslan/go-bladerf/gain_mode"
	"github.com/erayarslan/go-bladerf/loopback"
	"github.com/mattn/go-pointer"
	"unsafe"
)

func GetError(code C.int) error {
	return exception.New(int(code))
}

//export StreamCallback
func StreamCallback(
	dev *C.struct_bladerf,
	stream *C.struct_bladerf_stream,
	metadata *C.struct_bladerf_metadata,
	samples unsafe.Pointer,
	numSamples C.size_t,
	userData unsafe.Pointer,
) unsafe.Pointer {
	ud := pointer.Restore(userData).(UserData)

	for i := uint32(0); i < uint32(numSamples); i++ {
		ud.results[i] = int16(*((*C.int16_t)(unsafe.Pointer(uintptr(samples) + (C.sizeof_int16_t * uintptr(i))))))
	}

	ud.callback(ud.results)
	return samples
}

func GetVersion() Version {
	var version C.struct_bladerf_version
	C.bladerf_version(&version)
	return NewVersion(&version)
}

func PrintVersion(version Version) {
	fmt.Printf("v%d.%d.%d (\"%s\")", version.major, version.minor, version.patch, version.describe)
}

func LoadFpga(bladeRF BladeRF, imagePath string) error {
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))
	return GetError(C.bladerf_load_fpga(bladeRF.ref, path))
}

func FreeDeviceList(devInfo DevInfo) {
	C.bladerf_free_device_list(devInfo.ref)
}

func GetDeviceList() []DevInfo {
	var devInfo *C.struct_bladerf_devinfo
	var devices []DevInfo

	count := int(C.bladerf_get_device_list(&devInfo))

	if count > 0 {
		size := unsafe.Sizeof(*devInfo)

		for i := 0; i < count; i++ {
			devices = append(devices, NewDevInfo(
				(*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(devInfo))+(uintptr(i)*size))),
			))
		}

		FreeDeviceList(devices[0])
	}

	return devices
}

func GetBootloaderList() []DevInfo {
	var devInfo *C.struct_bladerf_devinfo
	var devices []DevInfo

	count := int(C.bladerf_get_bootloader_list(&devInfo))

	if count > 0 {
		size := unsafe.Sizeof(*devInfo)

		for i := 0; i < count; i++ {
			devices = append(devices, NewDevInfo(
				(*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(devInfo))+(uintptr(i)*size))),
			))
		}

		FreeDeviceList(devices[0])
	}

	return devices
}

func InitDevInfo() DevInfo {
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_init_devinfo(&devInfo)
	return NewDevInfo(&devInfo)
}

func GetDevInfo(bladeRF *BladeRF) DevInfo {
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_get_devinfo((*bladeRF).ref, &devInfo)
	return NewDevInfo(&devInfo)
}

func DevInfoMatches(a DevInfo, b DevInfo) bool {
	return bool(C.bladerf_devinfo_matches(a.ref, b.ref))
}

func DevStrMatches(devStr string, info DevInfo) bool {
	val := C.CString(devStr)
	defer C.free(unsafe.Pointer(val))
	return bool(C.bladerf_devstr_matches(val, info.ref))
}

func GetDevInfoFromStr(devStr string) DevInfo {
	val := C.CString(devStr)
	defer C.free(unsafe.Pointer(val))
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_get_devinfo_from_str(val, &devInfo)
	return NewDevInfo(&devInfo)
}

func OpenWithDevInfo(devInfo DevInfo) (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	err := GetError(C.bladerf_open_with_devinfo(&bladeRF, devInfo.ref))

	if err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func OpenWithDeviceIdentifier(identify string) BladeRF {
	var bladeRF *C.struct_bladerf
	C.bladerf_open(&bladeRF, C.CString(identify))
	return BladeRF{ref: bladeRF}
}

func Open() BladeRF {
	var bladeRF *C.struct_bladerf
	C.bladerf_open(&bladeRF, nil)
	return BladeRF{ref: bladeRF}
}

func Close(bladeRF BladeRF) {
	C.bladerf_close(bladeRF.ref)
}

func SetLoopback(bladeRF *BladeRF, loopback loopback.Loopback) error {
	return GetError(C.bladerf_set_loopback((*bladeRF).ref, C.bladerf_loopback(loopback)))
}

func SetFrequency(bladeRF *BladeRF, channel channel.Channel, frequency int) error {
	return GetError(C.bladerf_set_frequency((*bladeRF).ref, C.bladerf_channel(channel), C.ulonglong(frequency)))
}

func SetSampleRate(bladeRF *BladeRF, channel channel.Channel, sampleRate int) error {
	var actual C.uint
	err := GetError(C.bladerf_set_sample_rate((*bladeRF).ref, C.bladerf_channel(channel), C.uint(sampleRate), &actual))
	return err
}

func GetSampleRateRange(bladeRF *BladeRF, channel channel.Channel) (int, int, int, error) {
	var bfRange *C.struct_bladerf_range

	err := GetError(C.bladerf_get_sample_rate_range((*bladeRF).ref, C.bladerf_channel(channel), &bfRange))

	if err != nil {
		return 0, 0, 0, err
	}

	return int(bfRange.min), int(bfRange.max), int(bfRange.step), nil
}

func SetBandwidth(bladeRF *BladeRF, channel channel.Channel, bandwidth int) (int, error) {
	var actual C.bladerf_bandwidth
	return int(actual), GetError(C.bladerf_set_bandwidth((*bladeRF).ref, C.bladerf_channel(channel), C.uint(bandwidth), &actual))
}

func SetGain(bladeRF *BladeRF, channel channel.Channel, gain int) error {
	return GetError(C.bladerf_set_gain((*bladeRF).ref, C.bladerf_channel(channel), C.int(gain)))
}

func GetGain(bladeRF *BladeRF, channel channel.Channel) (int, error) {
	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain((*bladeRF).ref, C.bladerf_channel(channel), &gain))
	if err == nil {
		return int(gain), nil
	}

	return int(gain), err
}

func GetGainStage(bladeRF *BladeRF, channel channel.Channel, stage string) (int, error) {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))
	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain_stage((*bladeRF).ref, C.bladerf_channel(channel), val, &gain))
	if err == nil {
		return int(gain), nil
	}

	return int(gain), err
}

func GetGainMode(bladeRF *BladeRF, channel channel.Channel) (gain_mode.GainMode, error) {
	var mode C.bladerf_gain_mode

	err := GetError(C.bladerf_get_gain_mode((*bladeRF).ref, C.bladerf_channel(channel), &mode))
	result := gain_mode.GainMode(int(mode))
	if err == nil {
		return result, nil
	}

	return result, err
}

func SetGainStage(bladeRF *BladeRF, channel channel.Channel, stage string, gain int) error {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))
	return GetError(C.bladerf_set_gain_stage((*bladeRF).ref, C.bladerf_channel(channel), val, C.int(gain)))
}

func GetGainStageRange(bladeRF *BladeRF, channel channel.Channel, stage string) (Range, error) {
	var bfRange *C.struct_bladerf_range
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))

	err := GetError(C.bladerf_get_gain_stage_range((*bladeRF).ref, C.bladerf_channel(channel), val, &bfRange))

	if err == nil {
		return Range{
			ref:   bfRange,
			min:   int64(bfRange.min),
			max:   int64(bfRange.max),
			step:  int64(bfRange.step),
			scale: float64(bfRange.scale),
		}, nil
	}

	return Range{}, err
}

func GetGainRange(bladeRF *BladeRF, channel channel.Channel) (Range, error) {
	var bfRange *C.struct_bladerf_range

	err := GetError(C.bladerf_get_gain_range((*bladeRF).ref, C.bladerf_channel(channel), &bfRange))

	if err == nil {
		return Range{
			ref:   bfRange,
			min:   int64(bfRange.min),
			max:   int64(bfRange.max),
			step:  int64(bfRange.step),
			scale: float64(bfRange.scale),
		}, nil
	}

	return Range{}, err
}

func GetNumberOfGainStages(bladeRF *BladeRF, channel channel.Channel) int {
	count := int(C.bladerf_get_gain_stages((*bladeRF).ref, C.bladerf_channel(channel), nil, 0))

	if count < 1 {
		return 0
	}

	return count
}

func GetGainStages(bladeRF *BladeRF, channel channel.Channel) []string {
	var stage *C.char
	var stages []string

	count := int(C.bladerf_get_gain_stages(
		(*bladeRF).ref,
		C.bladerf_channel(channel),
		&stage,
		C.ulong(GetNumberOfGainStages(bladeRF, channel))),
	)

	if count > 0 {
		size := unsafe.Sizeof(*stage)

		for i := 0; i < count; i++ {
			stages = append(stages, C.GoString(
				(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(stage))+(uintptr(i)*size))),
			))
		}
	}

	return stages
}

func GetGainModes(bladeRF *BladeRF, channel channel.Channel) []GainModes {
	var gainMode *C.struct_bladerf_gain_modes
	var gainModes []GainModes

	count := int(C.bladerf_get_gain_modes((*bladeRF).ref, C.bladerf_channel(channel), &gainMode))

	if count > 0 {
		size := unsafe.Sizeof(*gainMode)

		for i := 0; i < count; i++ {
			gainModes = append(gainModes, NewGainModes(
				(*C.struct_bladerf_gain_modes)(unsafe.Pointer(uintptr(unsafe.Pointer(gainMode))+(uintptr(i)*size))),
			))
		}
	}

	return gainModes
}

func SetGainMode(bladeRF *BladeRF, channel channel.Channel, mode gain_mode.GainMode) error {
	return GetError(C.bladerf_set_gain_mode((*bladeRF).ref, C.bladerf_channel(channel), C.bladerf_gain_mode(mode)))
}

func EnableModule(bladeRF *BladeRF, channel channel.Channel) error {
	return GetError(C.bladerf_enable_module((*bladeRF).ref, C.bladerf_channel(channel), true))
}

func DisableModule(bladeRF *BladeRF, channel channel.Channel) error {
	return GetError(C.bladerf_enable_module((*bladeRF).ref, C.bladerf_channel(channel), false))
}

func SyncRX(bladeRF *BladeRF, bufferSize uintptr) ([]int16, error) {
	var metadata C.struct_bladerf_metadata

	start := C.malloc(C.size_t(C.sizeof_int16_t * bufferSize * 2 * 1))

	var err error
	var results []int16

	err = GetError(C.bladerf_sync_rx((*bladeRF).ref, start, C.uint(bufferSize), &metadata, 32))

	if err == nil {
		for i := 0; i < (int(metadata.actual_count)); i++ {
			n := (*C.int16_t)(unsafe.Pointer(uintptr(start) + (C.sizeof_int16_t * uintptr(i))))
			results = append(results, int16(*n))
		}

		return results, nil
	}

	return nil, err
}

func InitStream(bladeRF *BladeRF, format format.Format, numBuffers int, samplesPerBuffer int, numTransfers int, callback func(data []int16)) *Stream {
	var buffers *unsafe.Pointer
	var rxStream *C.struct_bladerf_stream

	stream := Stream{ref: rxStream}

	C.bladerf_init_stream(
		&((stream).ref),
		(*bladeRF).ref,
		(*[0]byte)((C.StreamCallback)),
		&buffers,
		C.ulong(numBuffers),
		C.bladerf_format(format),
		C.ulong(samplesPerBuffer),
		C.ulong(numTransfers),
		pointer.Save(NewUserData(callback, samplesPerBuffer)),
	)

	return &stream
}

func DeInitStream(stream *Stream) {
	C.bladerf_deinit_stream(stream.ref)
}

func GetStreamTimeout(bladeRF *BladeRF, direction direction.Direction) (int, error) {
	var timeout C.uint
	err := GetError(C.bladerf_get_stream_timeout((*bladeRF).ref, C.bladerf_direction(direction), &timeout))
	return int(timeout), err
}

func SetStreamTimeout(bladeRF *BladeRF, direction direction.Direction, timeout int) error {
	return GetError(C.bladerf_set_stream_timeout((*bladeRF).ref, C.bladerf_direction(direction), C.uint(timeout)))
}

func SyncConfig(bladeRF *BladeRF, layout channel_layout.ChannelLayout, format format.Format, numBuffers int, bufferSize int, numTransfers int, timeout int) error {
	return GetError(C.bladerf_sync_config((*bladeRF).ref, C.bladerf_channel_layout(layout), C.bladerf_format(format), C.uint(numBuffers), C.uint(bufferSize), C.uint(numTransfers), C.uint(timeout)))
}

func StartStream(stream *Stream, layout channel_layout.ChannelLayout) error {
	return GetError(C.bladerf_stream(stream.ref, C.bladerf_channel_layout(layout)))
}
