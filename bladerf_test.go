package bladerf

import (
	"errors"
	"math"
	"os"
	"testing"

	exception "github.com/erayarslan/go-bladerf/error"
	"github.com/erayarslan/go-bladerf/mock"
)

func TestMain(m *testing.M) {
	if os.Getenv("BLADERF_REMOTE_ADDR") == "" {
		devices, err := GetDeviceList()
		if err != nil || len(devices) == 0 {
			srv, srvErr := mock.NewServer("127.0.0.1:0")
			if srvErr == nil {
				_ = os.Setenv("BLADERF_REMOTE_ADDR", srv.Addr())
				code := m.Run()
				_ = os.Unsetenv("BLADERF_REMOTE_ADDR")
				_ = srv.Close()
				os.Exit(code)
			}
		} else {
			devices[0].FreeDeviceList()
		}
	}
	os.Exit(m.Run())
}

var Rx1Channel = ChannelRx(0)

func openTestDevice(t *testing.T) BladeRF {
	t.Helper()

	rf, err := Open()
	if err != nil {
		if errors.Is(err, exception.Nodev) {
			t.Skipf("no bladeRF device available: %v", err)
		}
		t.Fatalf("failed to open device: %v", err)
	}

	return rf
}

func requireDeviceList(t *testing.T) []DeviceInfo {
	t.Helper()

	devices, err := GetDeviceList()
	if err != nil {
		if errors.Is(err, exception.Nodev) {
			t.Skipf("no bladeRF device available: %v", err)
		}
		t.Fatalf("GetDeviceList failed: %v", err)
	}
	if len(devices) == 0 {
		t.Skip("no bladeRF device available")
	}

	return devices
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()

	if version.Major < 2 {
		t.Errorf("unexpected version.Major: got %d, want >= 2", version.Major)
	}
	if version.Describe == "" {
		t.Error("expected non-empty version.Describe")
	}
}

func TestPrintVersion(t *testing.T) {
	version := GetVersion()
	version.Print()
	if version.String() == "" {
		t.Error("expected non-empty version.String()")
	}
}

func TestLoadFpga(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	err := rf.LoadFpga("invalidAddr")
	if !errors.Is(err, exception.NoFile) {
		t.Errorf("LoadFpga(\"invalidAddr\") error = %v, want %v", err, exception.NoFile)
	}
}

func TestGetFpgaSize(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	size, err := rf.GetFpgaSize()
	if err != nil {
		t.Fatalf("GetFpgaSize failed: %v", err)
	}
	if size != FpgaSizeA4 {
		t.Errorf("got %v, want %v", size, FpgaSizeA4)
	}
}

func TestGetQuickTune(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if _, err := rf.GetQuickTune(Rx1Channel); err != nil {
		t.Fatalf("GetQuickTune failed: %v", err)
	}
}

func TestCancelScheduledReTunes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.CancelScheduledReTunes(Rx1Channel); err != nil {
		t.Fatalf("CancelScheduledReTunes failed: %v", err)
	}
}

func TestGetFpgaSource(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	source, err := rf.GetFpgaSource()
	if err != nil {
		t.Fatalf("GetFpgaSource failed: %v", err)
	}
	if source != FpgaSourceHost {
		t.Errorf("got %v, want %v", source, FpgaSourceHost)
	}
}

func TestGetFpgaBytes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	bytes, err := rf.GetFpgaBytes()
	if err != nil {
		t.Fatalf("GetFpgaBytes failed: %v", err)
	}
	if bytes != 2632660 {
		t.Errorf("got %v, want 2632660", bytes)
	}
}

func TestGetFpgaFlashSize(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	size, isGuess, err := rf.GetFpgaFlashSize()
	if err != nil {
		t.Fatalf("GetFpgaFlashSize failed: %v", err)
	}
	if size != 4194304 || isGuess {
		t.Errorf("got size=%v isGuess=%v, want size=4194304 isGuess=false", size, isGuess)
	}
}

func TestGetFirmwareVersion(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	version, err := rf.GetFirmwareVersion()
	if err != nil {
		t.Fatalf("GetFirmwareVersion failed: %v", err)
	}
	if version.Major != 2 || version.Minor != 3 || version.Patch != 2 || version.Describe != "2.3.2" {
		t.Errorf("got firmware version %+v, want 2.3.2", version)
	}
}

func TestIsFpgaConfigured(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	configured, err := rf.IsFpgaConfigured()
	if err != nil {
		t.Fatalf("IsFpgaConfigured failed: %v", err)
	}
	if !configured {
		t.Errorf("got %v, want true", configured)
	}
}

