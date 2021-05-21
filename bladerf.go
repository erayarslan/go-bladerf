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
	exception "github.com/erayarslan/go-bladerf/error"
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
	userDataPtr unsafe.Pointer,
) unsafe.Pointer {
	userData := pointer.Restore(userDataPtr).(UserData)

	for i := uint32(0); i < uint32(numSamples); i++ {
		userData.results[i] = int16(
			*((*C.int16_t)(unsafe.Pointer(uintptr(samples) + (C.sizeof_int16_t * uintptr(i))))),
		)
	}

	userData.callback(userData.results)
	return samples
}

func GetVersion() Version {
	var version C.struct_bladerf_version
	C.bladerf_version(&version)
	return NewVersion(&version)
}

func (version *Version) Print() {
	fmt.Printf("v%d.%d.%d (\"%s\")", version.major, version.minor, version.patch, version.describe)
}

func (bladeRF *BladeRF) LoadFpga(imagePath string) error {
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))
	return GetError(C.bladerf_load_fpga(bladeRF.ref, path))
}

func (bladeRF *BladeRF) GetFpgaSize() (FpgaSize, error) {
	var size C.bladerf_fpga_size
	err := GetError(C.bladerf_get_fpga_size(bladeRF.ref, &size))

	if err != nil {
		return 0, err
	}

	return FpgaSize(size), nil
}

func (bladeRF *BladeRF) GetQuickTune(channel Channel) (QuickTune, error) {
	var quickTune C.struct_bladerf_quick_tune

	err := GetError(C.bladerf_get_quick_tune(bladeRF.ref, C.bladerf_channel(channel), &quickTune))

	if err != nil {
		return QuickTune{}, err
	}

	return QuickTune{ref: &quickTune}, nil
}

func (bladeRF *BladeRF) CancelScheduledReTunes(channel Channel) error {
	return GetError(C.bladerf_cancel_scheduled_retunes(bladeRF.ref, C.bladerf_channel(channel)))
}

func (bladeRF *BladeRF) GetFpgaSource() (FpgaSource, error) {
	var source C.bladerf_fpga_source
	err := GetError(C.bladerf_get_fpga_source(bladeRF.ref, &source))

	if err != nil {
		return 0, err
	}

	return FpgaSource(source), nil
}

func (bladeRF *BladeRF) GetFpgaBytes() (uint32, error) {
	var size C.size_t
	err := GetError(C.bladerf_get_fpga_bytes(bladeRF.ref, &size))

	if err != nil {
		return 0, err
	}

	return uint32(size), nil
}

func (bladeRF *BladeRF) GetFpgaFlashSize() (uint32, bool, error) {
	var size C.uint32_t
	var isGuess C.bool
	err := GetError(C.bladerf_get_flash_size(bladeRF.ref, &size, &isGuess))

	if err != nil {
		return 0, false, err
	}

	return uint32(size), bool(isGuess), nil
}

func (bladeRF *BladeRF) GetFirmwareVersion() (Version, error) {
	var version C.struct_bladerf_version
	err := GetError(C.bladerf_fw_version(bladeRF.ref, &version))

	if err != nil {
		return Version{}, err
	}

	return NewVersion(&version), nil
}

func (bladeRF *BladeRF) IsFpgaConfigured() (bool, error) {
	out := C.bladerf_is_fpga_configured(bladeRF.ref)

	if out < 0 {
		return false, GetError(out)
	}

	return out == 1, nil
}

func (bladeRF *BladeRF) GetDeviceSpeed() DeviceSpeed {
	return DeviceSpeed(int(C.bladerf_device_speed(bladeRF.ref)))
}

func (bladeRF *BladeRF) GetFpgaVersion() (Version, error) {
	var version C.struct_bladerf_version
	err := GetError(C.bladerf_fpga_version(bladeRF.ref, &version))

	if err != nil {
		return Version{}, err
	}

	return NewVersion(&version), nil
}

func (deviceInfo *DeviceInfo) FreeDeviceList() {
	C.bladerf_free_device_list(deviceInfo.ref)
}

func GetDeviceList() ([]DeviceInfo, error) {
	var deviceInfo *C.struct_bladerf_devinfo
	var devices []DeviceInfo

	codeOrCount := C.bladerf_get_device_list(&deviceInfo)

	if codeOrCount < 0 {
		return nil, GetError(codeOrCount)
	}

	count := int(codeOrCount)

	if count == 0 {
		return make([]DeviceInfo, 0), nil
	}

	if count > 0 {
		size := unsafe.Sizeof(*deviceInfo)

		for i := 0; i < count; i++ {
			devices = append(devices, NewDeviceInfo(
				(*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(deviceInfo))+(uintptr(i)*size))),
			))
		}

		devices[0].FreeDeviceList()
	}

	return devices, nil
}

