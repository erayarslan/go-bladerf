package bladerf

// #include <stdlib.h>
// #include <libbladeRF.h>
import "C"
import "unsafe"

func cStringArrayToGo(raw unsafe.Pointer, count int) []string {
	if count <= 0 || raw == nil {
		return make([]string, 0)
	}

	ptrs := (*[1 << 28]*C.char)(raw)[:count:count]
	out := make([]string, count)
	for i := 0; i < count; i++ {
		if ptrs[i] != nil {
			out[i] = C.GoString(ptrs[i])
		}
	}

	return out
}

func (bladeRF *BladeRF) EnableModule(channel Channel) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_enable_module(bladeRF.ref, C.bladerf_channel(channel), true))
}

func (bladeRF *BladeRF) DisableModule(channel Channel) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_enable_module(bladeRF.ref, C.bladerf_channel(channel), false))
}

func (bladeRF *BladeRF) SelectBand(channel Channel, frequency uint64) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_select_band(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_frequency(frequency)))
}

func (bladeRF *BladeRF) SetFrequency(channel Channel, frequency uint64) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_frequency(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_frequency(frequency)))
}

func (bladeRF *BladeRF) GetFrequency(channel Channel) (uint64, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var frequency C.uint64_t
	if err := GetError(C.bladerf_get_frequency(bladeRF.ref, C.bladerf_channel(channel), &frequency)); err != nil {
		return 0, err
	}

	return uint64(frequency), nil
}

func (bladeRF *BladeRF) GetFrequencyRange(channel Channel) (Range, error) {
	if !bladeRF.valid() {
		return Range{}, invalErr()
	}
	var bfRange *C.struct_bladerf_range
	if err := GetError(C.bladerf_get_frequency_range(bladeRF.ref, C.bladerf_channel(channel), &bfRange)); err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) GetQuickTune(channel Channel) (QuickTune, error) {
	if !bladeRF.valid() {
		return QuickTune{}, invalErr()
	}
	var quickTune C.struct_bladerf_quick_tune
	if err := GetError(C.bladerf_get_quick_tune(bladeRF.ref, C.bladerf_channel(channel), &quickTune)); err != nil {
		return QuickTune{}, err
	}

	return QuickTune{ref: &quickTune}, nil
}

func (bladeRF *BladeRF) ScheduleReTune(
	channel Channel,
	timestamp Timestamp,
	frequency uint64,
	quickTune QuickTune,
) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_schedule_retune(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_timestamp(timestamp),
		C.bladerf_frequency(frequency),
		quickTune.ref,
	))
}

func (bladeRF *BladeRF) CancelScheduledReTunes(channel Channel) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_cancel_scheduled_retunes(bladeRF.ref, C.bladerf_channel(channel)))
}

func (bladeRF *BladeRF) SetTuningMode(mode TuningMode) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_tuning_mode(bladeRF.ref, C.bladerf_tuning_mode(mode)))
}

func (bladeRF *BladeRF) GetTuningMode() (TuningMode, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var mode C.bladerf_tuning_mode
	if err := GetError(C.bladerf_get_tuning_mode(bladeRF.ref, &mode)); err != nil {
		return 0, err
	}

	return TuningMode(mode), nil
}

func (bladeRF *BladeRF) SetSampleRate(channel Channel, sampleRate uint) (uint, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var actual C.uint
	err := GetError(C.bladerf_set_sample_rate(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_sample_rate(sampleRate),
		&actual,
	))
	if err != nil {
		return 0, err
	}

	return uint(actual), nil
}

func (bladeRF *BladeRF) GetSampleRate(channel Channel) (uint, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var sampleRate C.uint
	if err := GetError(C.bladerf_get_sample_rate(bladeRF.ref, C.bladerf_channel(channel), &sampleRate)); err != nil {
		return 0, err
	}

	return uint(sampleRate), nil
}

func (bladeRF *BladeRF) SetRationalSampleRate(channel Channel, rationalRate RationalRate) (RationalRate, error) {
	if !bladeRF.valid() {
		return RationalRate{}, invalErr()
	}
	var actual C.struct_bladerf_rational_rate
	cRate := rationalRate.toC()
	err := GetError(C.bladerf_set_rational_sample_rate(
		bladeRF.ref,
		C.bladerf_channel(channel),
		&cRate,
		&actual,
	))
	if err != nil {
		return RationalRate{}, err
	}

	return NewRationalRate(&actual), nil
}

func (bladeRF *BladeRF) GetRationalSampleRate(channel Channel) (RationalRate, error) {
	if !bladeRF.valid() {
		return RationalRate{}, invalErr()
	}
	var rate C.struct_bladerf_rational_rate
	if err := GetError(C.bladerf_get_rational_sample_rate(bladeRF.ref, C.bladerf_channel(channel), &rate)); err != nil {
		return RationalRate{}, err
	}

	return NewRationalRate(&rate), nil
}

