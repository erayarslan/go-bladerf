package bladerf

// #cgo CFLAGS: -I${SRCDIR}/build/dist/include
// #cgo linux LDFLAGS: -L${SRCDIR}/build/dist/lib -Wl,-Bstatic -lbladeRF -Wl,-Bdynamic -ldl -lm -lpthread
// #cgo darwin CFLAGS: -I/usr/local/include
// #cgo darwin LDFLAGS: -L${SRCDIR}/build/dist/lib -L/usr/local/lib -lbladeRF -lusb-1.0 -lm -lpthread
// #include <stdlib.h>
// #include <libbladeRF.h>
import "C"
import (
	"fmt"
	"unsafe"

	exception "github.com/erayarslan/go-bladerf/error"
)

func GetError(code C.int) error {
	return exception.New(int(code))
}

func (bladeRF *BladeRF) valid() bool {
	return bladeRF != nil && bladeRF.ref != nil
}

func invalErr() error {
	return exception.New(int(exception.Inval))
}

func GetVersion() Version {
	var version C.struct_bladerf_version
	C.bladerf_version(&version)
	return NewVersion(&version)
}

func (version Version) String() string {
	return fmt.Sprintf("v%d.%d.%d (\"%s\")", version.Major, version.Minor, version.Patch, version.Describe)
}

func (version Version) Print() {
	fmt.Print(version.String())
}

func (backend Backend) String() string {
	return C.GoString(C.bladerf_backend_str(C.bladerf_backend(backend)))
}

func SetUSBResetOnOpen(enabled bool) {
	C.bladerf_set_usb_reset_on_open(C.bool(enabled))
}

func deviceListFromC(list *C.struct_bladerf_devinfo, codeOrCount C.int) ([]DeviceInfo, error) {
	if list != nil {
		defer C.bladerf_free_device_list(list)
	}
	if codeOrCount < 0 {
		return nil, GetError(codeOrCount)
	}

	count := int(codeOrCount)
	if count == 0 {
		return make([]DeviceInfo, 0), nil
	}

	cDevices := (*[1 << 28]C.struct_bladerf_devinfo)(unsafe.Pointer(list))[:count:count]
	devices := make([]DeviceInfo, count)
	for i := 0; i < count; i++ {
		devices[i] = NewDeviceInfo(&cDevices[i])
	}

	return devices, nil
}

func GetDeviceList() ([]DeviceInfo, error) {
	var list *C.struct_bladerf_devinfo
	codeOrCount := C.bladerf_get_device_list(&list)
	return deviceListFromC(list, codeOrCount)
}

func GetBootloaderList() ([]DeviceInfo, error) {
	var list *C.struct_bladerf_devinfo
	codeOrCount := C.bladerf_get_bootloader_list(&list)
	return deviceListFromC(list, codeOrCount)
}

func (deviceInfo *DeviceInfo) FreeDeviceList() {
	if deviceInfo == nil || deviceInfo.listRef == nil {
		return
	}

	C.bladerf_free_device_list(deviceInfo.listRef)
	deviceInfo.listRef = nil
}

func InitDeviceInfo() DeviceInfo {
	var deviceInfo C.struct_bladerf_devinfo
	C.bladerf_init_devinfo(&deviceInfo)
	return NewDeviceInfo(&deviceInfo)
}

func GetDeviceInfoFromString(deviceString string) (DeviceInfo, error) {
	val := C.CString(deviceString)
	defer C.free(unsafe.Pointer(val))

	var deviceInfo C.struct_bladerf_devinfo
	if err := GetError(C.bladerf_get_devinfo_from_str(val, &deviceInfo)); err != nil {
		return DeviceInfo{}, err
	}

	return NewDeviceInfo(&deviceInfo), nil
}

func (deviceInfo *DeviceInfo) DeviceInfoMatches(target DeviceInfo) bool {
	a := deviceInfo.cStruct()
	b := target.cStruct()
	return bool(C.bladerf_devinfo_matches(&a, &b))
}

func (deviceInfo *DeviceInfo) DeviceStringMatches(deviceString string) bool {
	val := C.CString(deviceString)
	defer C.free(unsafe.Pointer(val))

	info := deviceInfo.cStruct()
	return bool(C.bladerf_devstr_matches(val, &info))
}

func (deviceInfo *DeviceInfo) Open() (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	info := deviceInfo.cStruct()
	if err := GetError(C.bladerf_open_with_devinfo(&bladeRF, &info)); err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func OpenWithDeviceIdentifier(identify string) (BladeRF, error) {
	cIdentify := C.CString(identify)
	defer C.free(unsafe.Pointer(cIdentify))

	var bladeRF *C.struct_bladerf
	if err := GetError(C.bladerf_open(&bladeRF, cIdentify)); err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func Open() (BladeRF, error) {
	var bladeRF *C.struct_bladerf
	if err := GetError(C.bladerf_open(&bladeRF, nil)); err != nil {
		return BladeRF{}, err
	}

	return BladeRF{ref: bladeRF}, nil
}

func OpenRDMA(addr string) (BladeRF, error) {
	return OpenWithDeviceIdentifier("rdma:addr=" + addr)
}

func OpenTCP(addr string) (BladeRF, error) {
	return OpenWithDeviceIdentifier("tcp:addr=" + addr)
}

func (bladeRF *BladeRF) Close() {
	if bladeRF == nil || bladeRF.ref == nil {
		return
	}

	C.bladerf_close(bladeRF.ref)
	bladeRF.ref = nil
}

func (bladeRF *BladeRF) GetDeviceInfo() (DeviceInfo, error) {
	if !bladeRF.valid() {
		return DeviceInfo{}, invalErr()
	}
	var deviceInfo C.struct_bladerf_devinfo
	if err := GetError(C.bladerf_get_devinfo(bladeRF.ref, &deviceInfo)); err != nil {
		return DeviceInfo{}, err
	}

	return NewDeviceInfo(&deviceInfo), nil
}

func (bladeRF *BladeRF) GetDeviceSpeed() DeviceSpeed {
	if !bladeRF.valid() {
		return SpeedUnknown
	}
	return DeviceSpeed(C.bladerf_device_speed(bladeRF.ref))
}

func (bladeRF *BladeRF) GetBoardName() string {
	if !bladeRF.valid() {
		return ""
	}
	return C.GoString(C.bladerf_get_board_name(bladeRF.ref))
}

func (bladeRF *BladeRF) GetSerial() (string, error) {
	if !bladeRF.valid() {
		return "", invalErr()
	}
	var serial [C.BLADERF_SERIAL_LENGTH]C.char
	if err := GetError(C.bladerf_get_serial(bladeRF.ref, &serial[0])); err != nil {
		return "", err
	}

	return C.GoString(&serial[0]), nil
}

func (bladeRF *BladeRF) GetSerialStruct() (Serial, error) {
	if !bladeRF.valid() {
		return Serial{}, invalErr()
	}
	var serial C.struct_bladerf_serial
	if err := GetError(C.bladerf_get_serial_struct(bladeRF.ref, &serial)); err != nil {
		return Serial{}, err
	}

	return NewSerial(&serial), nil
}
