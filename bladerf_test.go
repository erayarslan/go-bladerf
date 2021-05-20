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
	x := BackendString(BackendLibUSB)
	y := BackendString(BackendDummy)
	z := BackendString(BackendAny)
	a := BackendString(BackendCypress)
	b := BackendString(BackendLinux)
	fmt.Println(x, y, z, a, b)
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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	serial, _ := GetSerial(rf)
	fmt.Println(serial)
}

func TestGetSerialStruct(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	serial, _ := GetSerialStruct(rf)
	fmt.Print(serial)
}

func TestGetFpgaSize(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	size, _ := GetFpgaSize(rf)
	fmt.Print(size)
}

func TestGetFpgaSource(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	source, _ := GetFpgaSource(rf)
	fmt.Print(source)
}

func TestGetDeviceSpeed(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	speed := GetDeviceSpeed(rf)
	fmt.Print(speed)
}

func TestGetBoardName(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	name := GetBoardName(rf)
	fmt.Print(name)
}

func TestGetRationalSampleRate(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	rate, _ := GetRationalSampleRate(rf, ChannelRx(1))
	fmt.Print(rate)
}

func TestGetFpgaBytes(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	bytes, _ := GetFpgaBytes(rf)
	fmt.Print(bytes)
}

func TestGetFpgaFlashSize(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	size, guess, err := GetFpgaFlashSize(rf)

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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	version, err := GetFirmwareVersion(rf)

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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	version, _ := GetFpgaVersion(rf)
	fmt.Print(version.describe)
}

func TestGetLoopbackModes(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	modes, _ := GetLoopbackModes(rf)
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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	channel := ChannelRx(1)

	quickTune, err := GetQuickTune(rf, channel)

	if err != nil {
		panic(err)
	}

	err = ScheduleReTune(rf, channel, ReTuneNow, 96600000, quickTune)

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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	board, err := GetAttachedExpansionBoard(rf)

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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	channel := ChannelRx(0)
	signal := TriggerSignalJ714

	triggerMaster, err := TriggerInit(rf, channel, signal)

	if err != nil {
		panic(err)
	}

	triggerMaster.SetRole(TriggerRoleMaster)

	triggerSlave, err := TriggerInit(rf, channel, signal)

	if err != nil {
		panic(err)
	}

	triggerSlave.SetRole(TriggerRoleSlave)

	err = TriggerArm(rf, triggerMaster, true, 0, 0)

	if err != nil {
		panic(err)
	}

	err = TriggerFire(rf, triggerMaster)

	if err != nil {
		panic(err)
	}

	a, b, c, x, y, err := TriggerState(rf, triggerMaster)

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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	isConfigured, err := IsFpgaConfigured(rf)

	if err != nil {
		panic(err)
	}

	fmt.Print(isConfigured)
}

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Error)

	PrintVersion(GetVersion())

	bootloaders, _ := GetBootloaderList()
	fmt.Printf("Bootloaders Len: %d\n", len(bootloaders))

	devices, _ := GetDeviceList()
	fmt.Printf("Devices Len: %d\n", len(devices))
	rf, _ := OpenWithDeviceInfo(devices[0])
	_ = LoadFpga(rf, "/Users/erayarslan/Downloads/hostedxA4-latest.rbf")
	Close(rf)

	rf, _ = Open()
	info, _ := GetDeviceInfo(rf)
	Close(rf)
	h, _ := GetDeviceInfo(rf)
	out, _ := OpenWithDeviceInfo(h)
	Close(out)
	c1, _ := Open()
	Close(c1)
	c2, _ := OpenWithDeviceIdentifier("*:serial=" + info.serial)
	Close(c2)
	o1, _ := GetDeviceInfoFromString("*:serial=" + info.serial)
	out2, _ := OpenWithDeviceInfo(o1)
	Close(out2)

	g, _ := GetDeviceInfo(rf)
	result := DeviceInfoMatches(g, g)
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	g0, _ := GetDeviceInfo(rf)
	result = DeviceStringMatches("*:serial="+info.serial, g0)
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	rf, _ = Open()

	err := EnableModule(rf, ChannelRx(1))
	if err != nil {
		fmt.Println(err)
	}
	_, _ = InitStream(rf, FormatSc16Q11, 16, audioBufferSize, 8, cb)
	// _ = StartStream(stream, RX_X1)
}

func TestSetGainStage(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	stages, _ := GetGainStages(rf, ChannelRx(1))
	fmt.Println(len(stages))
	bfRange, _ := GetGainStageRange(rf, ChannelRx(1), stages[0])
	_ = SetGainStage(rf, ChannelRx(1), stages[0], int(bfRange.max))
	gain, _ := GetGainStage(rf, ChannelRx(1), stages[0])
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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	_ = SetFrequency(rf, channel, 96600000)
	_range, _ := GetSampleRateRange(rf, channel)
	fmt.Printf("Min: %d, Max: %d, Step: %d\n", _range.min, _range.max, _range.step)
	_, _ = SetSampleRate(rf, channel, 4e6)
	_ = SyncConfig(rf, RxX2, FormatSc16Q11, 16, audioBufferSize, 8, 32)
	actual, _ := SetBandwidth(rf, channel, 240000)
	fmt.Println(actual)
	_ = EnableModule(rf, channel)
	_ = SetGainMode(rf, channel, GainModeHybridAgc)

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
			data, _ := SyncRX(rf, audioBufferSize)
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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	_, _ = GetGainModes(rf, ChannelRx(1))
}

func TestGetGainRange(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices, _ := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	bfRange, _ := GetGainRange(rf, ChannelRx(1))
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

	rf, _ := OpenWithDeviceInfo(devices[0])
	defer Close(rf)

	_ = SetFrequency(rf, channel, 1525420000)
	_ = SyncConfig(rf, RxX2, FormatSc16Q11, 16, audioBufferSize, 8, 32)
	_, _ = SetSampleRate(rf, channel, 2600000)
	_, _ = SetBandwidth(rf, channel, 2500000)
	_ = SetGainMode(rf, channel, GainModeHybridAgc)
	_ = EnableModule(rf, channel)

	for {
		data, _ := SyncRX(rf, audioBufferSize)
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

	rf, err := OpenWithDeviceInfo(devices[0])

	if err != nil {
		panic(err)
	}

	defer Close(rf)

	_ = SetFrequency(rf, channel, 96600000)
	_, _ = SetSampleRate(rf, channel, 4e6)
	_, _ = SetBandwidth(rf, channel, 240000)
	_ = SetGainMode(rf, channel, GainModeHybridAgc)
	_ = EnableModule(rf, channel)

	rxStream, err := InitStream(rf, FormatSc16Q11, 16, audioBufferSize, 8, cb)
	if err != nil {
		panic(err)
	}
	defer DeInitStream(rxStream)

	_ = SetStreamTimeout(rf, Rx, 32)
	timeout, _ := GetStreamTimeout(rf, Rx)
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
		err = StartStream(rxStream, RxX2)
		if err != nil {
			panic(err)
		}
	}()

	<-sig
	fmt.Println("done")
}
