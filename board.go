package bladerf

// #include <libbladeRF.h>
import "C"

func (bladeRF *BladeRF) TriggerInit(channel Channel, signal TriggerSignal) (Trigger, error) {
	if !bladeRF.valid() {
		return Trigger{}, invalErr()
	}
	var trigger C.struct_bladerf_trigger
	err := GetError(C.bladerf_trigger_init(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_trigger_signal(signal),
		&trigger,
	))
	if err != nil {
		return Trigger{}, err
	}

	return Trigger{ref: &trigger}, nil
}

func (bladeRF *BladeRF) TriggerArm(trigger Trigger, arm bool, resV1 uint64, resV2 uint64) error {
	if !bladeRF.valid() || trigger.ref == nil {
		return invalErr()
	}
	return GetError(C.bladerf_trigger_arm(
		bladeRF.ref,
		trigger.ref,
		C.bool(arm),
		C.uint64_t(resV1),
		C.uint64_t(resV2),
	))
}

func (bladeRF *BladeRF) TriggerFire(trigger Trigger) error {
	if !bladeRF.valid() || trigger.ref == nil {
		return invalErr()
	}
	return GetError(C.bladerf_trigger_fire(bladeRF.ref, trigger.ref))
}

func (bladeRF *BladeRF) TriggerState(trigger Trigger) (bool, bool, bool, uint64, uint64, error) {
	if !bladeRF.valid() || trigger.ref == nil {
		return false, false, false, 0, 0, invalErr()
	}
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
		&resV2,
	))
	if err != nil {
		return false, false, false, 0, 0, err
	}

	return bool(isArmed), bool(hasFired), bool(fireRequested), uint64(resV1), uint64(resV2), nil
}

func (trigger *Trigger) SetRole(role TriggerRole) {
	if trigger == nil || trigger.ref == nil {
		return
	}

	trigger.ref.role = C.bladerf_trigger_role(role)
}

func (bladeRF *BladeRF) ReadTrigger(channel Channel, signal TriggerSignal) (uint8, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var val C.uint8_t
	err := GetError(C.bladerf_read_trigger(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_trigger_signal(signal),
		&val,
	))
	if err != nil {
		return 0, err
	}

	return uint8(val), nil
}

func (bladeRF *BladeRF) WriteTrigger(channel Channel, signal TriggerSignal, val uint8) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_write_trigger(
		bladeRF.ref,
		C.bladerf_channel(channel),
		C.bladerf_trigger_signal(signal),
		C.uint8_t(val),
	))
}

func (bladeRF *BladeRF) AttachExpansionBoard(expansionBoard ExpansionBoard) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_expansion_attach(bladeRF.ref, C.bladerf_xb(expansionBoard)))
}

func (bladeRF *BladeRF) GetAttachedExpansionBoard() (ExpansionBoard, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var expansionBoard C.bladerf_xb
	if err := GetError(C.bladerf_expansion_get_attached(bladeRF.ref, &expansionBoard)); err != nil {
		return 0, err
	}

	return ExpansionBoard(expansionBoard), nil
}

func (bladeRF *BladeRF) SetVctcxoTamerMode(mode VctcxoTamerMode) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_vctcxo_tamer_mode(bladeRF.ref, C.bladerf_vctcxo_tamer_mode(mode)))
}

func (bladeRF *BladeRF) GetVctcxoTamerMode() (VctcxoTamerMode, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var mode C.bladerf_vctcxo_tamer_mode
	if err := GetError(C.bladerf_get_vctcxo_tamer_mode(bladeRF.ref, &mode)); err != nil {
		return 0, err
	}

	return VctcxoTamerMode(mode), nil
}

func (bladeRF *BladeRF) GetVctcxoTrim() (uint16, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var trim C.uint16_t
	if err := GetError(C.bladerf_get_vctcxo_trim(bladeRF.ref, &trim)); err != nil {
		return 0, err
	}

	return uint16(trim), nil
}

func (bladeRF *BladeRF) TrimDacRead() (uint16, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var val C.uint16_t
	if err := GetError(C.bladerf_trim_dac_read(bladeRF.ref, &val)); err != nil {
		return 0, err
	}

	return uint16(val), nil
}

func (bladeRF *BladeRF) TrimDacWrite(val uint16) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_trim_dac_write(bladeRF.ref, C.uint16_t(val)))
}

func (bladeRF *BladeRF) ConfigGpioRead() (uint32, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var val C.uint32_t
	if err := GetError(C.bladerf_config_gpio_read(bladeRF.ref, &val)); err != nil {
		return 0, err
	}

	return uint32(val), nil
}

func (bladeRF *BladeRF) ConfigGpioWrite(val uint32) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_config_gpio_write(bladeRF.ref, C.uint32_t(val)))
}
