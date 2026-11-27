package bladerf

// #include "macro_wrapper.h"
import "C"
import (
	"math"
	"unsafe"

	"github.com/mattn/go-pointer"
)

//export StreamCallback
func StreamCallback(
	dev *C.struct_bladerf,
	stream *C.struct_bladerf_stream,
	metadata *C.struct_bladerf_metadata,
	samples unsafe.Pointer,
	numSamples C.size_t,
	userDataPtr unsafe.Pointer,
) unsafe.Pointer {
	if userDataPtr == nil {
		return StreamShutdown
	}
	restored := pointer.Restore(userDataPtr)
	userData, ok := restored.(UserData)
	if !ok || userData.callback == nil {
		return StreamShutdown
	}

	count := int(numSamples) * 2
	if count > len(userData.results) {
		count = len(userData.results)
	}
	if count > 0 && samples != nil {
		copy(userData.results[:count], (*[1 << 28]int16)(samples)[:count:count])
	}

	switch userData.callback(userData.results[:count]) {
	case GoStreamNoData:
		return StreamNoData
	case GoStreamShutdown:
		return StreamShutdown
	default:
		return samples
	}
}

func (bladeRF *BladeRF) InitStream(
	format Format,
	numBuffers int,
	samplesPerBuffer int,
	numTransfers int,
	callback func(data []int16) GoStream,
) (Stream, error) {
	if !bladeRF.valid() || callback == nil || samplesPerBuffer <= 0 {
		return Stream{}, invalErr()
	}
	var buffers *unsafe.Pointer
	var rxStream *C.struct_bladerf_stream

	userDataPtr := pointer.Save(NewUserData(callback, samplesPerBuffer*2))

	err := GetError(C.bladerf_init_stream(
		&rxStream,
		bladeRF.ref,
		(*[0]byte)(C.StreamCallback),
		&buffers,
		C.size_t(numBuffers),
		C.bladerf_format(format),
		C.size_t(samplesPerBuffer),
		C.size_t(numTransfers),
		userDataPtr,
	))
	if err != nil {
		pointer.Unref(userDataPtr)
		return Stream{}, err
	}

	return Stream{ref: rxStream, userData: userDataPtr}, nil
}

func (stream *Stream) Start(layout ChannelLayout) error {
	if stream == nil || stream.ref == nil {
		return invalErr()
	}
	return GetError(C.bladerf_stream(stream.ref, C.bladerf_channel_layout(layout)))
}

func (stream *Stream) DeInit() {
	if stream == nil {
		return
	}

	if stream.ref != nil {
		C.bladerf_deinit_stream(stream.ref)
		stream.ref = nil
	}

	if stream.userData != nil {
		pointer.Unref(stream.userData)
		stream.userData = nil
	}
}

func (bladeRF *BladeRF) GetStreamTimeout(direction Direction) (uint, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var timeout C.uint
	if err := GetError(C.bladerf_get_stream_timeout(bladeRF.ref, C.bladerf_direction(direction), &timeout)); err != nil {
		return 0, err
	}

	return uint(timeout), nil
}

func (bladeRF *BladeRF) SetStreamTimeout(direction Direction, timeout uint) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_set_stream_timeout(bladeRF.ref, C.bladerf_direction(direction), C.uint(timeout)))
}

func (bladeRF *BladeRF) SyncConfig(
	layout ChannelLayout,
	format Format,
	numBuffers uint,
	bufferSize uint,
	numTransfers uint,
	timeout uint,
) error {
	if !bladeRF.valid() {
		return invalErr()
	}
	return GetError(C.bladerf_sync_config(
		bladeRF.ref,
		C.bladerf_channel_layout(layout),
		C.bladerf_format(format),
		C.uint(numBuffers),
		C.uint(bufferSize),
		C.uint(numTransfers),
		C.uint(timeout),
	))
}

func (bladeRF *BladeRF) SyncTX(input []int16, metadata Metadata, timeout uint) (Metadata, error) {
	if !bladeRF.valid() {
		return metadata, invalErr()
	}
	if len(input) < 2 {
		return metadata, nil
	}

	cMeta := metadata.toC()
	err := GetError(C.bladerf_sync_tx(
		bladeRF.ref,
		unsafe.Pointer(&input[0]),
		C.uint(len(input)/2),
		&cMeta,
		C.uint(timeout),
	))
	if err != nil {
		return metadata, err
	}

	return LoadMetadata(&cMeta), nil
}

// SyncRXInto reads interleaved SC16_Q11 IQ int16 samples directly into dst (length = 2 * numSamples)
// without allocating a new slice per call, returning the number of int16 elements written.
func (bladeRF *BladeRF) SyncRXInto(dst []int16, metadata Metadata, timeout uint) (int, Metadata, error) {
	if !bladeRF.valid() {
		return 0, metadata, invalErr()
	}
	numSamples := len(dst) / 2
	if numSamples == 0 {
		return 0, metadata, nil
	}

	cMeta := metadata.toC()
	err := GetError(C.bladerf_sync_rx(
		bladeRF.ref,
		unsafe.Pointer(&dst[0]),
		C.uint(numSamples),
		&cMeta,
		C.uint(timeout),
	))
	if err != nil {
		return 0, metadata, err
	}

	actual := int(cMeta.actual_count)
	if actual == 0 {
		actual = numSamples
	}
	count := actual * 2
	if count > len(dst) {
		count = len(dst)
	}
	return count, LoadMetadata(&cMeta), nil
}

func (bladeRF *BladeRF) SyncRX(bufferSize uintptr, metadata Metadata, timeout uint) ([]int16, Metadata, error) {
	if !bladeRF.valid() {
		return nil, metadata, invalErr()
	}
	if bufferSize == 0 {
		return make([]int16, 0), metadata, nil
	}

	buf := make([]int16, bufferSize*2)
	count, meta, err := bladeRF.SyncRXInto(buf, metadata, timeout)
	if err != nil {
		return nil, metadata, err
	}
	return buf[:count:count], meta, nil
}

func (bladeRF *BladeRF) GetTimestamp(direction Direction) (Timestamp, error) {
	if !bladeRF.valid() {
		return 0, invalErr()
	}
	var timestamp C.bladerf_timestamp
	if err := GetError(C.bladerf_get_timestamp(bladeRF.ref, C.bladerf_direction(direction), &timestamp)); err != nil {
		return 0, err
	}

	return Timestamp(timestamp), nil
}

func clampSC16Q11(v float32) int16 {
	if math.IsNaN(float64(v)) {
		return 0
	}
	scaled := v * 2048
	if scaled >= 2047 {
		return 2047
	}
	if scaled <= -2048 {
		return -2048
	}
	return int16(scaled)
}

func Complex64ToInt16(input []complex64) []int16 {
	out := make([]int16, len(input)*2)
	for i, c := range input {
		out[2*i] = clampSC16Q11(real(c))
		out[2*i+1] = clampSC16Q11(imag(c))
	}

	return out
}

func Int16ToComplex64(input []int16) []complex64 {
	out := make([]complex64, len(input)/2)
	for i := 0; i < len(out); i++ {
		out[i] = complex(float32(input[2*i])/2048, float32(input[2*i+1])/2048)
	}

	return out
}
