package bladerf

// #include <stdlib.h>
// #include <libbladeRF.h>
import "C"
import "unsafe"

func (bladeRF *BladeRF) SetGain(channel Channel, gain int) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_gain(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_gain(gain)))
}

func (bladeRF *BladeRF) GetGain(channel Channel) (int, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var gain C.bladerf_gain
	if err := GetError(C.bladerf_get_gain(bladeRF.ref, C.bladerf_channel(channel), &gain)); err != nil {
		return 0, err
	}

	return int(gain), nil
}

func (bladeRF *BladeRF) GetGainRange(channel Channel) (Range, error) {
	if !bladeRF.valid() {
		return Range{}, invalErr()
	}
	var bfRange *C.struct_bladerf_range
	if err := GetError(C.bladerf_get_gain_range(bladeRF.ref, C.bladerf_channel(channel), &bfRange)); err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) SetGainMode(channel Channel, mode GainMode) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_gain_mode(bladeRF.ref, C.bladerf_channel(channel), C.bladerf_gain_mode(mode)))
}

func (bladeRF *BladeRF) GetGainMode(channel Channel) (GainMode, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var mode C.bladerf_gain_mode
	if err := GetError(C.bladerf_get_gain_mode(bladeRF.ref, C.bladerf_channel(channel), &mode)); err != nil {
		return 0, err
	}

	return GainMode(mode), nil
}

func (bladeRF *BladeRF) GetGainModes(channel Channel) ([]GainModes, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	var gainMode *C.struct_bladerf_gain_modes
	countOrCode := C.bladerf_get_gain_modes(bladeRF.ref, C.bladerf_channel(channel), &gainMode)
	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	count := int(countOrCode)
	if count == 0 {
		return make([]GainModes, 0), nil
	}

	cModes := (*[1 << 28]C.struct_bladerf_gain_modes)(unsafe.Pointer(gainMode))[:count:count]
	gainModes := make([]GainModes, count)
	for i := 0; i < count; i++ {
		gainModes[i] = NewGainModes(&cModes[i])
	}

	return gainModes, nil
}

func (bladeRF *BladeRF) SetGainStage(channel Channel, stage string, gain int) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	cStage := C.CString(stage)
	defer C.free(unsafe.Pointer(cStage))

	return GetError(C.bladerf_set_gain_stage(bladeRF.ref, C.bladerf_channel(channel), cStage, C.bladerf_gain(gain)))
}

func (bladeRF *BladeRF) GetGainStage(channel Channel, stage string) (int, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	cStage := C.CString(stage)
	defer C.free(unsafe.Pointer(cStage))

	var gain C.bladerf_gain
	if err := GetError(C.bladerf_get_gain_stage(bladeRF.ref, C.bladerf_channel(channel), cStage, &gain)); err != nil {
		return 0, err
	}

	return int(gain), nil
}

func (bladeRF *BladeRF) GetGainStageRange(channel Channel, stage string) (Range, error) {
	if !bladeRF.valid() {
		return Range{}, invalErr()
	}
	cStage := C.CString(stage)
	defer C.free(unsafe.Pointer(cStage))

	var bfRange *C.struct_bladerf_range
	if err := GetError(C.bladerf_get_gain_stage_range(bladeRF.ref, C.bladerf_channel(channel), cStage, &bfRange)); err != nil {
		return Range{}, err
	}

	return NewRange(bfRange), nil
}

func (bladeRF *BladeRF) GetNumberOfGainStages(channel Channel) (int, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	countOrCode := C.bladerf_get_gain_stages(bladeRF.ref, C.bladerf_channel(channel), nil, 0)
	if countOrCode < 0 {
		return 0, GetError(countOrCode)
	}

	return int(countOrCode), nil
}

func (bladeRF *BladeRF) GetGainStages(channel Channel) ([]string, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	numberOfGainStages, err := bladeRF.GetNumberOfGainStages(channel)
	if err != nil {
		return nil, err
	}
	if numberOfGainStages == 0 {
		return make([]string, 0), nil
	}

	raw := C.calloc(C.size_t(numberOfGainStages), C.size_t(unsafe.Sizeof((*C.char)(nil))))
	defer C.free(raw)

	countOrCode := C.bladerf_get_gain_stages(
		bladeRF.ref,
		C.bladerf_channel(channel),
		(**C.char)(raw),
		C.size_t(numberOfGainStages),
	)
	if countOrCode < 0 {
		return nil, GetError(countOrCode)
	}

	count := int(countOrCode)
	if count > numberOfGainStages {
		count = numberOfGainStages
	}

	return cStringArrayToGo(raw, count), nil
}

func (bladeRF *BladeRF) SetCorrection(channel Channel, correction Correction, correctionValue int16) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_correction(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_correction(correction),
		C.bladerf_correction_value(correctionValue),
	))
}

func (bladeRF *BladeRF) GetCorrection(channel Channel, correction Correction) (int16, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var correctionValue C.int16_t
	err := GetError(C.bladerf_get_correction(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_correction(correction),
		&correctionValue,
	))
	if err != nil {
		return 0, err
	}

	return int16(correctionValue), nil
}