func TestGetDeviceSpeed(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	speed := rf.GetDeviceSpeed()
	if speed != SpeedHigh {
		t.Errorf("got %v, want %v", speed, SpeedHigh)
	}
}

func TestGetFpgaVersion(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	version, err := rf.GetFpgaVersion()
	if err != nil {
		t.Fatalf("GetFpgaVersion failed: %v", err)
	}
	if version.Major != 0 || version.Minor != 11 || version.Patch != 0 || version.Describe != "0.11.0" {
		t.Errorf("got FPGA version %+v, want 0.11.0", version)
	}
}

func TestFreeDeviceList(t *testing.T) {
	devices := requireDeviceList(t)
	defer devices[0].FreeDeviceList()
}

func TestGetDeviceList(t *testing.T) {
	devices := requireDeviceList(t)
	defer devices[0].FreeDeviceList()

	if len(devices) != 1 {
		t.Errorf("got %v devices, want 1", len(devices))
	}
}

func TestGetBootloaderList(t *testing.T) {
	bootloaders, err := GetBootloaderList()
	if err != nil {
		if errors.Is(err, exception.Nodev) {
			return
		}
		if _, usbErr := os.Stat("/dev/bus/usb"); usbErr != nil {
			t.Skipf("USB bus unavailable (%v); GetBootloaderList returned: %v", usbErr, err)
		}
		t.Fatalf("GetBootloaderList failed: %v", err)
	}
	if len(bootloaders) > 0 {
		defer bootloaders[0].FreeDeviceList()
	}
	if len(bootloaders) != 0 {
		t.Errorf("got %v bootloaders, want 0", len(bootloaders))
	}
}

func TestInitDeviceInfo(t *testing.T) {
	deviceInfo := InitDeviceInfo()
	if deviceInfo.Serial != "ANY" {
		t.Errorf("got %q, want %q", deviceInfo.Serial, "ANY")
	}

	// FreeDeviceList must be safe on a non-list DeviceInfo.
	deviceInfo.FreeDeviceList()
}

func TestGetDeviceInfo(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}
	if deviceInfo.Product != "bladeRF 2.0" {
		t.Errorf("got %q, want %q", deviceInfo.Product, "bladeRF 2.0")
	}
}

func TestDeviceInfoMatches(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}
	if !deviceInfo.DeviceInfoMatches(deviceInfo) {
		t.Error("expected DeviceInfoMatches to return true")
	}
}

func TestDeviceStringMatches(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}
	if !deviceInfo.DeviceStringMatches("*:serial=" + deviceInfo.Serial) {
		t.Error("expected DeviceStringMatches to return true")
	}
}

func TestGetDeviceInfoFromString(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	deviceInfo, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}

	deviceInfoFromString, err := GetDeviceInfoFromString("*:serial=" + deviceInfo.Serial)
	if err != nil {
		t.Fatalf("GetDeviceInfoFromString failed: %v", err)
	}
	if deviceInfoFromString.Serial != deviceInfo.Serial {
		t.Errorf("got serial %q, want %q", deviceInfoFromString.Serial, deviceInfo.Serial)
	}
}

func TestOpenWithDeviceInfo(t *testing.T) {
	devices := requireDeviceList(t)
	defer devices[0].FreeDeviceList()

	rf, err := devices[0].Open()
	if err != nil {
		t.Fatalf("devices[0].Open failed: %v", err)
	}
	defer rf.Close()
}

func TestOpenWithDeviceIdentifier(t *testing.T) {
	devices := requireDeviceList(t)
	defer devices[0].FreeDeviceList()

	rf, err := OpenWithDeviceIdentifier("*:serial=" + devices[0].Serial)
	if err != nil {
		t.Fatalf("OpenWithDeviceIdentifier failed: %v", err)
	}
	defer rf.Close()
}

func TestOpen(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()
}

func TestClose(t *testing.T) {
	rf := openTestDevice(t)
	rf.Close()
	rf.Close()
}

func TestSetLoopback(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetLoopback(LoopbackDisabled); err != nil {
		t.Fatalf("SetLoopback failed: %v", err)
	}
}

func TestIsLoopbackModeSupported(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if !rf.IsLoopbackModeSupported(LoopbackDisabled) {
		t.Error("expected LoopbackDisabled to be supported")
	}
}

func TestGetLoopback(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	loopback, err := rf.GetLoopback()
	if err != nil {
		t.Fatalf("GetLoopback failed: %v", err)
	}
	if loopback != LoopbackDisabled {
		t.Errorf("got %v, want %v", loopback, LoopbackDisabled)
	}
}

