package bladerf

// #include "macro_wrapper.h"
import "C"

// #cgo darwin CFLAGS: -I/usr/local/include
// #cgo darwin LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lbladeRF
// #include <libbladeRF.h>
//
// extern void* cbGo(struct bladerf *dev, struct bladerf_stream *stream, struct bladerf_metadata *md, void* samples, size_t num_samples, void* user_data);
import "C"
import (
	error2 "bladerf/error"
	"fmt"
	"github.com/mattn/go-pointer"
	"unsafe"
)

const size = 2

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

type GainMode int
type Backend int
type Direction int
type ChannelLayout int
type Correction int
type Format int
type Loopback int
type RXMux int
type ClockSelect int
type PowerSource int
type PMICRegister int
type IOModule int
type Channel int

const (
	IOTX IOModule = C.BLADERF_MODULE_TX
	IORX IOModule = C.BLADERF_MODULE_RX
)

const (
	Default        GainMode = C.BLADERF_GAIN_DEFAULT
	Manual         GainMode = C.BLADERF_GAIN_MGC
	FastAttack_AGC GainMode = C.BLADERF_GAIN_FASTATTACK_AGC
	SlowAttack_AGC GainMode = C.BLADERF_GAIN_SLOWATTACK_AGC
	Hybrid_AGC     GainMode = C.BLADERF_GAIN_HYBRID_AGC
)

const (
	Any     Backend = C.BLADERF_BACKEND_ANY
	Linux   Backend = C.BLADERF_BACKEND_LINUX
	LibUSB  Backend = C.BLADERF_BACKEND_LIBUSB
	Cypress Backend = C.BLADERF_BACKEND_CYPRESS
	Dummy   Backend = C.BLADERF_BACKEND_DUMMY
)

const (
	TX Direction = C.BLADERF_TX
	RX Direction = C.BLADERF_RX
)

const (
	RX_X1 ChannelLayout = C.BLADERF_RX_X1
	TX_X1 ChannelLayout = C.BLADERF_TX_X1
	RX_X2 ChannelLayout = C.BLADERF_RX_X2
	TX_X2 ChannelLayout = C.BLADERF_TX_X2
)

const (
	DCOFF_I Correction = C.BLADERF_CORR_DCOFF_I
	DCOFF_Q Correction = C.BLADERF_CORR_DCOFF_Q
	PHASE   Correction = C.BLADERF_CORR_PHASE
	GAIN    Correction = C.BLADERF_CORR_GAIN
)

const (
	SC16_Q11      Format = C.BLADERF_FORMAT_SC16_Q11
	SC16_Q11_META Format = C.BLADERF_FORMAT_SC16_Q11_META
)

const (
	Disabled         Loopback = C.BLADERF_LB_NONE
	Firmware         Loopback = C.BLADERF_LB_FIRMWARE
	BB_TXLPF_RXVGA2  Loopback = C.BLADERF_LB_BB_TXLPF_RXVGA2
	BB_TXVGA1_RXVGA2 Loopback = C.BLADERF_LB_BB_TXVGA1_RXVGA2
	BB_TXLPF_RXLPF   Loopback = C.BLADERF_LB_BB_TXLPF_RXLPF
	BB_TXVGA1_RXLPF  Loopback = C.BLADERF_LB_BB_TXVGA1_RXLPF
	RF_LNA1          Loopback = C.BLADERF_LB_RF_LNA1
	RF_LNA2          Loopback = C.BLADERF_LB_RF_LNA2
	RF_LNA3          Loopback = C.BLADERF_LB_RF_LNA3
	RFIC_BIST        Loopback = C.BLADERF_LB_RFIC_BIST
)

const (
	Invalid          RXMux = C.BLADERF_RX_MUX_INVALID
	Baseband         RXMux = C.BLADERF_RX_MUX_BASEBAND
	Counter_12bit    RXMux = C.BLADERF_RX_MUX_12BIT_COUNTER
	Counter_32bit    RXMux = C.BLADERF_RX_MUX_32BIT_COUNTER
	Digital_Loopback RXMux = C.BLADERF_RX_MUX_DIGITAL_LOOPBACK
)

const (
	ClockSelectUnknown  ClockSelect = -99
	ClockSelectVCTCXO   ClockSelect = C.CLOCK_SELECT_ONBOARD
	ClockSelectExternal ClockSelect = C.CLOCK_SELECT_EXTERNAL
)

const (
	PowerSourceUnknown   PowerSource = C.BLADERF_UNKNOWN
	PowerSourceDC_Barrel PowerSource = C.BLADERF_PS_DC
	PowerSourceUSB_VBUS  PowerSource = C.BLADERF_PS_USB_VBUS
)

