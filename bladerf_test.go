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
}
