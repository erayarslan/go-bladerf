package bladerf

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
	"unsafe"
)

//export cbGo
func cbGo(dev *C.struct_bladerf,
	stream *C.struct_bladerf_stream,
	metadata *C.struct_bladerf_metadata,
	samples unsafe.Pointer,
	numSamples C.size_t,
	userData unsafe.Pointer) unsafe.Pointer {

	data := (*[0]uint16)(samples)

	if len(*data) > 0 {
		println(data)
	} else {
		println(data)
	}

	var rv unsafe.Pointer
	return rv
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
	C.bladerf_close(bladeRF.bladeRF)
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

func CHANNEL_RX(ch int) int {
	return (ch << 1) | RX_V
}

func CHANNEL_TX(ch int) int {
	return (ch << 1) | TX_V
}

func GetDevInfo() DevInfo {
	var devInfo C.struct_bladerf_devinfo
	C.bladerf_init_devinfo(&devInfo)
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

func SetFrequency(bladeRF *BladeRF, module IOModule, frequency int) {
	err := GetError(C.bladerf_set_frequency((*bladeRF).bladeRF, C.bladerf_module(module), C.ulonglong(frequency)))
	fmt.Println(err)
}

func SetSampleRate(bladeRF *BladeRF, module IOModule, sampleRate int) {
	C.bladerf_set_sample_rate((*bladeRF).bladeRF, C.bladerf_module(module), C.uint(sampleRate), nil)
}

func SetBandwidth(bladeRF *BladeRF, module IOModule, bandwidth int) {
	C.bladerf_set_bandwidth((*bladeRF).bladeRF, C.bladerf_module(module), C.uint(bandwidth), nil)
}

func SetGain(bladeRF *BladeRF, module IOModule, gain int) {
	C.bladerf_set_gain((*bladeRF).bladeRF, C.bladerf_module(module), C.int(gain))
}

func EnableModule(bladeRF *BladeRF, direction Direction) {
	C.bladerf_enable_module((*bladeRF).bladeRF, C.bladerf_module(direction), true)
}

func DisableModule(bladeRF *BladeRF, direction Direction) {
	C.bladerf_enable_module((*bladeRF).bladeRF, C.bladerf_module(direction), false)
}

func SyncRX(bladeRF *BladeRF) {
	var metadata *C.struct_bladerf_metadata

	var buff *C.int16_t
	size := unsafe.Sizeof(*buff)
	start := C.malloc(C.size_t(size * 10000 * 2 * 1))

	err := GetError(C.bladerf_sync_rx((*bladeRF).bladeRF, start, 10000, metadata, 5000))
	fmt.Println(err)
}

func InitStream(bladeRF *BladeRF, format Format, numBuffers int, samplesPerBuffer int, numTransfers int) *Stream {
	var buffers *unsafe.Pointer
	var data unsafe.Pointer
	var rxStream *C.struct_bladerf_stream
	stream := Stream{stream: rxStream}
	C.bladerf_init_stream(&((stream).stream), (*bladeRF).bladeRF, (*[0]byte)((C.cbGo)), &buffers, C.ulong(numBuffers), C.bladerf_format(format), C.ulong(samplesPerBuffer), C.ulong(numTransfers), data)
	return &stream
}

func DeInitStream(stream *Stream) {
	C.bladerf_deinit_stream(stream.stream)
}

func GetStreamTimeout(bladeRF *BladeRF, direction Direction) int {
	var timeout C.uint
	err := GetError(C.bladerf_get_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), &timeout))
	fmt.Println(err)
	return int(timeout)
}

func SetStreamTimeout(bladeRF *BladeRF, direction Direction, timeout int) {
	err := GetError(C.bladerf_set_stream_timeout((*bladeRF).bladeRF, C.bladerf_direction(direction), C.uint(timeout)))
	fmt.Println(err)
}

func SyncConfig(bladeRF *BladeRF, layout ChannelLayout, format Format, numBuffers int, bufferSize int, numTransfers int, timeout int) {
	err := GetError(C.bladerf_sync_config((*bladeRF).bladeRF, C.bladerf_channel_layout(layout), C.bladerf_format(format), C.uint(numBuffers), C.uint(bufferSize), C.uint(numTransfers), C.uint(timeout)))
	fmt.Println(err)
}

func StartStream(stream *Stream, layout ChannelLayout) {
	err := GetError(C.bladerf_stream(stream.stream, C.bladerf_channel_layout(layout)))
	fmt.Println(err)
}
