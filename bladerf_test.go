package bladerf

import (
	"fmt"
	"github.com/erayarslan/go-bladerf/log"
	"github.com/gordonklaus/portaudio"
	fifo "github.com/racerxdl/go.fifo"
	"github.com/racerxdl/segdsp/demodcore"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

const audioBufferSize = 8192 / 4

var audioStream *portaudio.Stream
var audioFifo = fifo.NewQueue()

var demodulator demodcore.DemodCore

/*
RX Gain Stage names: lna, rxvga1, rxvga2
TX Gain Stage names: txvga1, txvga2
*/

func ProcessAudio(out []float32) {
	if audioFifo.Len() > 0 {
		var z = audioFifo.Next().([]float32)
		copy(out, z)
	} else {
		for i := range out {
			out[i] = 0
		}
	}
}

func GetFinalData(input []int16) []complex64 {
	var complexFloat = make([]complex64, len(input)/2)

	for i := 0; i < len(complexFloat); i++ {
		complexFloat[i] = complex(float32(input[2*i])/2048, float32(input[2*i+1])/2048)
	}

	return complexFloat
}

func GetFinalDataString(input []int16) string {
	var runes []rune

	for i := 0; i < len(input)/2; i++ {
		runes = append(runes, rune(input[2*i+1]))
	}

	return string(runes)
}

func TestBackendSTR(t *testing.T) {
	x := BackendLibUSB
	y := BackendDummy
	z := BackendAny
	a := BackendCypress
	b := BackendLinux
	fmt.Println(x.String(), y.String(), z.String(), a.String(), b.String())
}

func TestSetUSBResetOnOpen(t *testing.T) {
	SetUSBResetOnOpen(true)
	SetUSBResetOnOpen(false)
}

func TestGetSerial(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	serial, _ := rf.GetSerial()
	fmt.Println(serial)
}

func TestGetSerialStruct(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	serial, _ := rf.GetSerialStruct()
	fmt.Print(serial)
}

func TestGetFpgaSize(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	size, _ := rf.GetFpgaSize()
	fmt.Print(size)
}

func TestGetFpgaSource(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	source, _ := rf.GetFpgaSource()
	fmt.Print(source)
}

func TestGetDeviceSpeed(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	speed := rf.GetDeviceSpeed()
	fmt.Print(speed)
}

func TestGetBoardName(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	name := rf.GetBoardName()
	fmt.Print(name)
}

func TestGetRationalSampleRate(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	rate, _ := rf.GetRationalSampleRate(ChannelRx(1))
	fmt.Print(rate)
}

func TestGetFpgaBytes(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	bytes, _ := rf.GetFpgaBytes()
	fmt.Print(bytes)
}

func TestGetFpgaFlashSize(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	size, guess, err := rf.GetFpgaFlashSize()

	if err != nil {
		panic(err)
	} else {
		fmt.Print(size)
		fmt.Print(guess)
	}
}

func TestGetFirmwareVersion(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	version, err := rf.GetFirmwareVersion()

	if err != nil {
		panic(err)
	}

	fmt.Print(version.describe)
}

func TestGetFpgaVersion(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	version, _ := rf.GetFpgaVersion()
	fmt.Print(version.describe)
}

func TestGetLoopbackModes(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	modes, _ := rf.GetLoopbackModes()
	for _, v := range modes {
		fmt.Printf("%s\n\n", v.name)
	}
}

func TestGetQuickTune(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	channel := ChannelRx(1)

	quickTune, err := rf.GetQuickTune(channel)

	if err != nil {
		panic(err)
	}

	err = rf.ScheduleReTune(channel, ReTuneNow, 96600000, quickTune)

	if err != nil {
		panic(err)
	}
}

func TestExpansionBoard(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	board, err := rf.GetAttachedExpansionBoard()

	if err != nil {
		panic(err)
	}

	fmt.Println(board)
}

func TestReTuneNow(t *testing.T) {
	fmt.Println(ReTuneNow)
}

func TestTrigger(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	channel := ChannelRx(0)
	signal := TriggerSignalJ714

	triggerMaster, err := rf.TriggerInit(channel, signal)

	if err != nil {
		panic(err)
	}

	triggerMaster.SetRole(TriggerRoleMaster)

	triggerSlave, err := rf.TriggerInit(channel, signal)

	if err != nil {
		panic(err)
	}

	triggerSlave.SetRole(TriggerRoleSlave)

	err = rf.TriggerArm(triggerMaster, true, 0, 0)

	if err != nil {
		panic(err)
	}

	err = rf.TriggerFire(triggerMaster)

	if err != nil {
		panic(err)
	}

	a, b, c, x, y, err := rf.TriggerState(triggerMaster)

	if err != nil {
		panic(err)
	}

	fmt.Println(a, b, c, x, y)
}

func TestIsFpgaConfigured(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	isConfigured, err := rf.IsFpgaConfigured()

	if err != nil {
		panic(err)
	}

	fmt.Print(isConfigured)
}

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Error)

	version := GetVersion()
	version.Print()

	bootloaders, _ := GetBootloaderList()
	fmt.Printf("Bootloaders Len: %d\n", len(bootloaders))

	devices, _ := GetDeviceList()
	fmt.Printf("Devices Len: %d\n", len(devices))
	rf, _ := devices[0].Open()
	_ = rf.LoadFpga("/Users/erayarslan/Downloads/hostedxA4-latest.rbf")
	rf.Close()

	rf, _ = Open()
	info, _ := rf.GetDeviceInfo()
	rf.Close()
	h, _ := rf.GetDeviceInfo()
	out, _ := h.Open()
	out.Close()
	c1, _ := Open()
	c1.Close()
	c2, _ := OpenWithDeviceIdentifier("*:serial=" + info.serial)
	c2.Close()
	o1, _ := GetDeviceInfoFromString("*:serial=" + info.serial)
	out2, _ := o1.Open()
	out2.Close()

	g, _ := rf.GetDeviceInfo()
	result := g.DeviceInfoMatches(g)
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	g0, _ := rf.GetDeviceInfo()
	result = g0.DeviceStringMatches("*:serial=" + info.serial)
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	rf, _ = Open()

	err := rf.EnableModule(ChannelRx(1))
	if err != nil {
		fmt.Println(err)
	}
	_, _ = rf.InitStream(FormatSc16Q11, 16, audioBufferSize, 8, cb)
	// _ = StartStream(stream, RX_X1)
}

