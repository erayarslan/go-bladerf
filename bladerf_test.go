package bladerf

import (
	"fmt"
	"testing"
)

var Rx1Channel = ChannelRx(0)

func TestGetVersion(t *testing.T) {
	version := GetVersion()

	if version.Major != 2 {
		t.Error("version.Major FAILED")
	} else {
		t.Log("version.Major PASSED")
	}

	if version.Minor != 2 {
		t.Error("version.Minor FAILED")
	} else {
		t.Log("version.Minor PASSED")
	}

	if version.Patch != 1 {
		t.Error("version.Patch FAILED")
	} else {
		t.Log("version.Patch PASSED")
	}

	if version.Describe != "2.2.1-git-45521019" {
		t.Error("version.Describe FAILED")
	} else {
		t.Log("version.Describe PASSED")
	}
}

func TestPrintVersion(t *testing.T) {
	version := GetVersion()
	version.Print()
	t.Log("PASSED")
}

func TestLoadFpga(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.LoadFpga("invalidAddr")

	if err.Error() != "File not found" {
		t.Errorf("FAILED cause got %v", err.Error())
	} else {
		t.Log("PASSED")
	}
}

func TestGetFpgaSize(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	size, err := rf.GetFpgaSize()

	if size != FpgaSizeA4 {
		t.Errorf("FAILED cause got %v", size)
	} else {
		t.Log("PASSED")
	}
}

func TestGetQuickTune(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_, err = rf.GetQuickTune(Rx1Channel)

	if err != nil {
		t.Errorf("FAILED cause got %v", err.Error())
	} else {
		t.Log("PASSED")
	}
}

func TestCancelScheduledReTunes(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.CancelScheduledReTunes(Rx1Channel)

	if err != nil {
		t.Errorf("FAILED cause got %v", err.Error())
	} else {
		t.Log("PASSED")
	}
}

func TestGetFpgaSource(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	source, err := rf.GetFpgaSource()

	if source == FpgaSourceHost {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", source)
	}
}

func TestGetFpgaBytes(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	bytes, err := rf.GetFpgaBytes()

	if bytes == 2632660 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", bytes)
	}
}

func TestGetFpgaFlashSize(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	size, isGuess, err := rf.GetFpgaFlashSize()

	if size == 4194304 && !isGuess {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v,%v", size, isGuess)
	}
}

func TestGetFirmwareVersion(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	version, err := rf.GetFirmwareVersion()

	if version.Major != 2 {
		t.Error("version.Major FAILED")
	} else {
		t.Log("version.Major PASSED")
	}

	if version.Minor != 3 {
		t.Error("version.Minor FAILED")
	} else {
		t.Log("version.Minor PASSED")
	}

	if version.Patch != 2 {
		t.Error("version.Patch FAILED")
	} else {
		t.Log("version.Patch PASSED")
	}

	if version.Describe != "2.3.2" {
		t.Error("version.Describe FAILED")
	} else {
		t.Log("version.Describe PASSED")
	}
}

func TestIsFpgaConfigured(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	configured, err := rf.IsFpgaConfigured()

	if configured {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", configured)
	}
}

func TestGetDeviceSpeed(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	speed := rf.GetDeviceSpeed()

	if speed == SpeedHigh {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", speed)
	}
}

func TestGetFpgaVersion(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	version, err := rf.GetFpgaVersion()

	if version.Major != 0 {
		t.Error("version.Major FAILED")
	} else {
		t.Log("version.Major PASSED")
	}

	if version.Minor != 11 {
		t.Error("version.Minor FAILED")
	} else {
		t.Log("version.Minor PASSED")
	}

	if version.Patch != 0 {
		t.Error("version.Patch FAILED")
	} else {
		t.Log("version.Patch PASSED")
	}

	if version.Describe != "0.11.0" {
		t.Error("version.Describe FAILED")
	} else {
		t.Log("version.Describe PASSED")
	}
}

func TestFreeDeviceList(t *testing.T) {
	devices, _ := GetDeviceList()
	defer devices[0].FreeDeviceList()
}

