package bladerf

// #include <stdlib.h>
// #include <libbladeRF.h>
import "C"
import (
	"unsafe"

	exception "github.com/erayarslan/go-bladerf/error"
)

func (s FpgaSize) String() string {
	switch s {
	case FpgaSize40kle:
		return "x40"
	case FpgaSize115kle:
		return "x115"
	case FpgaSizeA4:
		return "xA4"
	case FpgaSizeA9:
		return "xA9"
	default:
		return "unknown"
	}
}

func (s FpgaSource) String() string {
	switch s {
	case FpgaSourceFlash:
		return "flash"
	case FpgaSourceHost:
		return "host"
	default:
		return "unknown"
	}
}

func (s DeviceSpeed) String() string {
	switch s {
	case SpeedSuper:
		return "SuperSpeed (USB 3.0)"
	case SpeedHigh:
		return "HighSpeed (USB 2.0)"
	default:
		return "unknown"
	}
}

func (bladeRF *BladeRF) LoadFpga(imagePath string) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))
	return GetError(C.bladerf_load_fpga(bladeRF.ref, path))
}

func (bladeRF *BladeRF) FlashFpga(imagePath string) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))
	return GetError(C.bladerf_flash_fpga(bladeRF.ref, path))
}

func (bladeRF *BladeRF) EraseStoredFpga() error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_erase_stored_fpga(bladeRF.ref))
}

func (bladeRF *BladeRF) FlashFirmware(imagePath string) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	path := C.CString(imagePath)
	defer C.free(unsafe.Pointer(path))
	return GetError(C.bladerf_flash_firmware(bladeRF.ref, path))
}

func (bladeRF *BladeRF) DeviceReset() error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_device_reset(bladeRF.ref))
}

func LoadFirmwareFromBootloader(deviceIdentifier string, backend Backend, bus, addr uint8, imagePath string) error {
	var cID *C.char
	if deviceIdentifier != "" {
		cID = C.CString(deviceIdentifier)
		defer C.free(unsafe.Pointer(cID))
	}
	cFile := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cFile))
	return GetError(C.bladerf_load_fw_from_bootloader(cID, C.bladerf_backend(backend), C.uint8_t(bus), C.uint8_t(addr), cFile))
}

func (bladeRF *BladeRF) IsFpgaConfigured() (bool, error) {
	if !bladeRF.valid() {
		return false, invalErr()
	}
	out := C.bladerf_is_fpga_configured(bladeRF.ref)
	if out < 0 {
		return false, GetError(out)
	}

	return out == 1, nil
}

func (bladeRF *BladeRF) GetFpgaSize() (FpgaSize, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var size C.bladerf_fpga_size
	if err := GetError(C.bladerf_get_fpga_size(bladeRF.ref, &size)); err != nil {
		return 0, err
	}

	return FpgaSize(size), nil
}

func (bladeRF *BladeRF) GetFpgaSource() (FpgaSource, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var source C.bladerf_fpga_source
	if err := GetError(C.bladerf_get_fpga_source(bladeRF.ref, &source)); err != nil {
		return 0, err
	}

	return FpgaSource(source), nil
}

func (bladeRF *BladeRF) GetFpgaBytes() (uint32, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var size C.size_t
	if err := GetError(C.bladerf_get_fpga_bytes(bladeRF.ref, &size)); err != nil {
		return 0, err
	}

	return uint32(size), nil
}

func (bladeRF *BladeRF) GetFpgaFlashSize() (uint32, bool, error) {
	if !bladeRF.valid() {
		return 0, false, invalErr()
	}
	var size C.uint32_t
	var isGuess C.bool
	if err := GetError(C.bladerf_get_flash_size(bladeRF.ref, &size, &isGuess)); err != nil {
		return 0, false, err
	}

	return uint32(size), bool(isGuess), nil
}

func (bladeRF *BladeRF) GetFirmwareVersion() (Version, error) {
	if !bladeRF.valid() {
		return Version{}, invalErr()
	}
	var version C.struct_bladerf_version
	if err := GetError(C.bladerf_fw_version(bladeRF.ref, &version)); err != nil {
		return Version{}, err
	}

	return NewVersion(&version), nil
}

func (bladeRF *BladeRF) GetFpgaVersion() (Version, error) {
	if !bladeRF.valid() {
		return Version{}, invalErr()
	}
	var version C.struct_bladerf_version
	if err := GetError(C.bladerf_fpga_version(bladeRF.ref, &version)); err != nil {
		return Version{}, err
	}

	return NewVersion(&version), nil
}

func (bladeRF *BladeRF) EraseFlash(eraseBlock uint32, count uint32) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_erase_flash(bladeRF.ref, C.uint32_t(eraseBlock), C.uint32_t(count)))
}

func (bladeRF *BladeRF) EraseFlashBytes(address uint32, length uint32) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_erase_flash_bytes(bladeRF.ref, C.uint32_t(address), C.uint32_t(length)))
}

func (bladeRF *BladeRF) LockOtp() error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_lock_otp(bladeRF.ref))
}

func (bladeRF *BladeRF) ReadFlashBytes(address uint32, bytes uint32) ([]uint8, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	if bytes == 0 {
		return make([]uint8, 0), nil
	}

	output := make([]uint8, bytes)
	err := GetError(C.bladerf_read_flash_bytes(
		bladeRF.ref,
		(*C.uint8_t)(unsafe.Pointer(&output[0])),
		C.uint32_t(address),
		C.uint32_t(bytes),
	))
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (bladeRF *BladeRF) WriteFlashBytes(input []uint8, address uint32, bytes uint32) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	if uint32(len(input)) < bytes {
		return exception.New(int(exception.Inval))
	}
	if bytes == 0 {
		return nil
	}

	return GetError(C.bladerf_write_flash_bytes(
		bladeRF.ref,
		(*C.uint8_t)(unsafe.Pointer(&input[0])),
		C.uint32_t(address),
		C.uint32_t(bytes),
	))
}

func (bladeRF *BladeRF) ReadOtp() ([]uint8, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	output := make([]uint8, OtpSize)
	if err := GetError(C.bladerf_read_otp(bladeRF.ref, (*C.uint8_t)(unsafe.Pointer(&output[0])))); err != nil {
		return nil, err
	}

	return output, nil
}

func (bladeRF *BladeRF) WriteOtp(input []uint8) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	if len(input) < OtpSize {
		return exception.New(int(exception.Inval))
	}

	return GetError(C.bladerf_write_otp(bladeRF.ref, (*C.uint8_t)(unsafe.Pointer(&input[0]))))
}

func (bladeRF *BladeRF) ReadFlash(page uint32, count uint32) ([]uint8, error) {
	if !bladeRF.valid() {
		return nil, invalErr()
	}
	totalBytes := uint64(count) * FlashPageSize
	if totalBytes == 0 {
		return make([]uint8, 0), nil
	}

	output := make([]uint8, totalBytes)
	err := GetError(C.bladerf_read_flash(
		bladeRF.ref,
		(*C.uint8_t)(unsafe.Pointer(&output[0])),
		C.uint32_t(page),
		C.uint32_t(count),
	))
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (bladeRF *BladeRF) WriteFlash(input []uint8, page uint32, count uint32) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	totalBytes := uint64(count) * FlashPageSize
	if uint64(len(input)) < totalBytes {
		return exception.New(int(exception.Inval))
	}
	if totalBytes == 0 {
		return nil
	}

	return GetError(C.bladerf_write_flash(
		bladeRF.ref,
		(*C.uint8_t)(unsafe.Pointer(&input[0])),
		C.uint32_t(page),
		C.uint32_t(count),
	))
}
