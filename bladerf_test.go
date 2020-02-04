package bladerf

import (
	"bladerf/log"
	"fmt"
	"testing"
)

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Debug)

	PrintVersion(GetVersion())

	LoadFpga(Open(), "~/test") // Auto Close
	Close(Open())
	Close(OpenWithDevInfo(GetDevInfo()))
	Close(OpenWithDeviceIdentifier("*:serial=NOTFOUND"))

	devices := GetDeviceList()
	fmt.Printf("Devices Len: %d\n", len(devices))

	bootloaders := GetBootloaderList()
	fmt.Printf("Bootloaders Len: %d\n", len(bootloaders))

	rf := Open()
	stream := InitStream(&rf, SC16_Q11, 2, 1024, 1)
	StartStream(stream, RX_X1)
	Close(rf)
}

func TestStream(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	SyncConfig(&rf, RX_X1, SC16_Q11, 32, 32768, 16, 16000)
	EnableModule(&rf, RX)
	SyncRX(&rf)
}

func TestAsyncStream(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	SetSampleRate(&rf, IORX, 1000000)
	EnableModule(&rf, RX)

	rxStream := InitStream(&rf, SC16_Q11, 32, 32768, 16)
	defer DeInitStream(rxStream)

	SetStreamTimeout(&rf, RX, 16)
	GetStreamTimeout(&rf, RX)

	StartStream(rxStream, RX_X1)
}