func TestGetDeviceList(t *testing.T) {
	devices, _ := GetDeviceList()
	defer devices[0].FreeDeviceList()

	if len(devices) == 1 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", len(devices))
	}
}

func TestGetBootloaderList(t *testing.T) {
	bootloaders, _ := GetBootloaderList()

	if len(bootloaders) == 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", len(bootloaders))
	}
}

func TestInitDeviceInfo(t *testing.T) {
	deviceInfo := InitDeviceInfo()

	if deviceInfo.Serial == "ANY" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", deviceInfo.Serial)
	}
}

func TestGetDeviceInfo(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()

	if deviceInfo.Product == "bladeRF 2.0" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", deviceInfo.Product)
	}
}

func TestDeviceInfoMatches(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()

	if err == nil && deviceInfo.DeviceInfoMatches(deviceInfo) {
		t.Log("PASSED")
	} else {
		t.Error("FAILED")
	}
}

func TestDeviceStringMatches(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()

	if err == nil && deviceInfo.DeviceStringMatches("*:serial="+deviceInfo.Serial) {
		t.Log("PASSED")
	} else {
		t.Error("FAILED")
	}
}

func TestGetDeviceInfoFromString(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()

	if err != nil {
		t.Errorf("FAILED cause got %v", err.Error())
	}

	deviceInfoFromString, err := GetDeviceInfoFromString("*:serial=" + deviceInfo.Serial)

	if err == nil && deviceInfoFromString.Serial == deviceInfo.Serial {
		t.Log("PASSED")
	} else {
		t.Error("FAILED")
	}
}

