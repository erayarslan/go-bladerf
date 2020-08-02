package bladerf

import (
	"bladerf/log"
	"fmt"
	"github.com/gordonklaus/portaudio"
	fifo "github.com/racerxdl/go.fifo"
	"github.com/racerxdl/segdsp/demodcore"
	log2 "log"
	"os"
	"os/signal"
	"testing"
)

const audioBufferSize = 8192 / 2

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

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Debug)

	PrintVersion(GetVersion())

	devices := GetDeviceList()
	fmt.Printf("Devices Len: %d\n", len(devices))
	rf := OpenWithDevInfo(devices[0])
	LoadFpga(rf, "/Users/erayarslan/Downloads/hostedxA4-latest.rbf")
	Close(rf)

	rf = Open()
	info := GetDevInfo(&rf)
	Close(rf)
	Close(OpenWithDevInfo(GetDevInfo(&rf)))
	Close(Open())
	Close(OpenWithDeviceIdentifier("*:serial=" + info.serial))
	Close(OpenWithDevInfo(GetDevInfoFromStr("*:serial=" + info.serial)))

	result := DevInfoMatches(GetDevInfo(&rf), GetDevInfo(&rf))
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	result = DevStrMatches("*:serial=" + info.serial, GetDevInfo(&rf))
	fmt.Println("---------")
	fmt.Println(result)
	fmt.Println("---------")

	rf = Open()

	bootloaders := GetBootloaderList()
	fmt.Printf("Bootloaders Len: %d\n", len(bootloaders))

	err := EnableModule(&rf, RX)
	if err != nil {
		fmt.Println(err)
	}
	_ = InitStream(&rf, SC16_Q11, 16, audioBufferSize, 8, cb)
	// _ = StartStream(stream, RX_X1)
}

func TestSetGainStage(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	stages := GetGainStages(&rf, IORX)
	fmt.Println(len(stages))
	bfRange, _ := GetGainStageRange(&rf, IORX, stages[0])
	_ = SetGainStage(&rf, IORX, stages[0], int(bfRange.max))
	gain, _ := GetGainStage(&rf, IORX, stages[0])
	fmt.Println(gain)
}

func TestStream(t *testing.T) {
	var err error

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	err = SetFrequency(&rf, IORX, 96600000)
	if err != nil {
		log2.Fatal(err)
	}

	min, max, step, err := GetSampleRateRange(&rf, IORX)
	if err != nil {
		log2.Fatal(err)
	}
	fmt.Printf("Min: %d, Max: %d, Step: %d\n", min, max, step)
	err = SetSampleRate(&rf, IORX, 4e6)
	if err != nil {
		log2.Fatal(err)
	}
	err = SyncConfig(&rf, RX_X1, SC16_Q11, 16, audioBufferSize, 8, 32)
	if err != nil {
		log2.Fatal(err)
	}

	actual, err := SetBandwidth(&rf, IORX, 240000)
	if err != nil {
		log2.Fatal(err)
	} else {
		println(actual)
	}

	err = EnableModule(&rf, RX)
	if err != nil {
		log2.Fatal(err)
	}

	err = SetGainMode(&rf, IORX, Hybrid_AGC)
	if err != nil {
		log2.Fatal(err)
	}

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

	for {
		out := demodulator.Work(GetFinalData(SyncRX(&rf, audioBufferSize)))

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

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	GetGainModes(&rf, IORX)
}

func TestGetGainRange(t *testing.T) {
	log.SetVerbosity(log.Debug)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	bfRange, _ := GetGainRange(&rf, IORX)
	fmt.Println(bfRange.max)
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

	_ = SetFrequency(&rf, IORX, 96600000)
	_ = SetSampleRate(&rf, IORX, 4e6)
	_, _ = SetBandwidth(&rf, IORX, 240000)
	_ = SetGainMode(&rf, IORX, Hybrid_AGC)
	_ = EnableModule(&rf, RX)

	rxStream := InitStream(&rf, SC16_Q11, 16, audioBufferSize, 8, cb)
	defer DeInitStream(rxStream)

	_ = SetStreamTimeout(&rf, RX, 32)
	timeout, _ := GetStreamTimeout(&rf, RX)
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

	_ = StartStream(rxStream, RX_X1)
}