func TestScheduleReTune(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	tune, err := rf.GetQuickTune(Rx1Channel)
	if err != nil {
		t.Fatalf("GetQuickTune failed: %v", err)
	}
	if err := rf.ScheduleReTune(Rx1Channel, ReTuneNow, 1525420000, tune); err != nil {
		t.Fatalf("ScheduleReTune failed: %v", err)
	}
}

func TestSelectBand(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SelectBand(Rx1Channel, 1525420000); err != nil {
		t.Fatalf("SelectBand failed: %v", err)
	}
}

func TestSetFrequency(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetFrequency(Rx1Channel, 1525420000); err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}
}

func TestGetFrequency(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetFrequency(Rx1Channel, 1525420000); err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}
	freq, err := rf.GetFrequency(Rx1Channel)
	if err != nil {
		t.Fatalf("GetFrequency failed: %v", err)
	}
	if freq != 1525420000 {
		t.Errorf("got %v, want 1525420000", freq)
	}
}

func TestSetSampleRate(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	actual, err := rf.SetSampleRate(Rx1Channel, 2000000)
	if err != nil {
		t.Fatalf("SetSampleRate failed: %v", err)
	}
	if actual != 2000000 {
		t.Errorf("got %v, want 2000000", actual)
	}
}

func TestSetRxMux(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetRxMux(RxMuxBaseband); err != nil {
		t.Fatalf("SetRxMux failed: %v", err)
	}
}

func TestGetRxMux(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	rxMux, err := rf.GetRxMux()
	if err != nil {
		t.Fatalf("GetRxMux failed: %v", err)
	}
	if rxMux != RxMuxBaseband {
		t.Errorf("got %v, want %v", rxMux, RxMuxBaseband)
	}
}

func TestSetRationalSampleRate(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	rate, err := rf.GetRationalSampleRate(Rx1Channel)
	if err != nil {
		t.Fatalf("GetRationalSampleRate failed: %v", err)
	}
	if _, err := rf.SetRationalSampleRate(Rx1Channel, rate); err != nil {
		t.Fatalf("SetRationalSampleRate failed: %v", err)
	}
}

func TestGetSampleRate(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	actual, err := rf.SetSampleRate(Rx1Channel, 2000000)
	if err != nil {
		t.Fatalf("SetSampleRate failed: %v", err)
	}
	sampleRate, err := rf.GetSampleRate(Rx1Channel)
	if err != nil {
		t.Fatalf("GetSampleRate failed: %v", err)
	}
	if actual != 2000000 || sampleRate != 2000000 {
		t.Errorf("got actual=%v sampleRate=%v, want 2000000", actual, sampleRate)
	}
}

func TestGetRationalSampleRate(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	rate, err := rf.GetRationalSampleRate(Rx1Channel)
	if err != nil {
		t.Fatalf("GetRationalSampleRate failed: %v", err)
	}
	if rate.Integer != 30720000 {
		t.Errorf("got %v, want 30720000", rate.Integer)
	}
}

func TestGetSampleRateRange(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	r, err := rf.GetSampleRateRange(Rx1Channel)
	if err != nil {
		t.Fatalf("GetSampleRateRange failed: %v", err)
	}
	if r.Max != 61440000 {
		t.Errorf("got %v, want 61440000", r.Max)
	}
}

func TestGetFrequencyRange(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	r, err := rf.GetFrequencyRange(Rx1Channel)
	if err != nil {
		t.Fatalf("GetFrequencyRange failed: %v", err)
	}
	if r.Max != 6000000000 {
		t.Errorf("got %v, want 6000000000", r.Max)
	}
}

func TestSetBandwidth(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	actual, err := rf.SetBandwidth(Rx1Channel, 1500000)
	if err != nil {
		t.Fatalf("SetBandwidth failed: %v", err)
	}
	if actual != 1500000 {
		t.Errorf("got %v, want 1500000", actual)
	}
}

func TestGetBandwidth(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	actual, err := rf.SetBandwidth(Rx1Channel, 1500000)
	if err != nil {
		t.Fatalf("SetBandwidth failed: %v", err)
	}
	bandwidth, err := rf.GetBandwidth(Rx1Channel)
	if err != nil {
		t.Fatalf("GetBandwidth failed: %v", err)
	}
	if actual != 1500000 || bandwidth != 1500000 {
		t.Errorf("got actual=%v bandwidth=%v, want 1500000", actual, bandwidth)
	}
}

func TestGetBandwidthRange(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	r, err := rf.GetBandwidthRange(Rx1Channel)
	if err != nil {
		t.Fatalf("GetBandwidthRange failed: %v", err)
	}
	if r.Max != 56000000 {
		t.Errorf("got %v, want 56000000", r.Max)
	}
}