func (bladeRF *BladeRF) GetSampleRateRange(channel Channel) (Range, error) {
	if !bladeRF.valid() {
		return Range{}, invalErr()
	}
	var bfRange *C.struct_bladerf_range
	if err := GetError(C.bladerf_get_sample_rate_range(bladeRF.ref, C.bladerf_channel(channel), &bfRange)); err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) SetBandwidth(channel Channel, bandwidth uint) (uint, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var actual C.bladerf_bandwidth
	err := GetError(C.bladerf_set_bandwidth(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_bandwidth(bandwidth),
		&actual,
	))
	if err != nil {
		return 0, err
	}

	return uint(actual), nil
}

func (bladeRF *BladeRF) GetBandwidth(channel Channel) (uint, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var bandwidth C.bladerf_bandwidth
	if err := GetError(C.bladerf_get_bandwidth(bladeRF.ref, C.bladerf_channel(channel), &bandwidth)); err != nil {
		return 0, err
	}

	return uint(bandwidth), nil
}

func (bladeRF *BladeRF) GetBandwidthRange(channel Channel) (Range, error) {
	if !bladeRF.valid() {
		return Range{}, invalErr()
	}
	var bfRange *C.struct_bladerf_range
	if err := GetError(C.bladerf_get_bandwidth_range(bladeRF.ref, C.bladerf_channel(channel), &bfRange)); err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) SetRxMux(mux RxMux) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_rx_mux(bladeRF.ref, C.bladerf_rx_mux(mux)))
}

func (bladeRF *BladeRF) GetRxMux() (RxMux, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var rxMux C.bladerf_rx_mux
	if err := GetError(C.bladerf_get_rx_mux(bladeRF.ref, &rxMux)); err != nil {
		return 0, err
	}

	return RxMux(rxMux), nil
}

func (bladeRF *BladeRF) SetLoopback(loopback Loopback) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_loopback(bladeRF.ref, C.bladerf_loopback(loopback)))
}

func (bladeRF *BladeRF) GetLoopback() (Loopback, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var loopback C.bladerf_loopback
	if err := GetError(C.bladerf_get_loopback(bladeRF.ref, &loopback)); err != nil {
		return 0, err
	}

	return Loopback(loopback), nil
}

func (bladeRF *BladeRF) IsLoopbackModeSupported(loopback Loopback) bool {
	if !bladeRF.valid() {
		return false
	}
	return bool(C.bladerf_is_loopback_mode_supported(bladeRF.ref, C.bladerf_loopback(loopback)))
}

func (bladeRF *BladeRF) GetLoopbackModes() ([]LoopbackModes, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	var loopbackMode *C.struct_bladerf_loopback_modes
	countOrCode := C.bladerf_get_loopback_modes(bladeRF.ref, &loopbackMode)
	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	count := int(countOrCode)
	if count == 0 {
		return make([]LoopbackModes, 0), nil
	}

	cModes := (*[1 << 28]C.struct_bladerf_loopback_modes)(unsafe.Pointer(loopbackMode))[:count:count]
	loopbackModes := make([]LoopbackModes, count)
	for i := 0; i < count; i++ {
		loopbackModes[i] = NewLoopbackModes(&cModes[i])
	}

	return loopbackModes, nil
}

func (bladeRF *BladeRF) SetRfPort(channel Channel, port string) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	cPort := C.CString(port)
	defer C.free(unsafe.Pointer(cPort))
	return GetError(C.bladerf_set_rf_port(bladeRF.ref, C.bladerf_channel(channel), cPort))
}

func (bladeRF *BladeRF) GetRfPort(channel Channel) (string, error) {
	if !bladeRF.valid() {
		return "", invalErr()
	}
	var portPtr *C.char
	if err := GetError(C.bladerf_get_rf_port(bladeRF.ref, C.bladerf_channel(channel), &portPtr)); err != nil {
		return "", err
	}

	return C.GoString(portPtr), nil
}

func (bladeRF *BladeRF) GetNumberOfRfPorts(channel Channel) (int, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	countOrCode := C.bladerf_get_rf_ports(bladeRF.ref, C.bladerf_channel(channel), nil, 0)
	if countOrCode < 0 {
		return 0, GetError(countOrCode)
	}

	return int(countOrCode), nil
}

func (bladeRF *BladeRF) GetRfPorts(channel Channel) ([]string, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	numberOfRfPorts, err := bladeRF.GetNumberOfRfPorts(channel)
	if err != nil {
		return nil, err
	}
	if numberOfRfPorts == 0 {
		return make([]string, 0), nil
	}

	raw := C.calloc(C.size_t(numberOfRfPorts), C.size_t(unsafe.Sizeof((*C.char)(nil))))
	defer C.free(raw)

	countOrCode := C.bladerf_get_rf_ports(
		bladeRF.ref,
		C.bladerf_channel(channel),
		(**C.char)(raw),
		C.uint(numberOfRfPorts),
	)
	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	count := int(countOrCode)
	if count > numberOfRfPorts {
		count = numberOfRfPorts
	}

	return cStringArrayToGo(raw, count), nil
}
