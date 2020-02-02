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
	StartStream(stream, IOTX)
	Close(rf)
}

func TestStream(t *testing.T) {
	log.SetVerbosity(log.Debug)

	f_low := 250000000
	f_high := 700000000
	f_step := 1000000

	freq := f_low

	tx_gain := 56
	rx_gain := 3

	num_buffers := 24
	samples_per_buffer := 8192
	num_transfers := 8

	sample_rate := 4000000
	bandwidth := 1500000

	n_steps := 1 + (f_high-f_low)/f_step

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	SetLoopback(&rf, Disabled)

	SetFrequency(&rf, IOTX, freq)
	SetFrequency(&rf, IORX, freq)

	SetSampleRate(&rf, IOTX, sample_rate)
	SetSampleRate(&rf, IORX, sample_rate)

	SetBandwidth(&rf, IOTX, bandwidth)
	SetBandwidth(&rf, IORX, bandwidth)

	SetGain(&rf, IOTX, tx_gain)
	SetGain(&rf, IORX, rx_gain)

	EnableModule(&rf, IOTX)
	EnableModule(&rf, IORX)

	for i := 0; i < n_steps; i++ {

		//tx_samples_left := num_buffers * samples_per_buffer
		//rx_samples_left := num_buffers * samples_per_buffer

		tx_stream := InitStream(&rf, SC16_Q11, num_buffers, samples_per_buffer, num_transfers)
		rx_stream := InitStream(&rf, SC16_Q11, num_buffers, samples_per_buffer, num_transfers)

		StartStream(tx_stream, IOTX)
		StartStream(rx_stream, IORX)
	}
}