func TestSetGain(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
	if err := rf.SetGainMode(Rx1Channel, GainModeManual); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	if err := rf.SetGain(Rx1Channel, 60); err != nil {
		t.Fatalf("SetGain failed: %v", err)
	}
}

func TestGetGain(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
	if err := rf.SetGainMode(Rx1Channel, GainModeManual); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	if err := rf.SetGain(Rx1Channel, 60); err != nil {
		t.Fatalf("SetGain failed: %v", err)
	}
	gain, err := rf.GetGain(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGain failed: %v", err)
	}
	if gain != 60 {
		t.Errorf("got %v, want 60", gain)
	}
}

func TestGetGainStage(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	stage, err := rf.GetGainStage(Rx1Channel, "full")
	if err != nil {
		t.Fatalf("GetGainStage failed: %v", err)
	}
	if stage != 71 {
		t.Errorf("got %v, want 71", stage)
	}
}

func TestGetGainMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetGainMode(Rx1Channel, GainModeManual); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	mode, err := rf.GetGainMode(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainMode failed: %v", err)
	}
	if mode != GainModeManual {
		t.Errorf("got %v, want %v", mode, GainModeManual)
	}
}

func TestSetGainStage(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
	if err := rf.SetGainMode(Rx1Channel, GainModeManual); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	if err := rf.SetGainStage(Rx1Channel, "full", 70); err != nil {
		t.Fatalf("SetGainStage failed: %v", err)
	}
	stage, err := rf.GetGainStage(Rx1Channel, "full")
	if err != nil {
		t.Fatalf("GetGainStage failed: %v", err)
	}
	if stage != 70 {
		t.Errorf("got %v, want 70", stage)
	}
}

func TestGetGainStageRange(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	r, err := rf.GetGainStageRange(Rx1Channel, "full")
	if err != nil {
		t.Fatalf("GetGainStageRange failed: %v", err)
	}
	if r.Min != -4 || r.Max != 71 {
		t.Errorf("got min=%v max=%v, want min=-4 max=71", r.Min, r.Max)
	}
}

func TestGetGainRange(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	r, err := rf.GetGainRange(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainRange failed: %v", err)
	}
	if r.Min != -15 || r.Max != 60 {
		t.Errorf("got min=%v max=%v, want min=-15 max=60", r.Min, r.Max)
	}
}

func TestGetNumberOfGainStages(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	count, err := rf.GetNumberOfGainStages(Rx1Channel)
	if err != nil {
		t.Fatalf("GetNumberOfGainStages failed: %v", err)
	}
	if count != 1 {
		t.Errorf("got %v, want 1", count)
	}
}

func TestSetCorrection(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetCorrection(Rx1Channel, CorrectionDcoffI, 2048); err != nil {
		t.Fatalf("SetCorrection(I) failed: %v", err)
	}
	if err := rf.SetCorrection(Rx1Channel, CorrectionDcoffQ, 2048); err != nil {
		t.Fatalf("SetCorrection(Q) failed: %v", err)
	}
	correctionI, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffI)
	if err != nil {
		t.Fatalf("GetCorrection(I) failed: %v", err)
	}
	correctionQ, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffQ)
	if err != nil {
		t.Fatalf("GetCorrection(Q) failed: %v", err)
	}
	if correctionI != -2048 || correctionQ != -2048 {
		t.Errorf("got I=%v Q=%v, want -2048", correctionI, correctionQ)
	}
}

func TestGetCorrection(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	correctionI, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffI)
	if err != nil {
		t.Fatalf("GetCorrection(I) failed: %v", err)
	}
	correctionQ, err := rf.GetCorrection(Rx1Channel, CorrectionDcoffQ)
	if err != nil {
		t.Fatalf("GetCorrection(Q) failed: %v", err)
	}
	if correctionI >= 0 || correctionQ >= 0 {
		t.Errorf("got I=%v Q=%v, want < 0", correctionI, correctionQ)
	}
}

func TestBackendString(t *testing.T) {
	tests := []struct {
		backend Backend
		want    string
	}{
		{BackendLibUSB, "libusb"},
		{BackendDummy, "*"},
		{BackendAny, "*"},
		{BackendCypress, "cypress"},
		{BackendLinux, "linux"},
		{BackendRDMA, "rdma"},
		{BackendTCP, "tcp"},
	}

	for _, tc := range tests {
		if got := tc.backend.String(); got != tc.want {
			t.Errorf("Backend(%d).String() = %q, want %q", tc.backend, got, tc.want)
		}
	}
}

func TestGetBoardName(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if name := rf.GetBoardName(); name != "bladerf2" {
		t.Errorf("got %q, want %q", name, "bladerf2")
	}
}