func GetBootloaderList() ([]DeviceInfo, error) {
	var deviceInfo *C.struct_bladerf_devinfo
	var devices []DeviceInfo

	codeOrCount := C.bladerf_get_bootloader_list(&deviceInfo)

	if codeOrCount < 0 {
		return nil, GetError(codeOrCount)
	}

	count := int(codeOrCount)

	if count == 0 {
		return make([]DeviceInfo, 0), nil
	}

	if count > 0 {
		size := unsafe.Sizeof(*deviceInfo)

		for i := 0; i < count; i++ {
			devices = append(devices, NewDeviceInfo(
				(*C.struct_bladerf_devinfo)(unsafe.Pointer(uintptr(unsafe.Pointer(deviceInfo))+(uintptr(i)*size))),
			))
		}

		devices[0].FreeDeviceList()
	}

	return devices, nil
}

func InitDeviceInfo() DeviceInfo {
	var deviceInfo C.struct_bladerf_devinfo
	C.bladerf_init_devinfo(&deviceInfo)
	return NewDeviceInfo(&deviceInfo)
}

func (bladeRF *BladeRF) GetDeviceInfo() (DeviceInfo, error) {
	var deviceInfo C.struct_bladerf_devinfo
	err := GetError(C.bladerf_get_devinfo(bladeRF.ref, &deviceInfo))

	if err != nil {
		return DeviceInfo{}, err
	}

	return NewDeviceInfo(&deviceInfo), nil
}

func (deviceInfo *DeviceInfo) DeviceInfoMatches(target DeviceInfo) bool {
	return bool(C.bladerf_devinfo_matches(deviceInfo.ref, target.ref))
}

func (deviceInfo *DeviceInfo) DeviceStringMatches(deviceString string) bool {
	val := C.CString(deviceString)
	defer C.free(unsafe.Pointer(val))

	return bool(C.bladerf_devstr_matches(val, deviceInfo.ref))
}

func GetDeviceInfoFromString(deviceString string) (DeviceInfo, error) {
	val := C.CString(deviceString)
	defer C.free(unsafe.Pointer(val))

	var deviceInfo C.struct_bladerf_devinfo
	err := GetError(C.bladerf_get_devinfo_from_str(val, &deviceInfo))

	if err != nil {
		return DeviceInfo{}, err
	}

	return NewDeviceInfo(&deviceInfo), nil
}