func TestOpenWithDeviceInfo(t *testing.T) {
	devices, err := GetDeviceList()
	defer devices[0].FreeDeviceList()

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(devices) == 0 {
		t.Error("Device cannot be found")
		t.FailNow()
	}

	rf, err := devices[0].Open()
	defer rf.Close()

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestOpenWithDeviceIdentifier(t *testing.T) {
	devices, err := GetDeviceList()
	defer devices[0].FreeDeviceList()

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(devices) == 0 {
		t.Error("Device cannot be found")
		t.FailNow()
	}

	rf, err := OpenWithDeviceIdentifier("*:serial=" + devices[0].Serial)
	defer rf.Close()

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestOpen(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	t.Log("PASSED")
}

func TestClose(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	rf.Close()

	t.Log("PASSED")
}

func TestSetLoopback(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetLoopback(LoopbackDisabled)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestIsLoopbackModeSupported(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	supported := rf.IsLoopbackModeSupported(LoopbackDisabled)

	if supported {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", supported)
	}
}

func TestGetLoopback(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	loopback, err := rf.GetLoopback()

	if loopback == LoopbackDisabled {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", loopback)
	}
}

func TestScheduleReTune(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	tune, err := rf.GetQuickTune(Rx1Channel)
	err = rf.ScheduleReTune(Rx1Channel, ReTuneNow, 1525420000, tune)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSelectBand(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SelectBand(Rx1Channel, 1525420000)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetFrequency(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetFrequency(Rx1Channel, 1525420000)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetFrequency(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetFrequency(Rx1Channel, 1525420000)
	freq, err := rf.GetFrequency(Rx1Channel)

	if err == nil && freq == 1525420000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetSampleRate(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	actual, err := rf.SetSampleRate(Rx1Channel, 2000000)

	if err == nil && actual == 2000000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetRxMux(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetRxMux(RxMuxBaseband)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetRxMux(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	rxMux, err := rf.GetRxMux()

	if err == nil && rxMux == RxMuxBaseband {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetRationalSampleRate(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	rate, err := rf.GetRationalSampleRate(Rx1Channel)
	_, err = rf.SetRationalSampleRate(Rx1Channel, rate)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetSampleRate(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	actual, err := rf.SetSampleRate(Rx1Channel, 2000000)
	sampleRate, err := rf.GetSampleRate(Rx1Channel)

	if err == nil && actual == 2000000 && sampleRate == 2000000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetRationalSampleRate(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	rate, err := rf.GetRationalSampleRate(Rx1Channel)

	if err == nil && rate.Integer == 30720000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetSampleRateRange(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_range, err := rf.GetSampleRateRange(Rx1Channel)

	if err == nil && _range.Max == 61440000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetFrequencyRange(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_range, err := rf.GetFrequencyRange(Rx1Channel)

	if err == nil && _range.Max == 6000000000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetBandwidth(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	actual, err := rf.SetBandwidth(Rx1Channel, 1500000)

	if err == nil && actual == 1500000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetBandwidth(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	actual, err := rf.SetBandwidth(Rx1Channel, 1500000)
	bandwidth, err := rf.GetBandwidth(Rx1Channel)

	if err == nil && actual == 1500000 && bandwidth == 1500000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetBandwidthRange(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_range, err := rf.GetBandwidthRange(Rx1Channel)

	if err == nil && _range.Max == 56000000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetGain(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(Rx1Channel)
	err = rf.SetGainMode(Rx1Channel, GainModeManual)
	err = rf.SetGain(Rx1Channel, 60)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetGain(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(Rx1Channel)
	err = rf.SetGainMode(Rx1Channel, GainModeManual)
	err = rf.SetGain(Rx1Channel, 60)
	gain, err := rf.GetGain(Rx1Channel)

	if err == nil && gain == 60 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", gain)
	}
}

func TestGetGainStage(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	stage, err := rf.GetGainStage(Rx1Channel, "full")

	if err == nil && stage == 71 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", stage)
	}
}

func TestGetGainMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetGainMode(Rx1Channel, GainModeManual)
	mode, err := rf.GetGainMode(Rx1Channel)

	if err == nil && mode == GainModeManual {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", mode)
	}
}

func TestSetGainStage(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(Rx1Channel)
	err = rf.SetGainMode(Rx1Channel, GainModeManual)
	err = rf.SetGainStage(Rx1Channel, "full", 70)
	stage, err := rf.GetGainStage(Rx1Channel, "full")

	if err == nil && stage == 70 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", stage)
	}
}

func TestGetGainStageRange(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_range, err := rf.GetGainStageRange(Rx1Channel, "full")

	if err == nil && _range.Min == -4 && _range.Max == 71 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetGainRange(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_range, err := rf.GetGainRange(Rx1Channel)

	if err == nil && _range.Min == -15 && _range.Max == 60 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetNumberOfGainStages(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	count, err := rf.GetNumberOfGainStages(Rx1Channel)

	if err == nil && count == 1 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetCorrection(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetCorrection(Rx1Channel, CorrectionDcoffI, 2048)
	err = rf.SetCorrection(Rx1Channel, CorrectionDcoffQ, 2048)
	correctionI, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffI)
	correctionQ, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffQ)

	if err == nil && correctionI == -2048 && correctionQ == -2048 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetCorrection(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	correctionI, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffI)
	correctionQ, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffQ)

	if err == nil && correctionI < 0 && correctionQ < 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestBackendString(t *testing.T) {
	libUSB := BackendLibUSB
	dummy := BackendDummy
	any := BackendAny
	cypress := BackendCypress
	linux := BackendLinux

	if libUSB.String() == "libusb" &&
		dummy.String() == "*" &&
		any.String() == "*" &&
		cypress.String() == "cypress" &&
		linux.String() == "linux" {
		t.Log("PASSED")
	} else {
		t.Error("FAILED")
	}
}

func TestGetBoardName(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	name := rf.GetBoardName()

	if name == "bladerf2" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", name)
	}
}

func TestSetUSBResetOnOpen(t *testing.T) {
	SetUSBResetOnOpen(true)
	t.Log("PASSED")
}

func TestGetSerial(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	serial, err := rf.GetSerial()

	if err == nil && len(serial) > 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetSerialStruct(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	serial, err := rf.GetSerialStruct()

	if err == nil && len(serial.Serial) > 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetGainStages(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	stages, err := rf.GetGainStages(Rx1Channel)

	if err == nil && stages[0] == "full" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetGainModes(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	modes, err := rf.GetGainModes(Rx1Channel)

	if err == nil &&
		modes[0].Name == "automatic" &&
		modes[0].Mode == GainModeDefault &&
		len(modes) == 5 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetLoopbackModes(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	modes, err := rf.GetLoopbackModes()

	if err == nil &&
		modes[0].Name == "none" &&
		modes[0].Mode == LoopbackDisabled &&
		len(modes) == 3 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetGainMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetGainMode(Rx1Channel, GainModeHybridAgc)
	mode, err := rf.GetGainMode(Rx1Channel)

	if err == nil && mode == GainModeHybridAgc {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", mode)
	}
}

func TestEnableModule(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(Rx1Channel)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestDisableModule(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.DisableModule(Rx1Channel)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestTrigger(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	triggerMaster, err := rf.TriggerInit(Rx1Channel, TriggerSignalJ714)
	triggerMaster.SetRole(TriggerRoleMaster)
	triggerSlave, err := rf.TriggerInit(Rx1Channel, TriggerSignalJ714)
	triggerSlave.SetRole(TriggerRoleSlave)

	err = rf.TriggerArm(triggerMaster, true, 0, 0)
	err = rf.TriggerFire(triggerMaster)

	isArmed, hasFire, fireRequested, resV1, resV2, err := rf.TriggerState(triggerMaster)

	if err == nil && isArmed && hasFire && fireRequested && resV1 == 0 && resV2 == 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSyncTX(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SyncConfig(TxX1, FormatSc16Q11, 2, 1024, 1, 3500)
	err = rf.EnableModule(ChannelTx(0))

	data := make([]complex64, 4)

	for i := range data {
		data[i] = complex(0.5, 0.5)
	}

	_, err = rf.SyncTX(Complex64ToInt16(data), Metadata{}, 3500)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSyncRX(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SyncConfig(RxX1, FormatSc16Q11, 2, 1024, 1, 3500)
	err = rf.EnableModule(ChannelRx(0))

	data, _, err := rf.SyncRX(1024, Metadata{}, 3500)

	complexData := Int16ToComplex64(data)

	if err == nil && len(data) == 1024 && len(complexData) == 512 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestAsyncRX(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(ChannelRx(0))

	rxStream, err := rf.InitStream(
		FormatSc16Q11,
		2,
		1024,
		1,
		func(data []int16) GoStream {
			complexData := Int16ToComplex64(data)

			if len(complexData) == 512 {
				t.Log("PASSED")
			} else {
				t.Errorf("FAILED cause got %v", len(complexData))
			}

			return GoStreamShutdown
		})

	err = rxStream.Start(RxX1)
	if err != nil {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestDeInitAsyncRX(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EnableModule(ChannelRx(0))

	rxStream, err := rf.InitStream(
		FormatSc16Q11,
		2,
		1024,
		1,
		func(data []int16) GoStream {
			return GoStreamShutdown
		})

	rxStream.DeInit()
	t.Log("PASSED")
}

func TestGetStreamTimeout(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	timeout, err := rf.GetStreamTimeout(Rx)

	if err == nil && timeout == 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetStreamTimeout(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetStreamTimeout(Rx, 3000)
	timeout, err := rf.GetStreamTimeout(Rx)

	if err == nil && timeout == 3000 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestAttachExpansionBoard(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.AttachExpansionBoard(ExpansionBoard300)

	if err.Error() == "Operation not supported" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetAttachedExpansionBoard(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	board, err := rf.GetAttachedExpansionBoard()

	if err == nil && board == ExpansionBoardNone {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetVctcxoTamerMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetVctcxoTamerMode(VctcxoTamerModeDisabled)

	if err.Error() == "Operation not supported" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetVctcxoTamerMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	_, err = rf.GetVctcxoTamerMode()

	if err.Error() == "Operation not supported" {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetVctcxoTrim(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	trim, err := rf.GetVctcxoTrim()

	if err == nil && trim == 8091 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestTrimDacRead(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	val, err := rf.TrimDacRead()

	if err == nil && val == 8091 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestTrimDacWrite(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.TrimDacWrite(8090)
	val, err := rf.TrimDacRead()

	if err == nil && val == 8090 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestSetTuningMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetTuningMode(TuningModeHost)
	mode, err := rf.GetTuningMode()

	if err == nil && mode == TuningModeHost {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetTuningMode(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.SetTuningMode(TuningModeFpga)
	mode, err := rf.GetTuningMode()

	if err == nil && mode == TuningModeFpga {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestGetTimestamp(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	timestamp, err := rf.GetTimestamp(Rx)

	if err == nil && timestamp == 0 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestReadTrigger(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.WriteTrigger(Rx1Channel, TriggerSignalJ511, 10)
	val, err := rf.ReadTrigger(Rx1Channel, TriggerSignalJ511)

	if err == nil && val == 10 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestWriteTrigger(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.WriteTrigger(Rx1Channel, TriggerSignalJ511, 12)
	val, err := rf.ReadTrigger(Rx1Channel, TriggerSignalJ511)

	if err == nil && val == 12 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestConfigGpioRead(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	val, err := rf.ConfigGpioRead()

	if err == nil && val == 129 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestConfigGpioWrite(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.ConfigGpioWrite(512)
	val, err := rf.ConfigGpioRead()

	if err == nil && val == 129+512 {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestEraseFlash(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EraseFlash(4, 41)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func TestEraseFlashBytes(t *testing.T) {
	rf, err := Open()

	if err != nil {
		t.Error(err)
	}

	defer rf.Close()

	err = rf.EraseFlashBytes(0x00040000, 0x290000)

	if err == nil {
		t.Log("PASSED")
	} else {
		t.Errorf("FAILED cause got %v", err.Error())
	}
}

func Complex64ToInt16(input []complex64) []int16 {
	var ex = make([]int16, len(input)*2)

	for i := 0; i < len(input)*2; i += 2 {
		ex[i] = int16(imag(input[i/2]) * 2048)
		ex[i+1] = int16(real(input[i/2]) * 2048)
	}

	return ex
}

func Int16ToComplex64(input []int16) []complex64 {
	var complexFloat = make([]complex64, len(input)/2)

	for i := 0; i < len(complexFloat); i++ {
		complexFloat[i] = complex(float32(input[2*i])/2048, float32(input[2*i+1])/2048)
	}

	return complexFloat
}

func TestReadFlashBytes(t *testing.T) {
	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	buf := make([]uint8, 0x290000)
	for index, _ := range buf {
		buf[index] = 180
	}

	rf.EraseFlashBytes(0x00040000, 0x290000)
	rf.WriteFlashBytes(buf, 0x00040000, 0x290000)
	rf.ReadFlashBytes(0x00040000, 0x290000)
}

func TestReadFlash(t *testing.T) {
	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	buf := make([]uint8, 0x290000)
	for index, _ := range buf {
		buf[index] = 180
	}

	rf.EraseFlash(4, 41)
	rf.WriteFlash(buf, 1024, 10496)
	rf.ReadFlash(1024, 10496)
}

func TestRfPort(t *testing.T) {
	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	channel := ChannelRx(0)

	test, _ := rf.GetRfPort(channel)
	fmt.Println(test)
	rf.SetRfPort(channel, test)
	count, _ := rf.GetNumberOfRfPorts(channel)
	fmt.Println(count)
	ports, _ := rf.GetRfPorts(channel)
	fmt.Println(len(ports))
}

func TestReadOtp(t *testing.T) {
	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	buf := make([]uint8, 0x290000)
	for index, _ := range buf {
		buf[index] = 180
	}

	rf.ReadOtp()
}

func TestChannel(t *testing.T) {
	a := ChannelRx(0)
	b := ChannelTx(0)
	c := ChannelRx(1)
	d := ChannelTx(1)
	fmt.Println(a, b, c, d)
	fmt.Println(ChannelIsTx(0))
	fmt.Println(ChannelIsTx(1))
	fmt.Println(ChannelIsTx(2))
	fmt.Println(ChannelIsTx(3))
}
