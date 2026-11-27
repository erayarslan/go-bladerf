package bladerf

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

func startSimRemoteServer(t *testing.T) (string, func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen failed: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.Command("./build/dist/bin/bladerf-remote", "-l", addr, "--sim")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start bladerf-remote: %v", err)
	}

	ready := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if dialErr == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	stop := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}

	if !ready {
		stop()
		t.Fatalf("bladerf-remote did not listen on %s: %s", addr, stderr.String())
	}

	return addr, stop
}

func TestRemoteRDMAFallbackOpenAndControls(t *testing.T) {
	addr, stop := startSimRemoteServer(t)
	defer stop()

	rf, err := OpenRDMA(addr)
	if err != nil {
		t.Fatalf("OpenRDMA(%q) failed: %v", addr, err)
	}
	defer rf.Close()

	if name := rf.GetBoardName(); name != "bladerf2" {
		t.Errorf("GetBoardName() = %q, want %q", name, "bladerf2")
	}

	info, err := rf.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo() failed: %v", err)
	}
	if info.Serial == "" {
		t.Error("expected non-empty Serial")
	}
	if info.Addr != addr {
		t.Errorf("info.Addr = %q, want %q", info.Addr, addr)
	}
	if info.Backend != BackendTCP && info.Backend != BackendRDMA {
		t.Errorf("unexpected backend: %v (%s)", info.Backend, info.Backend.String())
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

	if err := rf.SetFrequency(Rx1Channel, 99500000); err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}
	freq, err := rf.GetFrequency(Rx1Channel)
	if err != nil {
		t.Fatalf("GetFrequency failed: %v", err)
	}
	if freq != 99500000 {
		t.Errorf("GetFrequency() = %d, want 99500000", freq)
	}

	actualRate, err := rf.SetSampleRate(Rx1Channel, 2400000)
	if err != nil {
		t.Fatalf("SetSampleRate failed: %v", err)
	}
	if actualRate != 2400000 {
		t.Errorf("SetSampleRate() = %d, want 2400000", actualRate)
	}
	rate, err := rf.GetSampleRate(Rx1Channel)
	if err != nil {
		t.Fatalf("GetSampleRate failed: %v", err)
	}
	if rate != 2400000 {
		t.Errorf("GetSampleRate() = %d, want 2400000", rate)
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
	gmode, err := rf.GetGainMode(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainMode failed: %v", err)
	}
	if gmode != GainModeManual {
		t.Errorf("GetGainMode() = %v, want %v", gmode, GainModeManual)
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

	stages, err := rf.GetGainStages(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainStages failed: %v", err)
	}
	if len(stages) == 0 {
		t.Error("expected non-empty gain stages")
	}

	gmodes, err := rf.GetGainModes(Rx1Channel)
	if err != nil {
		t.Fatalf("GetGainModes failed: %v", err)
	}
	if len(gmodes) == 0 {
		t.Error("expected non-empty gain modes")
	}

	lbModes, err := rf.GetLoopbackModes()
	if err != nil {
		t.Fatalf("GetLoopbackModes failed: %v", err)
	}
	if len(lbModes) == 0 {
		t.Error("expected non-empty loopback modes")
	}

	rfPorts, err := rf.GetRfPorts(Rx1Channel)
	if err != nil {
		t.Fatalf("GetRfPorts failed: %v", err)
	}
	if len(rfPorts) == 0 {
		t.Error("expected non-empty RF ports")
	}

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule failed: %v", err)
	}
	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule failed: %v", err)
	}

	tcpDev, err := OpenTCP(addr)
	if err != nil {
		t.Fatalf("OpenTCP(%q) failed: %v", addr, err)
	}
	defer tcpDev.Close()

	tcpInfo, err := tcpDev.GetDeviceInfo()
	if err != nil {
		t.Fatalf("OpenTCP GetDeviceInfo failed: %v", err)
	}
	if tcpInfo.Backend != BackendTCP {
		t.Errorf("OpenTCP backend = %v, want %v", tcpInfo.Backend, BackendTCP)
	}
}

func TestRemoteSyncRXAndTX(t *testing.T) {
	addr, stop := startSimRemoteServer(t)
	defer stop()

	rf, err := OpenRDMA(addr)
	if err != nil {
		t.Fatalf("OpenRDMA failed: %v", err)
	}
	defer rf.Close()

	if err := rf.SyncConfig(RxX1, FormatSc16Q11, 16, 4096, 8, 2000); err != nil {
		t.Fatalf("SyncConfig(RX) failed: %v", err)
	}
	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule(RX) failed: %v", err)
	}

	raw, meta, err := rf.SyncRX(1024, NewMetadata(0, MetaFlagRxNow), 2000)
	if err != nil {
		t.Fatalf("SyncRX failed: %v", err)
	}
	if len(raw) != 2048 {
		t.Fatalf("SyncRX len(raw) = %d, want 2048", len(raw))
	}
	if meta.ActualCount == 0 {
		t.Error("expected non-zero meta.ActualCount from SyncRX")
	}

	samples := Int16ToComplex64(raw)
	if len(samples) != 1024 {
		t.Fatalf("Int16ToComplex64 len = %d, want 1024", len(samples))
	}
	energy := 0.0
	for _, s := range samples {
		re := float64(real(s))
		im := float64(imag(s))
		energy += re*re + im*im
	}
	if energy == 0 || math.IsNaN(energy) {
		t.Errorf("unexpected RX IQ energy: %v", energy)
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

	txComplex := make([]complex64, 512)
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
}

func TestRemoteAsyncStreamRX(t *testing.T) {
	addr, stop := startSimRemoteServer(t)
	defer stop()

	rf, err := OpenRDMA(addr)
	if err != nil {
		t.Fatalf("OpenRDMA failed: %v", err)
	}
	defer rf.Close()

	if err := rf.EnableModule(Rx1Channel); err != nil {
		t.Fatalf("EnableModule(RX) failed: %v", err)
	}

	callbacks := 0
	nonZeroSamples := 0
	stream, err := rf.InitStream(FormatSc16Q11, 16, 4096, 8, func(data []int16) GoStream {
		callbacks++
		for _, v := range data {
			if v != 0 {
				nonZeroSamples++
				break
			}
		}
		if callbacks >= 4 {
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
	if callbacks < 4 {
		t.Errorf("expected >= 4 stream callbacks, got %d", callbacks)
	}
	if nonZeroSamples == 0 {
		t.Error("expected non-zero samples from async stream")
	}

	if err := rf.DisableModule(Rx1Channel); err != nil {
		t.Fatalf("DisableModule(RX) failed: %v", err)
	}
}

func TestRemoteEnvAutoDiscovery(t *testing.T) {
	addr, stop := startSimRemoteServer(t)
	defer stop()

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
		t.Fatalf("GetDeviceList failed: %v", err)
	}
	if len(devices) > 0 {
		defer devices[0].FreeDeviceList()
	}

	found := false
	for _, d := range devices {
		if d.Addr == addr {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GetDeviceList() did not include remote device with Addr=%q (got %d devices)", addr, len(devices))
	}

	rf, err := Open()
	if err != nil {
		t.Fatalf("Open() with BLADERF_REMOTE_ADDR=%q failed: %v", addr, err)
	}
	defer rf.Close()

	if name := rf.GetBoardName(); name != "bladerf2" {
		t.Errorf("GetBoardName() = %q, want %q", name, "bladerf2")
	}
}