func TestSetUSBResetOnOpen(t *testing.T) {
	SetUSBResetOnOpen(true)
}

func TestGetSerial(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	serial, err := rf.GetSerial()
	if err != nil {
		t.Fatalf("GetSerial failed: %v", err)
	}
	if len(serial) == 0 {
		t.Error("expected non-empty serial")
	}
}

func TestGetSerialStruct(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	serial, err := rf.GetSerialStruct()
	if err != nil {
		t.Fatalf("GetSerialStruct failed: %v", err)
	}
	if len(serial.Serial) == 0 {
		t.Error("expected non-empty serial.Serial")
	}
}

func TestGetGainStages(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	stages, err := rf.GetGainStages(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainStages failed: %v", err)
	}
	if len(stages) == 0 || stages[0] != "full" {
		t.Errorf("got %v, want first stage %q", stages, "full")
	}
}

func TestGetGainModes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	modes, err := rf.GetGainModes(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainModes failed: %v", err)
	}
	if len(modes) != 5 || modes[0].Name != "automatic" || modes[0].Mode != GainModeDefault {
		t.Errorf("unexpected gain modes: %+v", modes)
	}
}

func TestGetLoopbackModes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	modes, err := rf.GetLoopbackModes()
	if err != nil {
		t.Fatalf("GetLoopbackModes failed: %v", err)
	}
	if len(modes) != 3 || modes[0].Name != "none" || modes[0].Mode != LoopbackDisabled {
		t.Errorf("unexpected loopback modes: %+v", modes)
	}
}

func TestSetGainMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetGainMode(Rx1Channel, GainModeHybridAgc); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	mode, err := rf.GetGainMode(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainMode failed: %v", err)
	}
	if mode != GainModeHybridAgc {
		t.Errorf("got %v, want %v", mode, GainModeHybridAgc)
	}
}

func TestEnableModule(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
}

func TestDisableModule(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule failed: %v", err)
	}
}

func TestTrigger(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	triggerMaster, err := rf.TriggerInit(Rx1Channel, TriggerSignalJ714)
	if err != nil {
		t.Fatalf("TriggerInit(master) failed: %v", err)
	}
	triggerMaster.SetRole(TriggerRoleMaster)

	triggerSlave, err := rf.TriggerInit(Rx1Channel, TriggerSignalJ714)
	if err != nil {
		t.Fatalf("TriggerInit(slave) failed: %v", err)
	}
	triggerSlave.SetRole(TriggerRoleSlave)

	if err := rf.TriggerArm(triggerMaster, true, 0, 0); err != nil {
		t.Fatalf("TriggerArm failed: %v", err)
	}
	if err := rf.TriggerFire(triggerMaster); err != nil {
		t.Fatalf("TriggerFire failed: %v", err)
	}

	isArmed, hasFired, fireRequested, resV1, resV2, err := rf.TriggerState(triggerMaster)
	if err != nil {
		t.Fatalf("TriggerState failed: %v", err)
	}
	if !isArmed || !hasFired || !fireRequested || resV1 != 0 || resV2 != 0 {
		t.Errorf("unexpected trigger state: armed=%v fired=%v req=%v resV1=%d resV2=%d",
			isArmed, hasFired, fireRequested, resV1, resV2)
	}
}

func TestSyncTX(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SyncConfig(TxX1, FormatSc16Q11, 2, 1024, 1, 3500); err != nil {
		t.Fatalf("SyncConfig failed: %v", err)
	}
	if err := rf.EnableModule(ChannelTx(0)); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}

	data := make([]complex64, 4)
	for i := range data {
		data[i] = complex(0.5, 0.5)
	}

	if _, err := rf.SyncTX(Complex64ToInt16(data), Metadata{}, 3500); err != nil {
		t.Fatalf("SyncTX failed: %v", err)
	}
}

func TestSyncRX(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SyncConfig(RxX1, FormatSc16Q11, 2, 1024, 1, 3500); err != nil {
		t.Fatalf("SyncConfig failed: %v", err)
	}
	if err := rf.EnableModule(ChannelRx(0)); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}

	data, _, err := rf.SyncRX(1024, Metadata{}, 3500)
	if err != nil {
		t.Fatalf("SyncRX failed: %v", err)
	}

	complexData := Int16ToComplex64(data)
	if len(data) != 2048 || len(complexData) != 1024 {
		t.Errorf("got len(data)=%d len(complexData)=%d, want 2048 and 1024", len(data), len(complexData))
	}
}

