package bladerf

import (
	"fmt"
	"github.com/erayarslan/go-bladerf/channel_layout"
	"github.com/erayarslan/go-bladerf/direction"
	"github.com/erayarslan/go-bladerf/format"
	"github.com/erayarslan/go-bladerf/gain_mode"
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

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Error)

	PrintVersion(GetVersion())

	bootloaders := GetBootloaderList()
	fmt.Printf("Bootloaders Len: %d\n", len(bootloaders))

	devices := GetDeviceList()
	fmt.Printf("Devices Len: %d\n", len(devices))
	rf, _ := OpenWithDevInfo(devices[0])
	LoadFpga(rf, "/Users/erayarslan/Downloads/hostedxA4-latest.rbf")
	Close(rf)

	rf = Open()
	info := GetDevInfo(&rf)
	Close(rf)
	out, _ := OpenWithDevInfo(GetDevInfo(&rf))
	Close(out)
	Close(Open())
	Close(OpenWithDeviceIdentifier("*:serial=" + info.serial))
	out2, _ :=OpenWithDevInfo(GetDevInfoFromStr("*:serial=" + info.serial))
	Close(out2)

	result := DevInfoMatches(GetDevInfo(&rf), GetDevInfo(&rf))
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	result = DevStrMatches("*:serial="+info.serial, GetDevInfo(&rf))
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	rf = Open()

	err := EnableModule(&rf, CHANNEL_RX(1))
	if err != nil {
		fmt.Println(err)
	}
	_ = InitStream(&rf, format.SC16_Q11, 16, audioBufferSize, 8, cb)
	// _ = StartStream(stream, RX_X1)
}

func TestSetGainStage(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	stages := GetGainStages(&rf, CHANNEL_RX(1))
	fmt.Println(len(stages))
	bfRange, _ := GetGainStageRange(&rf, CHANNEL_RX(1), stages[0])
	_ = SetGainStage(&rf, CHANNEL_RX(1), stages[0], int(bfRange.max))
	gain, _ := GetGainStage(&rf, CHANNEL_RX(1), stages[0])
	fmt.Println(gain)
}

func TestStream(t *testing.T) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()
	channel := CHANNEL_RX(1)

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	_ = SetFrequency(&rf, channel, 96600000)
	min, max, step, _ := GetSampleRateRange(&rf, channel)
	fmt.Printf("Min: %d, Max: %d, Step: %d\n", min, max, step)
	_ = SetSampleRate(&rf, channel, 4e6)
	_ = SyncConfig(&rf, channel_layout.RX_X2, format.SC16_Q11, 16, audioBufferSize, 8, 32)
	actual, _ := SetBandwidth(&rf, channel, 240000)
	fmt.Println(actual)
	_ = EnableModule(&rf, channel)
	_ = SetGainMode(&rf, channel, gain_mode.Hybrid_AGC)

	demodulator = demodcore.MakeWBFMDemodulator(uint32(2e6), 80e3, 48000)

	portaudio.Initialize()
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
			data, _ := SyncRX(&rf, audioBufferSize)
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

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	GetGainModes(&rf, CHANNEL_RX(1))
}

func TestGetGainRange(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	bfRange, _ := GetGainRange(&rf, CHANNEL_RX(1))
	fmt.Println(bfRange.max)
}

func TestGPSData(t *testing.T) {
	log.SetVerbosity(log.Debug)
	channel := CHANNEL_RX(1)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	_ = SetFrequency(&rf, channel, 1525420000)
	_ = SyncConfig(&rf, channel_layout.RX_X2, format.SC16_Q11, 16, audioBufferSize, 8, 32)
	_ = SetSampleRate(&rf, channel, 2600000)
	_, _ = SetBandwidth(&rf, channel, 2500000)
	//_ = SetGainMode(&rf, channel, Hybrid_AGC)
	_ = EnableModule(&rf, channel)

	for {
		data, _ := SyncRX(&rf, audioBufferSize)
		out := GetFinalDataString(data)
		fmt.Println(out)
	}
}

func TestAsyncStream(t *testing.T) {
	log.SetVerbosity(log.Debug)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	channel := CHANNEL_RX(0)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf, _ := OpenWithDevInfo(devices[0])
	defer Close(rf)

	_ = SetFrequency(&rf, channel, 96600000)
	_ = SetSampleRate(&rf, channel, 4e6)
	_, _ = SetBandwidth(&rf, channel, 240000)
	//_ = SetGainMode(&rf, channel, Hybrid_AGC)
	_ = EnableModule(&rf, channel)

	rxStream := InitStream(&rf, format.SC16_Q11, 16, audioBufferSize, 8, cb)
	defer DeInitStream(rxStream)

	_ = SetStreamTimeout(&rf, direction.RX, 32)
	timeout, _ := GetStreamTimeout(&rf, direction.RX)
	println(timeout)

	demodulator = demodcore.MakeWBFMDemodulator(uint32(2e6), 80e3, 48000)

	portaudio.Initialize()
	h, _ := portaudio.DefaultHostApi()

	p := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
	p.Input.Channels = 0
	p.Output.Channels = 1
	p.SampleRate = 48000
	p.FramesPerBuffer = audioBufferSize

	audioStream, _ = portaudio.OpenStream(p, ProcessAudio)
	_ = audioStream.Start()

	go func() {
		_ = StartStream(rxStream, channel_layout.RX_X2)
	}()

	<-sig
	fmt.Println("hehe")
}