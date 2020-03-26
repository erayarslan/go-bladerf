package bladerf

import (
	"bladerf/log"
	"fmt"
	"github.com/gordonklaus/portaudio"
	log2 "log"
	"math"
	"os"
	"os/signal"
	"testing"
)

type agcState struct {
	gainNum    int32
	gainDen    int32
	gainMax    int32
	peakTarget int
	attackStep int
	decayStep  int
	err        int
}

type demodState struct {
	lowpassed []int16
	rateIn    int
	rateOut   int
	rateOut2  int
	nowR      int16
	nowJ      int16
	preR      int16
	preJ      int16
	prevIndex int
	// min 1, max 256
	downsample     int
	postDownsample int
	outputScale    int
	squelchLevel   int
	conseqSquelch  int
	squelchHits    int
	customAtan     int
	deemph         bool
	deemphA        int
	nowLpr         int
	prevLprIndex   int
	modeDemod      func(fm *demodState)
	agcEnable      bool
	agc            agcState
}

func polarDiscriminant(ar, aj, br, bj int) int {
	var cr, cj int
	var angle float64
	cr = ar*br - aj*-bj
	cj = aj*br + ar*-bj
	angle = math.Atan2(float64(cj), float64(cr))
	return int(angle / math.Pi * (1 << 14))
}

func fastAtan2(y, x int) int {
	var pi4, pi34, yabs, angle int
	pi4 = 1 << 12
	pi34 = 3 * (1 << 12) // note pi = 1<<14
	if x == 0 && y == 0 {
		return 0
	}
	yabs = y
	if yabs < 0 {
		yabs = -yabs
	}
	if x >= 0 {
		angle = pi4 - pi4*(x-yabs)/(x+yabs)
	} else {
		angle = pi34 - pi4*(x+yabs)/(yabs-x)
	}
	if y < 0 {
		return -angle
	}
	return angle
}

func polarDiscFast(ar, aj, br, bj int) int {
	var cr, cj int
	cr = ar*br - aj*-bj
	cj = aj*br + ar*-bj
	return fastAtan2(cj, cr)
}

var d = demodState{
	rateIn:       170000,
	rateOut:      170000,
	rateOut2:     32000,
	customAtan:   1,
	deemph:       true,
	squelchLevel: 0,
}

func fmDemod(fm *demodState) {
	var i, pcm int
	lp := fm.lowpassed
	lpLen := len(fm.lowpassed)
	pr := fm.preR
	pj := fm.preJ
	for i = 2; i < (lpLen - 1); i += 2 {
		switch fm.customAtan {
		case 0:
			pcm = polarDiscriminant(int(lp[i]), int(lp[i+1]), int(pr), int(pj))
		case 1:
			pcm = polarDiscFast(int(lp[i]), int(lp[i+1]), int(pr), int(pj))
		}
		pr = lp[i]
		pj = lp[i+1]

		fm.lowpassed[i/2] = int16(pcm)
	}
	fm.preR = pr
	fm.preJ = pj
	fm.lowpassed = fm.lowpassed[:lpLen/2]
}

func lowPass(d *demodState) {
	var i, i2 int
	for i < len(d.lowpassed) {
		d.nowR += d.lowpassed[i]
		d.nowJ += d.lowpassed[i+1]
		i += 2
		d.prevIndex++
		if d.prevIndex < d.downsample {
			continue
		}
		d.lowpassed[i2] = d.nowR   // * d.output_scale;
		d.lowpassed[i2+1] = d.nowJ // * d.output_scale;
		d.prevIndex = 0
		d.nowR = 0
		d.nowJ = 0
		i2 += 2
	}
	d.lowpassed = d.lowpassed[:i2]
}

func TestBladeRF(t *testing.T) {
	log.SetVerbosity(log.Critical)

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
	StartStream(stream, RX_X1)
	Close(rf)
}

type echo struct {
	*portaudio.Stream
	buffer []float32
	i      int
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

	err = SetFrequency(&rf, IORX, 90800000)
	if err != nil {
		log2.Fatal(err)
	}

	min, max, step, errRange := GetSampleRateRange(&rf, IORX)
	if errRange != nil {
		log2.Fatal(err)
	} else {
		fmt.Printf("Min: %d, Max: %d, Step: %d\n", min, max, step)
	}

	err = SetSampleRate(&rf, IORX, max)
	if err != nil {
		log2.Fatal(err)
	}

	err = SyncConfig(&rf, RX_X1, SC16_Q11, 32, 8192, 8, 3500)
	if err != nil {
		log2.Fatal(err)
	}

	err = EnableModule(&rf, RX)
	if err != nil {
		log2.Fatal(err)
	}

	err = SetGain(&rf, IORX, 9)
	if err != nil {
		log2.Fatal(err)
	}

	p := make([]int16, 10000)

	results := SyncRX(&rf)
	//var complexFloat = make([]float32, 10000)

	//for i := 0; i < 10000 ; i++ {
	//	complexFloat[i] = float32(results[2*i])/float32(2048) + float32(results[2*i+1])/float32(2048.0)
	//}

	d.lowpassed = results
	fmDemod(&d)
	lowPass(&d)

	err = DisableModule(&rf, RX)
	if err != nil {
		log2.Fatal(err)
	}

	out := make([]int16, 8192)

	e := &echo{buffer: make([]float32, 10000)}

	e.Stream, err = portaudio.OpenDefaultStream(0, 1, 10000, len(d.lowpassed), &out)
	defer stream.Close()

	stream.Start()

	defer stream.Close()
	defer stream.Stop()

	for {
		out = d.lowpassed
		stream.Write()
		select {
		case <-sig:
			return
		default:
		}
	}

}

func TestAsyncStream(t *testing.T) {
	log.SetVerbosity(log.Critical)

	devices := GetDeviceList()

	if len(devices) == 0 {
		fmt.Println("NO DEVICE")
		return
	}

	rf := OpenWithDevInfo(devices[0])
	defer Close(rf)

	SetSampleRate(&rf, IORX, 1000000)
	EnableModule(&rf, RX)

	rxStream := InitStream(&rf, SC16_Q11, 32, 32768, 16)
	defer DeInitStream(rxStream)

	SetStreamTimeout(&rf, RX, 16)
	GetStreamTimeout(&rf, RX)

	StartStream(rxStream, RX_X1)
}