func TestAsyncRX(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(ChannelRx(0)); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}

	rxStream, err := rf.InitStream(
		FormatSc16Q11,
		2,
		1024,
		1,
		func(data []int16) GoStream {
			complexData := Int16ToComplex64(data)
			if len(complexData) != 1024 {
				t.Errorf("got len(complexData)=%d, want 1024", len(complexData))
			}

			return GoStreamShutdown
		},
	)
	if err != nil {
		t.Fatalf("InitStream failed: %v", err)
	}
	defer rxStream.DeInit()

	if err := rxStream.Start(RxX1); err != nil {
		t.Fatalf("rxStream.Start failed: %v", err)
	}
}

func TestDeInitAsyncRX(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EnableModule(ChannelRx(0)); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}

	rxStream, err := rf.InitStream(
		FormatSc16Q11,
		2,
		1024,
		1,
		func(data []int16) GoStream {
			return GoStreamShutdown
		},
	)
	if err != nil {
		t.Fatalf("InitStream failed: %v", err)
	}

	rxStream.DeInit()
	rxStream.DeInit()
}

func TestGetStreamTimeout(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	timeout, err := rf.GetStreamTimeout(Rx)
	if err != nil {
		t.Fatalf("GetStreamTimeout failed: %v", err)
	}
	if timeout != 0 {
		t.Errorf("got %v, want 0", timeout)
	}
}

func TestSetStreamTimeout(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetStreamTimeout(Rx, 3000); err != nil {
		t.Fatalf("SetStreamTimeout failed: %v", err)
	}
	timeout, err := rf.GetStreamTimeout(Rx)
	if err != nil {
		t.Fatalf("GetStreamTimeout failed: %v", err)
	}
	if timeout != 3000 {
		t.Errorf("got %v, want 3000", timeout)
	}
}

func TestAttachExpansionBoard(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	err := rf.AttachExpansionBoard(ExpansionBoard300)
	if !errors.Is(err, exception.Unsupported) {
		t.Errorf("AttachExpansionBoard error = %v, want %v", err, exception.Unsupported)
	}
}

func TestGetAttachedExpansionBoard(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	board, err := rf.GetAttachedExpansionBoard()
	if err != nil {
		t.Fatalf("GetAttachedExpansionBoard failed: %v", err)
	}
	if board != ExpansionBoardNone {
		t.Errorf("got %v, want %v", board, ExpansionBoardNone)
	}
}

func TestSetVctcxoTamerMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	err := rf.SetVctcxoTamerMode(VctcxoTamerModeDisabled)
	if !errors.Is(err, exception.Unsupported) {
		t.Errorf("SetVctcxoTamerMode error = %v, want %v", err, exception.Unsupported)
	}
}

func TestGetVctcxoTamerMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	_, err := rf.GetVctcxoTamerMode()
	if !errors.Is(err, exception.Unsupported) {
		t.Errorf("GetVctcxoTamerMode error = %v, want %v", err, exception.Unsupported)
	}
}

func TestGetVctcxoTrim(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	trim, err := rf.GetVctcxoTrim()
	if err != nil {
		t.Fatalf("GetVctcxoTrim failed: %v", err)
	}
	if trim != 8091 {
		t.Errorf("got %v, want 8091", trim)
	}
}

func TestTrimDacRead(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	val, err := rf.TrimDacRead()
	if err != nil {
		t.Fatalf("TrimDacRead failed: %v", err)
	}
	if val != 8091 {
		t.Errorf("got %v, want 8091", val)
	}
}

func TestTrimDacWrite(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.TrimDacWrite(8090); err != nil {
		t.Fatalf("TrimDacWrite failed: %v", err)
	}
	val, err := rf.TrimDacRead()
	if err != nil {
		t.Fatalf("TrimDacRead failed: %v", err)
	}
	if val != 8090 {
		t.Errorf("got %v, want 8090", val)
	}
}

func TestSetTuningMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetTuningMode(TuningModeHost); err != nil {
		t.Fatalf("SetTuningMode failed: %v", err)
	}
	mode, err := rf.GetTuningMode()
	if err != nil {
		t.Fatalf("GetTuningMode failed: %v", err)
	}
	if mode != TuningModeHost {
		t.Errorf("got %v, want %v", mode, TuningModeHost)
	}
}

func TestGetTuningMode(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.SetTuningMode(TuningModeFpga); err != nil {
		t.Fatalf("SetTuningMode failed: %v", err)
	}
	mode, err := rf.GetTuningMode()
	if err != nil {
		t.Fatalf("GetTuningMode failed: %v", err)
	}
	if mode != TuningModeFpga {
		t.Errorf("got %v, want %v", mode, TuningModeFpga)
	}
}

func TestGetTimestamp(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	timestamp, err := rf.GetTimestamp(Rx)
	if err != nil {
		t.Fatalf("GetTimestamp failed: %v", err)
	}
	if timestamp != 0 {
		t.Errorf("got %v, want 0", timestamp)
	}
}