func TestSetGainStage(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	stages, _ := rf.GetGainStages(ChannelRx(1))
	fmt.Println(len(stages))
	bfRange, _ := rf.GetGainStageRange(ChannelRx(1), stages[0])
	_ = rf.SetGainStage(ChannelRx(1), stages[0], int(bfRange.max))
	gain, _ := rf.GetGainStage(ChannelRx(1), stages[0])
	fmt.Println(gain)
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

func TestStream(t *testing.T) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()
	channel := ChannelRx(1)

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	_ = rf.SetFrequency(channel, 96600000)
	_range, _ := rf.GetSampleRateRange(channel)
	fmt.Printf("Min: %d, Max: %d, Step: %d\n", _range.min, _range.max, _range.step)
	_, _ = rf.SetSampleRate(channel, 4e6)
	_ = rf.SyncConfig(RxX2, FormatSc16Q11, 16, audioBufferSize, 8, 32)
	actual, _ := rf.SetBandwidth(channel, 240000)
	fmt.Println(actual)
	_ = rf.EnableModule(channel)
	_ = rf.SetGainMode(channel, GainModeHybridAgc)

	demodulator = demodcore.MakeWBFMDemodulator(uint32(2e6), 80e3, 48000)

	_ = portaudio.Initialize()
	h, _ := portaudio.DefaultHostApi()

	p := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
	p.Input.Channels = 0
	p.Output.Channels = 1
	p.SampleRate = 48000
	p.FramesPerBuffer = audioBufferSize

	audioStream, _ = portaudio.OpenStream(p, ProcessAudio)
	_ = audioStream.Start()

	go func() {
		for {
			data, _ := rf.SyncRX(audioBufferSize)
			out := demodulator.Work(GetFinalData(data))

			if out != nil {
				var o = out.(demodcore.DemodData)
				var nBf = make([]float32, len(o.Data))
				copy(nBf, o.Data)
				var buffs = len(nBf) / audioBufferSize
				for i := 0; i < buffs; i++ {
					audioFifo.Add(nBf[audioBufferSize*i : audioBufferSize*(i+1)])
				}
			}
		}
	}()

	<-sig
	fmt.Println("shit")
}