func (deviceInfo *DeviceInfo) Open() (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	err := GetError(C.bladerf_open_with_devinfo(&bladeRF, deviceInfo.ref))

	if err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func OpenWithDeviceIdentifier(identify string) (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	err := GetError(C.bladerf_open(&bladeRF, C.CString(identify)))

	if err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func Open() (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	err := GetError(C.bladerf_open(&bladeRF, nil))

	if err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func (bladeRF *BladeRF) Close() {
	C.bladerf_close(bladeRF.ref)
}

func (bladeRF *BladeRF) SetLoopback(loopback Loopback) error {
	return GetError(C.bladerf_set_loopback(bladeRF.ref, C.bladerf_loopback(loopback)))
}

func (bladeRF *BladeRF) IsLoopbackModeSupported(loopback Loopback) bool {
	return bool(C.bladerf_is_loopback_mode_supported(bladeRF.ref, C.bladerf_loopback(loopback)))
}

func (bladeRF *BladeRF) GetLoopback() (Loopback, error) {
	var loopback C.bladerf_loopback
	err := GetError(C.bladerf_get_loopback(bladeRF.ref, &loopback))

	if err != nil {
		return 0, err
	}

	return Loopback(loopback), nil
}

func (bladeRF *BladeRF) ScheduleReTune(
	channel Channel,
	timestamp Timestamp,
	frequency uint64,
	quickTune QuickTune,
) error {
	return GetError(C.bladerf_schedule_retune(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_timestamp(timestamp),
		C.bladerf_frequency(frequency),
		quickTune.ref,
	))
}

func (bladeRF *BladeRF) SelectBand(channel Channel, frequency uint64) error {
	return GetError(C.bladerf_select_band(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_frequency(frequency)))
}

func (bladeRF *BladeRF) SetFrequency(channel Channel, frequency uint64) error {
	return GetError(C.bladerf_set_frequency(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_frequency(frequency)))
}

func (bladeRF *BladeRF) GetFrequency(channel Channel) (uint64, error) {
	var frequency C.uint64_t
	err := GetError(C.bladerf_get_frequency(bladeRF.ref, C.bladerf_channel(channel), &frequency))

	if err != nil {
		return 0, err
	}

	return uint64(frequency), nil
}

func (bladeRF *BladeRF) SetSampleRate(channel Channel, sampleRate uint) (uint, error) {
	var actual C.uint
	err := GetError(C.bladerf_set_sample_rate(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_sample_rate(sampleRate),
		&actual),
	)

	if err != nil {
		return 0, err
	}

	return uint(actual), nil
}

func (bladeRF *BladeRF) SetRxMux(mux RxMux) error {
	return GetError(C.bladerf_set_rx_mux(bladeRF.ref, C.bladerf_rx_mux(mux)))
}

func (bladeRF *BladeRF) GetRxMux() (RxMux, error) {
	var rxMux C.bladerf_rx_mux
	err := GetError(C.bladerf_get_rx_mux(bladeRF.ref, &rxMux))

	if err != nil {
		return 0, err
	}

	return RxMux(rxMux), nil
}

func (bladeRF *BladeRF) SetRationalSampleRate(channel Channel, rationalRate RationalRate) (RationalRate, error) {
	var actual C.struct_bladerf_rational_rate
	rationalSampleRate := C.struct_bladerf_rational_rate{
		num:     C.uint64_t(rationalRate.num),
		integer: C.uint64_t(rationalRate.integer),
		den:     C.uint64_t(rationalRate.den),
	}
	err := GetError(C.bladerf_set_rational_sample_rate(
		bladeRF.ref,
		C.bladerf_channel(channel),
		&rationalSampleRate,
		&actual),
	)

	if err != nil {
		return RationalRate{}, err
	}

	return NewRationalRate(&actual), nil
}

func (bladeRF *BladeRF) GetSampleRate(channel Channel) (uint, error) {
	var sampleRate C.uint
	err := GetError(C.bladerf_get_sample_rate(bladeRF.ref, C.bladerf_channel(channel), &sampleRate))

	if err != nil {
		return 0, err
	}

	return uint(sampleRate), nil
}

func (bladeRF *BladeRF) GetRationalSampleRate(channel Channel) (RationalRate, error) {
	var rate C.struct_bladerf_rational_rate
	err := GetError(C.bladerf_get_rational_sample_rate(bladeRF.ref, C.bladerf_channel(channel), &rate))

	if err != nil {
		return RationalRate{}, err
	}

	return NewRationalRate(&rate), nil
}

func (bladeRF *BladeRF) GetSampleRateRange(channel Channel) (Range, error) {
	var _range *C.struct_bladerf_range
	err := GetError(C.bladerf_get_sample_rate_range(bladeRF.ref, C.bladerf_channel(channel), &_range))

	if err != nil {
		return Range{}, err
	}

	return NewRange(_range), nil
}

func (bladeRF *BladeRF) GetFrequencyRange(channel Channel) (Range, error) {
	var _range *C.struct_bladerf_range
	err := GetError(C.bladerf_get_frequency_range(bladeRF.ref, C.bladerf_channel(channel), &_range))

	if err != nil {
		return Range{}, err
	}

	return NewRange(_range), nil
}

func (bladeRF *BladeRF) SetBandwidth(channel Channel, bandwidth uint) (uint, error) {
	var actual C.bladerf_bandwidth
	err := GetError(C.bladerf_set_bandwidth(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_bandwidth(bandwidth),
		&actual),
	)

	if err != nil {
		return 0, err
	}

	return uint(actual), nil
}

func (bladeRF *BladeRF) GetBandwidth(channel Channel) (uint, error) {
	var bandwidth C.bladerf_bandwidth
	err := GetError(C.bladerf_get_bandwidth(bladeRF.ref, C.bladerf_channel(channel), &bandwidth))

	if err != nil {
		return 0, err
	}

	return uint(bandwidth), nil
}

func (bladeRF *BladeRF) GetBandwidthRange(channel Channel) (Range, error) {
	var bfRange *C.struct_bladerf_range
	err := GetError(C.bladerf_get_bandwidth_range(bladeRF.ref, C.bladerf_channel(channel), &bfRange))

	if err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) SetGain(channel Channel, gain int) error {
	return GetError(C.bladerf_set_gain(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_gain(gain)))
}

func (bladeRF *BladeRF) GetGain(channel Channel) (int, error) {
	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain(bladeRF.ref, C.bladerf_channel(channel), &gain))

	if err != nil {
		return 0, err
	}

	return int(gain), nil
}

func (bladeRF *BladeRF) GetGainStage(channel Channel, stage string) (int, error) {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))

	var gain C.bladerf_gain
	err := GetError(C.bladerf_get_gain_stage(bladeRF.ref, C.bladerf_channel(channel), val, &gain))

	if err != nil {
		return 0, err
	}

	return int(gain), nil
}

