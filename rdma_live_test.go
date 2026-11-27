//go:build rdma_live
// +build rdma_live

package bladerf

import (
	"math"
	"os"
	"testing"
)

func requireLiveRDMAAddr(t *testing.T) string {
	t.Helper()

	addr := os.Getenv("BLADERF_TEST_RDMA_ADDR")
	if addr == "" {
		t.Fatal("BLADERF_TEST_RDMA_ADDR must be set")
	}
	return addr
}

func TestLiveRDMAControlAndBackendVerification(t *testing.T) {
	addr := requireLiveRDMAAddr(t)

	rf, err := OpenRDMA(addr)
	if err != nil {
		t.Fatalf("OpenRDMA(%q) failed: %v", addr, err)
	}
	defer rf.Close()

	info, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo() failed: %v", err)
	}
	if info.Backend != BackendRDMA || info.Backend.String() != "rdma" {
		t.Fatalf("expected BackendRDMA (rdma), got %v (%q) — RDMA fell back to TCP", info.Backend, info.Backend.String())
	}
	if info.Addr != addr {
		t.Errorf("info.Addr = %q, want %q", info.Addr, addr)
	}
	if name := rf.GetBoardName(); name != "bladerf2" {
		t.Errorf("GetBoardName() = %q, want %q", name, "bladerf2")
	}

	fwVer, err := rf.GetFirmwareVersion()
	if err != nil {
		t.Fatalf("GetFirmwareVersion() failed: %v", err)
	}
	if fwVer.Describe == "" {
		t.Error("expected non-empty firmware Describe")
	}

	fpgaVer, err := rf.GetFpgaVersion()
	if err != nil {
		t.Fatalf("GetFpgaVersion() failed: %v", err)
	}
	if fpgaVer.Describe == "" {
		t.Error("expected non-empty FPGA Describe")
	}

	if err := rf.SetFrequency(Rx1Channel, 101100000); err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}
	freq, err := rf.GetFrequency(Rx1Channel)
	if err != nil {
		t.Fatalf("GetFrequency failed: %v", err)
	}
	if freq != 101100000 {
		t.Errorf("GetFrequency() = %d, want 101100000", freq)
	}

	actualRate, err := rf.SetSampleRate(Rx1Channel, 2000000)
	if err != nil {
		t.Fatalf("SetSampleRate failed: %v", err)
	}
	if actualRate != 2000000 {
		t.Errorf("SetSampleRate() = %d, want 2000000", actualRate)
	}
	rate, err := rf.GetSampleRate(Rx1Channel)
	if err != nil {
		t.Fatalf("GetSampleRate failed: %v", err)
	}
	if rate != 2000000 {
		t.Errorf("GetSampleRate() = %d, want 2000000", rate)
	}

	actualBw, err := rf.SetBandwidth(Rx1Channel, 1500000)
	if err != nil {
		t.Fatalf("SetBandwidth failed: %v", err)
	}
	if actualBw != 1500000 {
		t.Errorf("SetBandwidth() = %d, want 1500000", actualBw)
	}
	bw, err := rf.GetBandwidth(Rx1Channel)
	if err != nil {
		t.Fatalf("GetBandwidth failed: %v", err)
	}
	if bw != 1500000 {
		t.Errorf("GetBandwidth() = %d, want 1500000", bw)
	}

	if err := rf.SetGainMode(Rx1Channel, GainModeManual); err != nil {
		t.Fatalf("SetGainMode failed: %v", err)
	}
	if err := rf.SetGain(Rx1Channel, 42); err != nil {
		t.Fatalf("SetGain failed: %v", err)
	}
	gain, err := rf.GetGain(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGain failed: %v", err)
	}
	if gain != 42 {
		t.Errorf("GetGain() = %d, want 42", gain)
	}

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule failed: %v", err)
	}

	prevAddr, hadAddr := os.LookupEnv("BLADERF_REMOTE_ADDR")
	if err := os.Setenv("BLADERF_REMOTE_ADDR", addr); err != nil {
		t.Fatalf("os.Setenv failed: %v", err)
	}
	defer func() {
		if hadAddr {
			_ = os.Setenv("BLADERF_REMOTE_ADDR", prevAddr)
		} else {
			_ = os.Unsetenv("BLADERF_REMOTE_ADDR")
		}
	}()

	devices, err := GetDeviceList()
	if err != nil {
		t.Fatalf("GetDeviceList() failed: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("GetDeviceList() returned 0 devices")
	}
	defer devices[0].FreeDeviceList()

	if devices[0].Backend != BackendRDMA {
		t.Fatalf("devices[0].Backend = %v (%q), want BackendRDMA (rdma)", devices[0].Backend, devices[0].Backend.String())
	}

	envDev, err := Open()
	if err != nil {
		t.Fatalf("Open() with BLADERF_REMOTE_ADDR=%q failed: %v", addr, err)
	}
	defer envDev.Close()

	envInfo, err := envDev.GetDeviceInfo()
	if err != nil {
		t.Fatalf("envDev.GetDeviceInfo() failed: %v", err)
	}
	if envInfo.Backend != BackendRDMA {
		t.Fatalf("envInfo.Backend = %v (%q), want BackendRDMA (rdma)", envInfo.Backend, envInfo.Backend.String())
	}
}