func cb(data []int16) {
	out := demodulator.Work(GetFinalData(data))

	if out != nil {
		var o = out.(demodcore.DemodData)
		var nBf = make([]float32, len(o.Data))
		copy(nBf, o.Data)
		var buffs = len(nBf) / audioBufferSize
		for i := 0; i < buffs; i++ {
			audioFifo.Add(nBf[audioBufferSize*i : audioBufferSize*(i+1)])
		}
	}
}

func TestGetGainModes(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	_, _ = rf.GetGainModes(ChannelRx(1))
}

func TestGetGainRange(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	bfRange, _ := rf.GetGainRange(ChannelRx(1))
	fmt.Println(bfRange.max)
}

func TestGPSData(t *testing.T) {
	log.SetVerbosity(log.Debug)
	channel := ChannelRx(1)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := devices[0].Open()
	defer rf.Close()

	_ = rf.SetFrequency(channel, 1525420000)
	_ = rf.SyncConfig(RxX2, FormatSc16Q11, 16, audioBufferSize, 8, 32)
	_, _ = rf.SetSampleRate(channel, 2600000)
	_, _ = rf.SetBandwidth(channel, 2500000)
	_ = rf.SetGainMode(channel, GainModeHybridAgc)
	_ = rf.EnableModule(channel)

	for {
		data, _ := rf.SyncRX(audioBufferSize)
		out := GetFinalDataString(data)
		fmt.Println(out)
	}
}

func TestAsyncStream(t *testing.T) {
	log.SetVerbosity(log.Debug)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	channel := ChannelRx(0)

	devices, err := GetDeviceList()

	if err != nil {
		panic(err)
	}

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, err := devices[0].Open()

	if err != nil {
		panic(err)
	}

	defer rf.Close()

	_ = rf.SetFrequency(channel, 96600000)
	_, _ = rf.SetSampleRate(channel, 4e6)
	_, _ = rf.SetBandwidth(channel, 240000)
	_ = rf.SetGainMode(channel, GainModeHybridAgc)
	_ = rf.EnableModule(channel)

	rxStream, err := rf.InitStream(FormatSc16Q11, 16, audioBufferSize, 8, cb)
	if err != nil {
		panic(err)
	}
	defer rxStream.DeInit()

	_ = rf.SetStreamTimeout(Rx, 32)
	timeout, _ := rf.GetStreamTimeout(Rx)
	println(timeout)

	demodulator = demodcore.MakeWBFMDemodulator(uint32(2e6), 80e3, 48000)

	_ = portaudio.Initialize()
	h, _ := portaudio.DefaultHostApi()

	p := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
	p.Input.Channels = 0
	p.Output.Channels = 1
	p.SampleRate = 48000
	p.FramesPerBuffer = audioBufferSize

	audioStream, _ = portaudio.OpenStream(p, ProcessAudio)
	_ = audioStream.Start()

	go func() {
		err = rxStream.Start(RxX2)
		if err != nil {
			panic(err)
		}
	}()

	<-sig
	fmt.Println("done")
}