const (
	Configuration PMICRegister = C.BLADERF_PMIC_CONFIGURATION
	Voltage_shunt PMICRegister = C.BLADERF_PMIC_VOLTAGE_SHUNT
	Voltage_bus   PMICRegister = C.BLADERF_PMIC_VOLTAGE_BUS
	Power         PMICRegister = C.BLADERF_PMIC_POWER
	Current       PMICRegister = C.BLADERF_PMIC_CURRENT
	Calibration   PMICRegister = C.BLADERF_PMIC_CALIBRATION
)

const (
	RX_V int = 0x0
	TX_V int = 0x1
)

const (
	BLADERF_VCTCXO_FREQUENCY = 38.4e6
	BLADERF_REFIN_DEFAULT    = 10.0e6
	BLADERF_SERIAL_LENGTH    = 33
)

func checkError(e error) {
	if e != nil {
		fmt.Printf(e.Error())
	}
}

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
	mode      GainMode
}

func GetVersion() Version {
	var version C.struct_bladerf_version
	C.bladerf_version(&version)
	return Version{version: &version, major: int(version.major), minor: int(version.minor), patch: int(version.patch), describe: C.GoString(version.describe)}
}

func PrintVersion(version Version) {
	fmt.Printf("v%d.%d.%d (\"%s\")", version.major, version.minor, version.patch, version.describe)
}

func GetError(e C.int) error {
	if e == 0 {
		return nil
	}

	return error2.Error(e)
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

func CHANNEL_RX(ch int) Channel {
	return Channel(C.ChannelRX(C.int(ch)))
}

func CHANNEL_TX(ch int) Channel {
	return Channel(C.ChannelTX(C.int(ch)))
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

func SetLoopback(bladeRF *BladeRF, loopback Loopback) {
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

func GetGainMode(bladeRF *BladeRF, channel Channel) (GainMode, error) {
	var mode C.bladerf_gain_mode

	err := GetError(C.bladerf_get_gain_mode((*bladeRF).bladeRF, C.bladerf_channel(channel), &mode))
	result := GainMode(int(mode))
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

func GetGainModes(bladeRF *BladeRF, module IOModule) []GainModes {
	var gainMode *C.struct_bladerf_gain_modes
	var gainModes []GainModes

	count := int(C.bladerf_get_gain_modes((*bladeRF).bladeRF, C.bladerf_module(module), &gainMode))

	if count < 1 {
		return gainModes
	}

	first := GainModes{
		gainModes: gainMode,
		name:      C.GoString(gainMode.name),
		mode:      GainMode(gainMode.mode),
	}

	gainModes = append(gainModes, first)

	for i := 0; i < count-1; i++ {
		size := unsafe.Sizeof(*gainMode)
		gainMode = (*C.struct_bladerf_gain_modes)(unsafe.Pointer(uintptr(unsafe.Pointer(gainMode)) + size))
		gainModes = append(gainModes, GainModes{
			gainModes: gainMode,
			name:      C.GoString(gainMode.name),
			mode:      GainMode(gainMode.mode),
		})
	}

	return gainModes
}

func SetGainMode(bladeRF *BladeRF, channel Channel, mode GainMode) error {
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

	// var buff *C.int16_t
	// size := unsafe.Sizeof(*buff)
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

func InitStream(bladeRF *BladeRF, format Format, numBuffers int, samplesPerBuffer int, numTransfers int, callback func(data []int16)) *Stream {
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

func GetStreamTimeout(bladeRF *BladeRF, direction Direction) (int, error) {
	var timeout C.uint
	err := GetError(C.bladerf_get_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), &timeout))
	return int(timeout), err
}

func SetStreamTimeout(bladeRF *BladeRF, direction Direction, timeout int) error {
	return GetError(C.bladerf_set_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), C.uint(timeout)))
}

func SyncConfig(bladeRF *BladeRF, layout ChannelLayout, format Format, numBuffers int, bufferSize int, numTransfers int, timeout int) error {
	err := GetError(C.bladerf_sync_config((*bladeRF).bladeRF, C.bladerf_channel_layout(layout), C.bladerf_format(format), C.uint(numBuffers), C.uint(bufferSize), C.uint(numTransfers), C.uint(timeout)))
	return err
}

func StartStream(stream *Stream, layout ChannelLayout) error {
	return GetError(C.bladerf_stream(stream.stream, C.bladerf_channel_layout(layout)))
}