func TestLiveRDMASyncAndAsyncIQStreaming(t *testing.T) {
	addr := requireLiveRDMAAddr(t)

	rf, err := OpenRDMA(addr)
	if err != nil {
		t.Fatalf("OpenRDMA(%q) failed: %v", addr, err)
	}
	defer rf.Close()

	info, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo() failed: %v", err)
	}
	if info.Backend != BackendRDMA {
		t.Fatalf("expected BackendRDMA (rdma), got %v (%q)", info.Backend, info.Backend.String())
	}

	if err := rf.SyncConfig(RxX1, FormatSc16Q11, 16, 4096, 8, 2000); err != nil {
		t.Fatalf("SyncConfig(RX) failed: %v", err)
	}
	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule(RX) failed: %v", err)
	}

	raw, meta, err := rf.SyncRX(4096, NewMetadata(0, MetaFlagRxNow), 2000)
	if err != nil {
		t.Fatalf("SyncRX failed: %v", err)
	}
	if len(raw) != 8192 {
		t.Fatalf("SyncRX len(raw) = %d, want 8192", len(raw))
	}
	if meta.ActualCount != 4096 {
		t.Errorf("SyncRX meta.ActualCount = %d, want 4096", meta.ActualCount)
	}

	samples := Int16ToComplex64(raw)
	if len(samples) != 4096 {
		t.Fatalf("Int16ToComplex64 len = %d, want 4096", len(samples))
	}
	energy := 0.0
	for _, s := range samples {
		re := float64(real(s))
		im := float64(imag(s))
		energy += re*re + im*im
	}
	if energy == 0 || math.IsNaN(energy) {
		t.Fatalf("unexpected RX IQ energy: %v", energy)
	}

	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule(RX) failed: %v", err)
	}

	tx1 := ChannelTx(0)
	if err := rf.SyncConfig(TxX1, FormatSc16Q11, 16, 4096, 8, 2000); err != nil {
		t.Fatalf("SyncConfig(TX) failed: %v", err)
	}
	if err := rf.EnableModule(tx1); err != nil {
		t.Fatalf("EnableModule(TX) failed: %v", err)
	}

	txComplex := make([]complex64, 1024)
	for i := range txComplex {
		phase := float64(i) * 0.1
		txComplex[i] = complex(float32(0.5*math.Cos(phase)), float32(0.5*math.Sin(phase)))
	}
	txRaw := Complex64ToInt16(txComplex)
	if _, err := rf.SyncTX(txRaw, NewMetadata(0, MetaFlagTxNow|MetaFlagTxBurstStart|MetaFlagTxBurstEnd), 2000); err != nil {
		t.Fatalf("SyncTX failed: %v", err)
	}
	if err := rf.DisableModule(tx1); err != nil {
		t.Fatalf("DisableModule(TX) failed: %v", err)
	}

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule(RX async) failed: %v", err)
	}

	callbacks := 0
	nonZeroCallbacks := 0
	stream, err := rf.InitStream(FormatSc16Q11, 16, 4096, 8, func(data []int16) GoStream {
		callbacks++
		if len(data) != 8192 {
			t.Errorf("async callback len(data) = %d, want 8192", len(data))
		}
		for _, v := range data {
			if v != 0 {
				nonZeroCallbacks++
				break
			}
		}
		if callbacks >= 8 {
			return GoStreamShutdown
		}
		return GoStreamNext
	})
	if err != nil {
		t.Fatalf("InitStream failed: %v", err)
	}
	defer stream.DeInit()

	if err := stream.Start(RxX1); err != nil {
		t.Fatalf("stream.Start failed: %v", err)
	}
	if callbacks < 8 {
		t.Errorf("expected >= 8 async callbacks, got %d", callbacks)
	}
	if nonZeroCallbacks < 8 {
		t.Errorf("expected non-zero samples in async callbacks, got %d", nonZeroCallbacks)
	}
	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule(RX async) failed: %v", err)
	}
}