func (bladeRF *BladeRF) GetGainMode(channel Channel) (GainMode, error) {
	var mode C.bladerf_gain_mode

	err := GetError(C.bladerf_get_gain_mode(bladeRF.ref, C.bladerf_channel(channel), &mode))

	if err != nil {
		return 0, err
	}

	return GainMode(mode), nil
}

func (bladeRF *BladeRF) SetGainStage(channel Channel, stage string, gain int) error {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))

	return GetError(C.bladerf_set_gain_stage(bladeRF.ref, C.bladerf_channel(channel), val, C.bladerf_gain(gain)))
}

func (bladeRF *BladeRF) GetGainStageRange(channel Channel, stage string) (Range, error) {
	val := C.CString(stage)
	defer C.free(unsafe.Pointer(val))

	var _range *C.struct_bladerf_range
	err := GetError(C.bladerf_get_gain_stage_range(bladeRF.ref, C.bladerf_channel(channel), val, &_range))

	if err != nil {
		return Range{}, err
	}

	return NewRange(_range), nil
}

func (bladeRF *BladeRF) GetGainRange(channel Channel) (Range, error) {
	var _range *C.struct_bladerf_range
	err := GetError(C.bladerf_get_gain_range(bladeRF.ref, C.bladerf_channel(channel), &_range))

	if err != nil {
		return Range{}, err
	}

	return NewRange(_range), nil
}

func (bladeRF *BladeRF) GetNumberOfGainStages(channel Channel) (int, error) {
	countOrCode := C.bladerf_get_gain_stages(bladeRF.ref, C.bladerf_channel(channel), nil, 0)

	if countOrCode < 0 {
		return 0, GetError(countOrCode)
	}

	return int(countOrCode), nil
}

func (bladeRF *BladeRF) SetCorrection(channel Channel, correction Correction, correctionValue int16) error {
	return GetError(C.bladerf_set_correction(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_correction(correction),
		C.bladerf_correction_value(correctionValue)),
	)
}

func (bladeRF *BladeRF) GetCorrection(channel Channel, correction Correction) (uint16, error) {
	var correctionValue C.int16_t
	err := GetError(C.bladerf_get_correction(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_correction(correction),
		&correctionValue),
	)

	if err != nil {
		return 0, err
	}

	return uint16(correctionValue), nil
}

func (backend *Backend) String() string {
	return C.GoString(C.bladerf_backend_str(C.bladerf_backend(*backend)))
}

func (bladeRF *BladeRF) GetBoardName() string {
	return C.GoString(C.bladerf_get_board_name(bladeRF.ref))
}

func SetUSBResetOnOpen(enabled bool) {
	C.bladerf_set_usb_reset_on_open(C.bool(enabled))
}

func (bladeRF *BladeRF) GetSerial() (string, error) {
	var serial C.char
	err := GetError(C.bladerf_get_serial(bladeRF.ref, &serial))

	if err != nil {
		return "", err
	}

	return C.GoString(&serial), nil
}

func (bladeRF *BladeRF) GetSerialStruct() (Serial, error) {
	var serial C.struct_bladerf_serial
	err := GetError(C.bladerf_get_serial_struct(bladeRF.ref, &serial))

	if err != nil {
		return Serial{}, err
	}

	return NewSerial(&serial), nil
}

func (bladeRF *BladeRF) GetGainStages(channel Channel) ([]string, error) {
	var stage *C.char
	var stages []string

	numberOfGainStages, err := bladeRF.GetNumberOfGainStages(channel)

	if err != nil {
		return nil, err
	}

	countOrCode := C.bladerf_get_gain_stages(
		bladeRF.ref,
		C.bladerf_channel(channel),
		&stage,
		C.ulong(numberOfGainStages),
	)

	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	if countOrCode == 0 {
		return make([]string, 0), nil
	}

	count := int(countOrCode)

	if count > 0 {
		size := unsafe.Sizeof(*stage)

		for i := 0; i < count; i++ {
			stages = append(stages, C.GoString(
				(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(stage))+(uintptr(i)*size))),
			))
		}
	}

	return stages, nil
}