func TestReadTrigger(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.WriteTrigger(Rx1Channel, TriggerSignalJ511, 10); err != nil {
		t.Fatalf("WriteTrigger failed: %v", err)
	}
	val, err := rf.ReadTrigger(Rx1Channel, TriggerSignalJ511)
	if err != nil {
		t.Fatalf("ReadTrigger failed: %v", err)
	}
	if val != 10 {
		t.Errorf("got %v, want 10", val)
	}
}

func TestWriteTrigger(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.WriteTrigger(Rx1Channel, TriggerSignalJ511, 12); err != nil {
		t.Fatalf("WriteTrigger failed: %v", err)
	}
	val, err := rf.ReadTrigger(Rx1Channel, TriggerSignalJ511)
	if err != nil {
		t.Fatalf("ReadTrigger failed: %v", err)
	}
	if val != 12 {
		t.Errorf("got %v, want 12", val)
	}
}

func TestConfigGpioRead(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	val, err := rf.ConfigGpioRead()
	if err != nil {
		t.Fatalf("ConfigGpioRead failed: %v", err)
	}
	if val != 129 {
		t.Errorf("got %v, want 129", val)
	}
}

func TestConfigGpioWrite(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.ConfigGpioWrite(512); err != nil {
		t.Fatalf("ConfigGpioWrite failed: %v", err)
	}
	val, err := rf.ConfigGpioRead()
	if err != nil {
		t.Fatalf("ConfigGpioRead failed: %v", err)
	}
	if val != 129+512 {
		t.Errorf("got %v, want %v", val, 129+512)
	}
}

func TestEraseFlash(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EraseFlash(4, 41); err != nil {
		t.Fatalf("EraseFlash failed: %v", err)
	}
}

func TestEraseFlashBytes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	if err := rf.EraseFlashBytes(0x00040000, 0x290000); err != nil {
		t.Fatalf("EraseFlashBytes failed: %v", err)
	}
}

func TestReadFlashBytes(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	buf := make([]uint8, 0x290000)
	for i := range buf {
		buf[i] = 180
	}

	if err := rf.EraseFlashBytes(0x00040000, 0x290000); err != nil {
		t.Fatalf("EraseFlashBytes failed: %v", err)
	}
	if err := rf.WriteFlashBytes(buf, 0x00040000, 0x290000); err != nil {
		t.Fatalf("WriteFlashBytes failed: %v", err)
	}
	if _, err := rf.ReadFlashBytes(0x00040000, 0x290000); err != nil {
		t.Fatalf("ReadFlashBytes failed: %v", err)
	}
}

func TestReadFlash(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	buf := make([]uint8, 0x290000)
	for i := range buf {
		buf[i] = 180
	}

	if err := rf.EraseFlash(4, 41); err != nil {
		t.Fatalf("EraseFlash failed: %v", err)
	}
	if err := rf.WriteFlash(buf, 1024, 10496); err != nil {
		t.Fatalf("WriteFlash failed: %v", err)
	}
	if _, err := rf.ReadFlash(1024, 10496); err != nil {
		t.Fatalf("ReadFlash failed: %v", err)
	}
}

func TestRfPort(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	channel := ChannelRx(0)

	port, err := rf.GetRfPort(channel)
	if err != nil {
		t.Fatalf("GetRfPort failed: %v", err)
	}
	if err := rf.SetRfPort(channel, port); err != nil {
		t.Fatalf("SetRfPort failed: %v", err)
	}
	count, err := rf.GetNumberOfRfPorts(channel)
	if err != nil {
		t.Fatalf("GetNumberOfRfPorts failed: %v", err)
	}
	ports, err := rf.GetRfPorts(channel)
	if err != nil {
		t.Fatalf("GetRfPorts failed: %v", err)
	}
	if len(ports) != count {
		t.Errorf("got len(ports)=%d, want %d", len(ports), count)
	}
}

func TestReadOtp(t *testing.T) {
	rf := openTestDevice(t)
	defer rf.Close()

	otp, err := rf.ReadOtp()
	if err != nil {
		t.Fatalf("ReadOtp failed: %v", err)
	}
	if len(otp) != OtpSize {
		t.Errorf("got len(otp)=%d, want %d", len(otp), OtpSize)
	}
}

