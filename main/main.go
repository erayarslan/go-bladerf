package main

import (
	bf "bladerf"
	"bladerf/log"
	"fmt"
	"github.com/gordonklaus/portaudio"
	fifo "github.com/racerxdl/go.fifo"
	"github.com/racerxdl/segdsp/demodcore"
	"os"
	"os/signal"
	"syscall"
)


const audioBufferSize = 8192 / 4
var demodulator demodcore.DemodCore
var audioStream *portaudio.Stream
var audioFifo = fifo.NewQueue()

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

func main() {
	log.SetVerbosity(log.Debug)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	channel := bf.CHANNEL_RX(1)

	devices := bf.GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := bf.OpenWithDevInfo(devices[0])
	defer bf.Close(rf)

	_ = bf.SetFrequency(&rf, channel, 96600000)
	_ = bf.SetSampleRate(&rf, channel, 4e6)
	_, _ = bf.SetBandwidth(&rf, channel, 240000)
	//_ = SetGainMode(&rf, channel, Hybrid_AGC)
	_ = bf.EnableModule(&rf, channel)

	rxStream := bf.InitStream(&rf, bf.SC16_Q11, 16, audioBufferSize, 8, cb)
	defer bf.DeInitStream(rxStream)

	_ = bf.SetStreamTimeout(&rf, bf.RX, 32)
	timeout, _ := bf.GetStreamTimeout(&rf, bf.RX)
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
		_ = bf.StartStream(rxStream, bf.RX_X2)
	}()

	<-sig
}