func (bladeRF *BladeRF) GetGainModes(channel Channel) ([]GainModes, error) {
	var gainMode *C.struct_bladerf_gain_modes
	var gainModes []GainModes

	countOrCode := C.bladerf_get_gain_modes(bladeRF.ref, C.bladerf_channel(channel), &gainMode)

	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	if countOrCode == 0 {
		return make([]GainModes, 0), nil
	}

	count := int(countOrCode)

	if count > 0 {
		size := unsafe.Sizeof(*gainMode)

		for i := 0; i < count; i++ {
			gainModes = append(gainModes, NewGainModes(
				(*C.struct_bladerf_gain_modes)(unsafe.Pointer(uintptr(unsafe.Pointer(gainMode))+(uintptr(i)*size))),
			))
		}
	}

	return gainModes, nil
}

func (bladeRF *BladeRF) GetLoopbackModes() ([]LoopbackModes, error) {
	var loopbackMode *C.struct_bladerf_loopback_modes
	var loopbackModes []LoopbackModes

	countOrCode := C.bladerf_get_loopback_modes(bladeRF.ref, &loopbackMode)

	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	if countOrCode == 0 {
		return make([]LoopbackModes, 0), nil
	}

	count := int(countOrCode)

	if count > 0 {
		size := unsafe.Sizeof(*loopbackMode)

		for i := 0; i < count; i++ {
			loopbackModes = append(loopbackModes, NewLoopbackModes(
				(*C.struct_bladerf_loopback_modes)(
					unsafe.Pointer(uintptr(unsafe.Pointer(loopbackMode))+(uintptr(i)*size)),
				),
			))
		}
	}

	return loopbackModes, nil
}

func (bladeRF *BladeRF) SetGainMode(channel Channel, mode GainMode) error {
	return GetError(C.bladerf_set_gain_mode(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_gain_mode(mode)))
}

func (bladeRF *BladeRF) EnableModule(channel Channel) error {
	return GetError(C.bladerf_enable_module(bladeRF.ref, C.bladerf_channel(channel), true))
}

func (bladeRF *BladeRF) DisableModule(channel Channel) error {
	return GetError(C.bladerf_enable_module(bladeRF.ref, C.bladerf_channel(channel), false))
}

func (bladeRF *BladeRF) TriggerInit(channel Channel, signal TriggerSignal) (Trigger, error) {
	var trigger C.struct_bladerf_trigger
	err := GetError(C.bladerf_trigger_init(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_trigger_signal(signal),
		&trigger),
	)

	if err != nil {
		return Trigger{}, err
	}

	return Trigger{ref: &trigger}, nil
}

func (bladeRF *BladeRF) TriggerArm(trigger Trigger, arm bool, resV1 uint64, resV2 uint64) error {
	return GetError(C.bladerf_trigger_arm(bladeRF.ref, trigger.ref, C.bool(arm), C.uint64_t(resV1), C.uint64_t(resV2)))
}

func (bladeRF *BladeRF) TriggerFire(trigger Trigger) error {
	return GetError(C.bladerf_trigger_fire(bladeRF.ref, trigger.ref))
}

func (bladeRF *BladeRF) TriggerState(trigger Trigger) (bool, bool, bool, uint64, uint64, error) {
	var isArmed C.bool
	var hasFired C.bool
	var fireRequested C.bool
	var resV1 C.uint64_t
	var resV2 C.uint64_t

	err := GetError(C.bladerf_trigger_state(
		bladeRF.ref,
		trigger.ref,
		&isArmed,
		&hasFired,
		&fireRequested,
		&resV1,
		&resV2),
	)

	if err != nil {
		return false, false, false, 0, 0, err
	}

	return bool(isArmed), bool(hasFired), bool(fireRequested), uint64(resV1), uint64(resV2), nil
}

func (trigger *Trigger) SetRole(role TriggerRole) {
	(*trigger.ref).role = C.bladerf_trigger_role(role)
}

func (bladeRF *BladeRF) SyncRX(bufferSize uintptr) ([]int16, error) {
	var metadata C.struct_bladerf_metadata
	start := C.malloc(C.size_t(C.sizeof_int16_t * bufferSize * 2 * 1))
	err := GetError(C.bladerf_sync_rx(bladeRF.ref, start, C.uint(bufferSize), &metadata, 32))

	if err != nil {
		return nil, err
	}

	var results []int16

	for i := 0; i < (int(metadata.actual_count)); i++ {
		n := (*C.int16_t)(unsafe.Pointer(uintptr(start) + (C.sizeof_int16_t * uintptr(i))))
		results = append(results, int16(*n))
	}

	return results, nil
}