func TestChannel(t *testing.T) {
	rx0 := ChannelRx(0)
	tx0 := ChannelTx(0)
	rx1 := ChannelRx(1)
	tx1 := ChannelTx(1)

	if rx0.IsTx() || ChannelIsTx(rx0) {
		t.Errorf("expected RX0 (%d) not to be TX", rx0)
	}
	if !tx0.IsTx() || !ChannelIsTx(tx0) {
		t.Errorf("expected TX0 (%d) to be TX", tx0)
	}
	if rx1.IsTx() || ChannelIsTx(rx1) {
		t.Errorf("expected RX1 (%d) not to be TX", rx1)
	}
	if !tx1.IsTx() || !ChannelIsTx(tx1) {
		t.Errorf("expected TX1 (%d) to be TX", tx1)
	}
}

func TestComplexConversionRoundTrip(t *testing.T) {
	in := []complex64{
		complex(0.25, -0.5),
		complex(-0.125, 0.75),
	}

	raw := Complex64ToInt16(in)
	wantRaw := []int16{512, -1024, -256, 1536}
	if len(raw) != len(wantRaw) {
		t.Fatalf("got len(raw)=%d, want %d", len(raw), len(wantRaw))
	}
	for i := range wantRaw {
		if raw[i] != wantRaw[i] {
			t.Errorf("raw[%d] = %d, want %d", i, raw[i], wantRaw[i])
		}
	}

	out := Int16ToComplex64(raw)
	if len(out) != len(in) {
		t.Fatalf("got len(out)=%d, want %d", len(out), len(in))
	}
	for i := range in {
		if math.Abs(float64(real(out[i])-real(in[i]))) > 1e-3 ||
			math.Abs(float64(imag(out[i])-imag(in[i]))) > 1e-3 {
			t.Errorf("out[%d] = %v, want %v", i, out[i], in[i])
		}
	}
}

func TestNewMetadata(t *testing.T) {
	meta := NewMetadata(12345, MetaFlagRxNow)
	if meta.Timestamp != 12345 || meta.Flags != MetaFlagRxNow {
		t.Fatalf("unexpected Metadata: %+v", meta)
	}

	meta.Status = 7
	meta.ActualCount = 64

	cMeta := meta.toC()
	roundTripped := LoadMetadata(&cMeta)
	if roundTripped != meta {
		t.Errorf("round-tripped metadata = %+v, want %+v", roundTripped, meta)
	}
}

func TestDeviceInfoMutationReflectsInCStruct(t *testing.T) {
	info := InitDeviceInfo()
	info.Serial = "1234567890abcdef"
	info.Backend = BackendTCP
	info.Addr = "127.0.0.1:5555"

	cInfo := info.cStruct()
	rt := NewDeviceInfo(&cInfo)
	if rt.Serial != "1234567890abcdef" || rt.Backend != BackendTCP || rt.Addr != "127.0.0.1:5555" {
		t.Fatalf("mutated DeviceInfo did not propagate to cStruct: got %+v", rt)
	}
}

func TestClosedHandleReturnsErrorSafely(t *testing.T) {
	var zero BladeRF
	if _, err := zero.GetDeviceInfo(); !errors.Is(err, exception.Inval) {
		t.Errorf("zero.GetDeviceInfo err = %v, want Inval", err)
	}
	if _, err := zero.GetFrequency(ChannelRx(0)); !errors.Is(err, exception.Inval) {
		t.Errorf("zero.GetFrequency err = %v, want Inval", err)
	}
	if _, err := zero.GetGain(ChannelRx(0)); !errors.Is(err, exception.Inval) {
		t.Errorf("zero.GetGain err = %v, want Inval", err)
	}
	if _, _, err := zero.SyncRX(1024, Metadata{}, 100); !errors.Is(err, exception.Inval) {
		t.Errorf("zero.SyncRX err = %v, want Inval", err)
	}
	if _, err := zero.IsFpgaConfigured(); !errors.Is(err, exception.Inval) {
		t.Errorf("zero.IsFpgaConfigured err = %v, want Inval", err)
	}
	if zero.GetBoardName() != "" || zero.GetDeviceSpeed() != SpeedUnknown || zero.IsLoopbackModeSupported(LoopbackFirmware) {
		t.Errorf("unexpected non-empty getter results on zero BladeRF")
	}

	var zeroStream Stream
	if err := zeroStream.Start(RxX1); !errors.Is(err, exception.Inval) {
		t.Errorf("zeroStream.Start err = %v, want Inval", err)
	}
	zeroStream.DeInit()
	zero.Close()
}

func TestComplex64ToInt16Clamping(t *testing.T) {
	in := []complex64{
		complex(1.0, -1.0),
		complex(2.5, -3.5),
		complex(float32(math.NaN()), 0.5),
	}
	got := Complex64ToInt16(in)
	want := []int16{2047, -2048, 2047, -2048, 0, 1024}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Complex64ToInt16[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