func (bladeRF *BladeRF) InitStream(
	format Format,
	numBuffers int,
	samplesPerBuffer int,
	numTransfers int,
	callback func(data []int16),
) (Stream, error) {
	var buffers *unsafe.Pointer
	var rxStream *C.struct_bladerf_stream

	stream := Stream{ref: rxStream}

	err := GetError(C.bladerf_init_stream(
		&((stream).ref),
		bladeRF.ref,
		(*[0]byte)((C.StreamCallback)),
		&buffers,
		C.ulong(numBuffers),
		C.bladerf_format(format),
		C.ulong(samplesPerBuffer),
		C.ulong(numTransfers),
		pointer.Save(NewUserData(callback, samplesPerBuffer)),
	))

	if err != nil {
		return Stream{}, err
	}

	return stream, nil
}

func (stream *Stream) DeInit() {
	C.bladerf_deinit_stream(stream.ref)
}

func (bladeRF *BladeRF) GetStreamTimeout(direction Direction) (uint, error) {
	var timeout C.uint
	err := GetError(C.bladerf_get_stream_timeout(bladeRF.ref, C.bladerf_direction(direction), &timeout))

	if err != nil {
		return 0, err
	}

	return uint(timeout), err
}

func (bladeRF *BladeRF) SetStreamTimeout(direction Direction, timeout uint) error {
	return GetError(C.bladerf_set_stream_timeout(bladeRF.ref, C.bladerf_direction(direction), C.uint(timeout)))
}

func (bladeRF *BladeRF) SyncConfig(
	layout ChannelLayout,
	format Format,
	numBuffers uint,
	bufferSize uint,
	numTransfers uint,
	timeout uint,
) error {
	return GetError(C.bladerf_sync_config(
		bladeRF.ref,
		C.bladerf_channel_layout(layout),
		C.bladerf_format(format),
		C.uint(numBuffers),
		C.uint(bufferSize),
		C.uint(numTransfers),
		C.uint(timeout)),
	)
}

func (stream *Stream) Start(layout ChannelLayout) error {
	return GetError(C.bladerf_stream(stream.ref, C.bladerf_channel_layout(layout)))
}

func (bladeRF *BladeRF) AttachExpansionBoard(expansionBoard ExpansionBoard) error {
	return GetError(C.bladerf_expansion_attach(bladeRF.ref, C.bladerf_xb(expansionBoard)))
}

func (bladeRF *BladeRF) GetAttachedExpansionBoard() (ExpansionBoard, error) {
	var expansionBoard C.bladerf_xb
	err := GetError(C.bladerf_expansion_get_attached(bladeRF.ref, &expansionBoard))

	if err != nil {
		return 0, err
	}

	return ExpansionBoard(expansionBoard), nil
}

func (bladeRF *BladeRF) SetGetVctcxoTamerMode(mode VctcxoTamerMode) error {
	return GetError(C.bladerf_set_vctcxo_tamer_mode(bladeRF.ref, C.bladerf_vctcxo_tamer_mode(mode)))
}

func (bladeRF *BladeRF) GetVctcxoTamerMode() (VctcxoTamerMode, error) {
	var mode C.bladerf_vctcxo_tamer_mode
	err := GetError(C.bladerf_get_vctcxo_tamer_mode(bladeRF.ref, &mode))

	if err != nil {
		return 0, err
	}

	return VctcxoTamerMode(mode), nil
}

func (bladeRF *BladeRF) GetVctcxoTrim() (uint16, error) {
	var trim C.uint16_t
	err := GetError(C.bladerf_get_vctcxo_trim(bladeRF.ref, &trim))

	if err != nil {
		return 0, err
	}

	return uint16(trim), nil
}

func (bladeRF *BladeRF) TrimDacRead() (uint16, error) {
	var val C.uint16_t
	err := GetError(C.bladerf_trim_dac_read(bladeRF.ref, &val))

	if err != nil {
		return 0, err
	}

	return uint16(val), nil
}

func (bladeRF *BladeRF) TrimDacWrite(val uint16) error {
	return GetError(C.bladerf_trim_dac_write(bladeRF.ref, C.uint16_t(val)))
}
