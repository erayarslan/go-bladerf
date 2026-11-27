package mock

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"
)

const (
	// DefaultPort is the default TCP port for bladerf-remote / bladerf-mock.
	DefaultPort = 5555

	protoMagic = 0x42524652 // "BRFR"
	maxPayload = 32 * 1024 * 1024
	maxStr     = 32
	maxModes   = 16
	maxList    = 16
	flashBytes = 4194304
)

// Opcodes matching host/libraries/libbladeRF/src/backend/remote/remote_proto.h.
const (
	opHello = 1
	opClose = 2

	opSetFrequency           = 10
	opGetFrequency           = 11
	opGetFrequencyRange      = 12
	opSelectBand             = 13
	opGetQuickTune           = 14
	opScheduleRetune         = 15
	opCancelScheduledRetunes = 16
	opSetTuningMode          = 17
	opGetTuningMode          = 18
	opSetSampleRate          = 20
	opGetSampleRate          = 21
	opSetRationalSampleRate  = 22
	opGetRationalSampleRate  = 23
	opGetSampleRateRange     = 24
	opSetBandwidth           = 25
	opGetBandwidth           = 26
	opGetBandwidthRange      = 27
	opSetGain                = 30
	opGetGain                = 31
	opGetGainRange           = 32
	opSetGainMode            = 33
	opGetGainMode            = 34
	opGetGainModes           = 35
	opSetGainStage           = 36
	opGetGainStage           = 37
	opGetGainStageRange      = 38
	opGetGainStages          = 39
	opSetCorrection          = 40
	opGetCorrection          = 41
	opEnableModule           = 50
	opSetRxMux               = 51
	opGetRxMux               = 52
	opSetLoopback            = 53
	opGetLoopback            = 54
	opGetLoopbackModes       = 55
	opSetRfPort              = 56
	opGetRfPort              = 57
	opGetRfPorts             = 58
	opSetStreamTimeout       = 60
	opGetStreamTimeout       = 61
	opGetTimestamp           = 62
	opSyncConfig             = 63
	opSyncRx                 = 64
	opSyncTx                 = 65
	opStreamStart            = 66
	opStreamStop             = 67
	opTriggerInit            = 70
	opTriggerArm             = 71
	opTriggerFire            = 72
	opTriggerState           = 73
	opReadTrigger            = 74
	opWriteTrigger           = 75
	opConfigGpioRead         = 76
	opConfigGpioWrite        = 77
	opGetVctcxoTrim          = 78
	opTrimDacRead            = 79
	opTrimDacWrite           = 80
	opSetVctcxoTamerMode     = 81
	opGetVctcxoTamerMode     = 82
	opExpansionAttach        = 83
	opExpansionGetAttached   = 84
	opEraseFlash             = 90
	opReadFlash              = 91
	opWriteFlash             = 92
	opReadOtp                = 93
	opWriteOtp               = 94
	opLockOtp                = 95
	opLoadFpga               = 96
)

const (
	errInval       int32 = -3
	errUnsupported int32 = -8
)

const (
	speedHigh      int32 = 1
	fpgaSizeA4     int32 = 49
	fpgaSourceHost int32 = 2

	corrDcoffI int32 = 0
	corrDcoffQ int32 = 1

	gainDefault int32 = 0
	gainMgc     int32 = 1
	gainFast    int32 = 2
	gainSlow    int32 = 3
	gainHybrid  int32 = 4

	lbNone     int32 = 0
	lbFirmware int32 = 1
	lbRficBist int32 = 9

	layoutRxX1 int32 = 0
	layoutTxX1 int32 = 1
	layoutRxX2 int32 = 2
	layoutTxX2 int32 = 3

	formatSc8Q7       int32 = 3
	formatSc8Q7Meta   int32 = 4
	formatSc16Q11Pack int32 = 1

	tuningModeHost int32 = 0
)

type connState struct {
	frequency     [4]uint64
	sampleRate    [4]uint32
	rationalNum   [4]uint64
	rationalDen   [4]uint64
	bandwidth     [4]uint32
	gain          [4]int32
	gainMode      [4]int32
	stageGain     [4]int32
	corr          [4][4]int16
	enabled       [4]bool
	rfPort        [4]string
	triggerReg    [4]uint8
	rxMux         int32
	loopback      int32
	tuningMode    int32
	streamTimeout [2]uint32
	configGpio    uint32
	trimDac       uint16
	timestamp     [2]uint64
	phase         float64
	otp           [256]byte
	flash         []byte
	trigArmed     bool
	trigFired     bool
	trigFireReq   bool
	trigResv1     uint64
	trigResv2     uint64
}

func newConnState() *connState {
	st := &connState{
		rxMux:      0,
		loopback:   lbNone,
		tuningMode: tuningModeHost,
		configGpio: 129,
		trimDac:    8091,
	}
	for i := 0; i < 4; i++ {
		st.frequency[i] = 100000000
		st.sampleRate[i] = 30720000
		st.rationalNum[i] = 0
		st.rationalDen[i] = 1
		st.bandwidth[i] = 1500000
		st.gain[i] = 60
		st.gainMode[i] = gainDefault
		st.stageGain[i] = 71
		st.corr[i][0] = -1
		st.corr[i][1] = -1
		if i == 1 || i == 3 {
			st.rfPort[i] = "TX1"
		} else {
			st.rfPort[i] = "A_BALANCED"
		}
	}
	return st
}

func (c *connState) ensureFlash() []byte {
	if c.flash == nil {
		c.flash = make([]byte, flashBytes)
		for i := range c.flash {
			c.flash[i] = 0xFF
		}
	}
	return c.flash
}

func (c *connState) fillSimIQ(dst []byte, numSamples uint32) {
	const dphase = 2.0 * math.Pi * 0.05
	for i := uint32(0); i < numSamples; i++ {
		iSample := int16(1024.0 * math.Cos(c.phase))
		qSample := int16(1024.0 * math.Sin(c.phase))
		off := int(i) * 4
		binary.LittleEndian.PutUint16(dst[off:off+2], uint16(iSample))
		binary.LittleEndian.PutUint16(dst[off+2:off+4], uint16(qSample))
		c.phase += dphase
		if c.phase >= 2.0*math.Pi {
			c.phase -= 2.0 * math.Pi
		}
	}
	c.timestamp[0] += uint64(numSamples)
}

func chIdx(ch int32) int {
	switch ch {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	default:
		return 0
	}
}

func dirIdx(dir int32) int {
	if dir == 1 {
		return 1
	}
	return 0
}

func isTxChannel(ch int32) bool {
	return (ch & 1) != 0
}

func formatSampleBytes(format int32, numSamples uint32) int {
	switch format {
	case formatSc8Q7, formatSc8Q7Meta:
		return int(numSamples) * 2
	case formatSc16Q11Pack:
		return int(numSamples) * 3
	default:
		return int(numSamples) * 4
	}
}

// Server implements a pure-Go simulated bladeRF 2.0 xA4 TCP server compatible with libbladeRF's remote backend.
type Server struct {
	ln     net.Listener
	mu     sync.Mutex
	closed bool
	conns  map[net.Conn]struct{}
	wg     sync.WaitGroup
}

// NewServer starts a mock bladeRF TCP server listening on listenAddr (e.g. "127.0.0.1:0" or "0.0.0.0:5555").
func NewServer(listenAddr string) (*Server, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("listen %q: %w", listenAddr, err)
	}
	s := &Server{
		ln:    ln,
		conns: make(map[net.Conn]struct{}),
	}
	s.wg.Add(1)
	go s.acceptLoop()
	return s, nil
}

// Addr returns the server's bound host:port address.
func (s *Server) Addr() string {
	return s.ln.Addr().String()
}

// Close stops accepting connections, closes active clients, and waits for server goroutines to exit.
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.wg.Wait()
		return nil
	}
	s.closed = true
	err := s.ln.Close()
	for c := range s.conns {
		_ = c.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	return err
}

func (s *Server) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func (s *Server) trackConn(c net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.conns[c] = struct{}{}
	return true
}

func (s *Server) untrackConn(c net.Conn) {
	s.mu.Lock()
	delete(s.conns, c)
	s.mu.Unlock()
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		c, err := s.ln.Accept()
		if err != nil {
			if s.isClosed() {
				return
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return
		}
		if !s.trackConn(c) {
			_ = c.Close()
			continue
		}
		s.wg.Add(1)
		go func(conn net.Conn) {
			defer s.wg.Done()
			defer s.untrackConn(conn)
			defer conn.Close()
			s.serveConn(conn)
		}(c)
	}
}

func sendReply(w io.Writer, opcode uint16, status int32, payload []byte) error {
	var hdr [16]byte
	binary.LittleEndian.PutUint32(hdr[0:4], protoMagic)
	binary.LittleEndian.PutUint16(hdr[4:6], opcode)
	binary.LittleEndian.PutUint16(hdr[6:8], 0)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(status))
	binary.LittleEndian.PutUint32(hdr[12:16], uint32(len(payload)))
	if err := writeAll(w, hdr[:]); err != nil {
		return err
	}
	if len(payload) > 0 {
		return writeAll(w, payload)
	}
	return nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

func putCString(dst []byte, s string) {
	for i := range dst {
		dst[i] = 0
	}
	if len(dst) == 0 {
		return
	}
	n := len(s)
	if n > len(dst)-1 {
		n = len(dst) - 1
	}
	copy(dst[:n], s[:n])
}

func (s *Server) serveConn(conn net.Conn) {
	st := newConnState()
	var hdrBuf [16]byte

	for {
		if _, err := io.ReadFull(conn, hdrBuf[:]); err != nil {
			return
		}
		magic := binary.LittleEndian.Uint32(hdrBuf[0:4])
		opcode := binary.LittleEndian.Uint16(hdrBuf[4:6])
		payloadLen := binary.LittleEndian.Uint32(hdrBuf[12:16])

		if magic != protoMagic || payloadLen > maxPayload {
			return
		}

		var payload []byte
		if payloadLen > 0 {
			payload = make([]byte, payloadLen)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return
			}
		}

		done, err := s.handleRequest(conn, st, opcode, payload)
		if err != nil || done {
			return
		}
	}
}

func encodeHello() []byte {
	var resp [246]byte
	putCString(resp[0:16], "bladerf2")
	putCString(resp[16:49], "00000000000000000000000000000001")
	putCString(resp[49:82], "Nuand")
	putCString(resp[82:115], "bladeRF 2.0")
	binary.LittleEndian.PutUint32(resp[115:119], uint32(speedHigh))
	binary.LittleEndian.PutUint32(resp[119:123], uint32(fpgaSizeA4))
	binary.LittleEndian.PutUint32(resp[123:127], uint32(fpgaSourceHost))
	binary.LittleEndian.PutUint32(resp[127:131], 1)
	binary.LittleEndian.PutUint64(resp[131:139], 2632660)
	binary.LittleEndian.PutUint32(resp[139:143], flashBytes)
	resp[143] = 0
	binary.LittleEndian.PutUint32(resp[144:148], 256)
	binary.LittleEndian.PutUint32(resp[148:152], 65536)
	binary.LittleEndian.PutUint16(resp[152:154], 2)
	binary.LittleEndian.PutUint16(resp[154:156], 3)
	binary.LittleEndian.PutUint16(resp[156:158], 2)
	putCString(resp[158:191], "2.3.2")
	binary.LittleEndian.PutUint16(resp[191:193], 0)
	binary.LittleEndian.PutUint16(resp[193:195], 11)
	binary.LittleEndian.PutUint16(resp[195:197], 0)
	putCString(resp[197:230], "0.11.0")
	binary.LittleEndian.PutUint64(resp[230:238], 0xFFFFFFFFFFFFFFFF)
	binary.LittleEndian.PutUint32(resp[238:242], 2)
	binary.LittleEndian.PutUint32(resp[242:246], 2)
	return resp[:]
}

func encodeRange(min, max, step int64, scale float32) []byte {
	var resp [28]byte
	binary.LittleEndian.PutUint64(resp[0:8], uint64(min))
	binary.LittleEndian.PutUint64(resp[8:16], uint64(max))
	binary.LittleEndian.PutUint64(resp[16:24], uint64(step))
	binary.LittleEndian.PutUint32(resp[24:28], math.Float32bits(scale))
	return resp[:]
}

type modeEntry struct {
	mode int32
	name string
}

func encodeModeList(items []modeEntry) []byte {
	var resp [580]byte
	n := len(items)
	if n > maxModes {
		n = maxModes
	}
	binary.LittleEndian.PutUint32(resp[0:4], uint32(n))
	for i := 0; i < n; i++ {
		off := 4 + i*36
		binary.LittleEndian.PutUint32(resp[off:off+4], uint32(items[i].mode))
		putCString(resp[off+4:off+36], items[i].name)
	}
	return resp[:]
}

func encodeStringList(items []string) []byte {
	var resp [516]byte
	n := len(items)
	if n > maxList {
		n = maxList
	}
	binary.LittleEndian.PutUint32(resp[0:4], uint32(n))
	for i := 0; i < n; i++ {
		off := 4 + i*maxStr
		putCString(resp[off:off+maxStr], items[i])
	}
	return resp[:]
}

func (s *Server) handleRequest(conn net.Conn, st *connState, opcode uint16, payload []byte) (bool, error) {
	switch opcode {
	case opHello:
		return false, sendReply(conn, opcode, 0, encodeHello())

	case opClose:
		_ = sendReply(conn, opcode, 0, nil)
		return true, nil

	case opSetFrequency:
		if len(payload) < 12 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.frequency[chIdx(ch)] = binary.LittleEndian.Uint64(payload[4:12])
		return false, sendReply(conn, opcode, 0, nil)

	case opGetFrequency:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [8]byte
		binary.LittleEndian.PutUint64(resp[0:8], st.frequency[chIdx(ch)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetFrequencyRange:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeRange(70000000, 6000000000, 1, 1.0))

	case opSelectBand, opCancelScheduledRetunes:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, nil)

	case opGetQuickTune:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		var resp [16]byte
		return false, sendReply(conn, opcode, 0, resp[:])

	case opScheduleRetune:
		if len(payload) < 37 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		freq := binary.LittleEndian.Uint64(payload[12:20])
		st.frequency[chIdx(ch)] = freq
		return false, sendReply(conn, opcode, 0, nil)

	case opSetTuningMode:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.tuningMode = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetTuningMode:
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.tuningMode))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opSetSampleRate:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		rate := binary.LittleEndian.Uint32(payload[4:8])
		idx := chIdx(ch)
		st.sampleRate[idx] = rate
		st.rationalNum[idx] = 0
		st.rationalDen[idx] = 1
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], rate)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetSampleRate:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], st.sampleRate[chIdx(ch)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opSetRationalSampleRate:
		if len(payload) < 28 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		integer := binary.LittleEndian.Uint64(payload[4:12])
		num := binary.LittleEndian.Uint64(payload[12:20])
		den := binary.LittleEndian.Uint64(payload[20:28])
		if den == 0 {
			den = 1
		}
		idx := chIdx(ch)
		st.sampleRate[idx] = uint32(integer + num/den)
		st.rationalNum[idx] = num
		st.rationalDen[idx] = den
		var resp [24]byte
		binary.LittleEndian.PutUint64(resp[0:8], integer)
		binary.LittleEndian.PutUint64(resp[8:16], num)
		binary.LittleEndian.PutUint64(resp[16:24], den)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetRationalSampleRate:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		idx := chIdx(ch)
		var resp [24]byte
		binary.LittleEndian.PutUint64(resp[0:8], uint64(st.sampleRate[idx]))
		binary.LittleEndian.PutUint64(resp[8:16], st.rationalNum[idx])
		binary.LittleEndian.PutUint64(resp[16:24], st.rationalDen[idx])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetSampleRateRange:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeRange(520834, 61440000, 1, 1.0))

	case opSetBandwidth:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		bw := binary.LittleEndian.Uint32(payload[4:8])
		st.bandwidth[chIdx(ch)] = bw
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], bw)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetBandwidth:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], st.bandwidth[chIdx(ch)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetBandwidthRange:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeRange(200000, 56000000, 1, 1.0))

	case opSetGain:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.gain[chIdx(ch)] = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetGain:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.gain[chIdx(ch)]))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetGainRange:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeRange(-15, 60, 1, 1.0))

	case opSetGainMode:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.gainMode[chIdx(ch)] = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetGainMode:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.gainMode[chIdx(ch)]))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetGainModes:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		if isTxChannel(ch) {
			return false, sendReply(conn, opcode, 0, encodeModeList(nil))
		}
		return false, sendReply(conn, opcode, 0, encodeModeList([]modeEntry{
			{gainDefault, "automatic"},
			{gainMgc, "manual"},
			{gainFast, "fastattack_agc"},
			{gainSlow, "slowattack_agc"},
			{gainHybrid, "hybrid_agc"},
		}))

	case opSetGainStage:
		if len(payload) < 40 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.stageGain[chIdx(ch)] = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetGainStage:
		if len(payload) < 40 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.stageGain[chIdx(ch)]))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetGainStageRange:
		if len(payload) < 40 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeRange(-4, 71, 1, 1.0))

	case opGetGainStages:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, encodeStringList([]string{"full"}))

	case opSetCorrection:
		if len(payload) < 10 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		corr := int32(binary.LittleEndian.Uint32(payload[4:8]))
		val := int16(binary.LittleEndian.Uint16(payload[8:10]))
		cidx := 0
		if corr >= 0 && corr < 4 {
			cidx = int(corr)
		}
		if corr == corrDcoffI || corr == corrDcoffQ {
			val = int16(uint16(val)<<4) >> 4
		}
		st.corr[chIdx(ch)][cidx] = val
		return false, sendReply(conn, opcode, 0, nil)

	case opGetCorrection:
		if len(payload) < 10 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		corr := int32(binary.LittleEndian.Uint32(payload[4:8]))
		cidx := 0
		if corr >= 0 && corr < 4 {
			cidx = int(corr)
		}
		var resp [2]byte
		binary.LittleEndian.PutUint16(resp[0:2], uint16(st.corr[chIdx(ch)][cidx]))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opEnableModule:
		if len(payload) < 5 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.enabled[chIdx(ch)] = payload[4] != 0
		return false, sendReply(conn, opcode, 0, nil)

	case opSetRxMux:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.rxMux = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetRxMux:
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.rxMux))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opSetLoopback:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.loopback = int32(binary.LittleEndian.Uint32(payload[4:8]))
		return false, sendReply(conn, opcode, 0, nil)

	case opGetLoopback:
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.loopback))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetLoopbackModes:
		return false, sendReply(conn, opcode, 0, encodeModeList([]modeEntry{
			{lbNone, "none"},
			{lbFirmware, "firmware"},
			{lbRficBist, "rf_bist"},
		}))

	case opSetRfPort:
		if len(payload) < 36 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.rfPort[chIdx(ch)] = cString(payload[4:36])
		return false, sendReply(conn, opcode, 0, nil)

	case opGetRfPort:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [maxStr]byte
		putCString(resp[:], st.rfPort[chIdx(ch)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetRfPorts:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		if isTxChannel(ch) {
			return false, sendReply(conn, opcode, 0, encodeStringList([]string{"TX1", "TX2"}))
		}
		return false, sendReply(conn, opcode, 0, encodeStringList([]string{"A_BALANCED", "B_BALANCED", "C_BALANCED"}))

	case opSetStreamTimeout:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		dir := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.streamTimeout[dirIdx(dir)] = binary.LittleEndian.Uint32(payload[4:8])
		return false, sendReply(conn, opcode, 0, nil)

	case opGetStreamTimeout:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		dir := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], st.streamTimeout[dirIdx(dir)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opGetTimestamp:
		if len(payload) < 4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		dir := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [8]byte
		binary.LittleEndian.PutUint64(resp[0:8], st.timestamp[dirIdx(dir)])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opSyncConfig:
		if len(payload) < 24 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		return false, sendReply(conn, opcode, 0, nil)

	case opSyncRx:
		if len(payload) < 25 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		numSamples := binary.LittleEndian.Uint32(payload[0:4])
		byteLen := binary.LittleEndian.Uint32(payload[21:25])
		if byteLen > maxPayload || uint32(byteLen) < numSamples*4 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		frame := make([]byte, 24+int(byteLen))
		st.fillSimIQ(frame[24:], numSamples)
		binary.LittleEndian.PutUint64(frame[0:8], st.timestamp[0])
		binary.LittleEndian.PutUint32(frame[16:20], numSamples)
		binary.LittleEndian.PutUint32(frame[20:24], byteLen)
		return false, sendReply(conn, opcode, 0, frame)

	case opSyncTx:
		if len(payload) < 25 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		numSamples := binary.LittleEndian.Uint32(payload[0:4])
		byteLen := binary.LittleEndian.Uint32(payload[21:25])
		if uint64(25)+uint64(byteLen) > uint64(len(payload)) {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.timestamp[1] += uint64(numSamples)
		var resp [24]byte
		binary.LittleEndian.PutUint64(resp[0:8], st.timestamp[1])
		binary.LittleEndian.PutUint32(resp[16:20], numSamples)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opStreamStart:
		if len(payload) < 24 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		layout := int32(binary.LittleEndian.Uint32(payload[0:4]))
		format := int32(binary.LittleEndian.Uint32(payload[4:8]))
		samplesPerBuffer := binary.LittleEndian.Uint32(payload[8:12])
		s.runStreamSession(conn, st, layout, format, samplesPerBuffer)
		return true, nil

	case opTriggerInit:
		if len(payload) < 9 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		var resp [20]byte
		copy(resp[0:4], payload[0:4])
		binary.LittleEndian.PutUint32(resp[4:8], 0)
		copy(resp[8:12], payload[4:8])
		return false, sendReply(conn, opcode, 0, resp[:])

	case opTriggerArm:
		if len(payload) < 37 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.trigArmed = payload[20] != 0
		st.trigResv1 = binary.LittleEndian.Uint64(payload[21:29])
		st.trigResv2 = binary.LittleEndian.Uint64(payload[29:37])
		return false, sendReply(conn, opcode, 0, nil)

	case opTriggerFire:
		if len(payload) < 20 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.trigFired = true
		st.trigFireReq = true
		return false, sendReply(conn, opcode, 0, nil)

	case opTriggerState:
		if len(payload) < 20 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		var resp [19]byte
		if st.trigArmed {
			resp[0] = 1
		}
		if st.trigFired {
			resp[1] = 1
		}
		if st.trigFireReq {
			resp[2] = 1
		}
		binary.LittleEndian.PutUint64(resp[3:11], st.trigResv1)
		binary.LittleEndian.PutUint64(resp[11:19], st.trigResv2)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opReadTrigger:
		if len(payload) < 9 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], uint32(st.triggerReg[chIdx(ch)]))
		return false, sendReply(conn, opcode, 0, resp[:])

	case opWriteTrigger:
		if len(payload) < 9 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		ch := int32(binary.LittleEndian.Uint32(payload[0:4]))
		st.triggerReg[chIdx(ch)] = payload[8]
		return false, sendReply(conn, opcode, 0, nil)

	case opConfigGpioRead:
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], st.configGpio)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opConfigGpioWrite:
		if len(payload) < 8 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.configGpio = binary.LittleEndian.Uint32(payload[4:8]) | 129
		return false, sendReply(conn, opcode, 0, nil)

	case opGetVctcxoTrim, opTrimDacRead:
		var resp [2]byte
		binary.LittleEndian.PutUint16(resp[0:2], st.trimDac)
		return false, sendReply(conn, opcode, 0, resp[:])

	case opTrimDacWrite:
		if len(payload) < 2 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		st.trimDac = binary.LittleEndian.Uint16(payload[0:2])
		return false, sendReply(conn, opcode, 0, nil)

	case opSetVctcxoTamerMode:
		return false, sendReply(conn, opcode, errUnsupported, nil)

	case opGetVctcxoTamerMode:
		var resp [4]byte
		binary.LittleEndian.PutUint32(resp[0:4], 0xFFFFFFFF)
		return false, sendReply(conn, opcode, errUnsupported, resp[:])

	case opExpansionAttach:
		return false, sendReply(conn, opcode, errUnsupported, nil)

	case opExpansionGetAttached:
		var resp [4]byte
		return false, sendReply(conn, opcode, 0, resp[:])

	case opEraseFlash:
		if len(payload) < 12 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		eb := binary.LittleEndian.Uint32(payload[0:4])
		count := binary.LittleEndian.Uint32(payload[4:8])
		off := uint64(eb) * 65536
		length := uint64(count) * 65536
		if off+length > flashBytes {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		flash := st.ensureFlash()
		for i := off; i < off+length; i++ {
			flash[i] = 0xFF
		}
		return false, sendReply(conn, opcode, 0, nil)

	case opReadFlash:
		if len(payload) < 12 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		page := binary.LittleEndian.Uint32(payload[0:4])
		byteLen := binary.LittleEndian.Uint32(payload[8:12])
		off := uint64(page) * 256
		if uint64(byteLen) > maxPayload || off+uint64(byteLen) > flashBytes {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		flash := st.ensureFlash()
		return false, sendReply(conn, opcode, 0, flash[off:off+uint64(byteLen)])

	case opWriteFlash:
		if len(payload) < 12 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		page := binary.LittleEndian.Uint32(payload[0:4])
		byteLen := binary.LittleEndian.Uint32(payload[8:12])
		off := uint64(page) * 256
		if uint64(12)+uint64(byteLen) > uint64(len(payload)) || off+uint64(byteLen) > flashBytes {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		flash := st.ensureFlash()
		copy(flash[off:off+uint64(byteLen)], payload[12:12+byteLen])
		return false, sendReply(conn, opcode, 0, nil)

	case opReadOtp:
		return false, sendReply(conn, opcode, 0, st.otp[:])

	case opWriteOtp:
		if len(payload) < 256 {
			return false, sendReply(conn, opcode, errInval, nil)
		}
		copy(st.otp[:], payload[:256])
		return false, sendReply(conn, opcode, 0, nil)

	case opLockOtp:
		return false, sendReply(conn, opcode, 0, nil)

	case opLoadFpga:
		return false, sendReply(conn, opcode, errUnsupported, nil)

	default:
		return false, sendReply(conn, opcode, errUnsupported, nil)
	}
}

func (s *Server) runStreamSession(conn net.Conn, st *connState, layout int32, format int32, samplesPerBuffer uint32) {
	bufBytes := formatSampleBytes(format, samplesPerBuffer)
	if bufBytes <= 0 || bufBytes > maxPayload {
		_ = sendReply(conn, opStreamStart, errInval, nil)
		return
	}
	if err := sendReply(conn, opStreamStart, 0, nil); err != nil {
		return
	}

	isTx := layout == layoutTxX1 || layout == layoutTxX2
	if !isTx {
		frame := make([]byte, 28+bufBytes)
		for !s.isClosed() {
			st.fillSimIQ(frame[28:], samplesPerBuffer)
			binary.LittleEndian.PutUint32(frame[0:4], 0)
			binary.LittleEndian.PutUint64(frame[4:12], st.timestamp[0])
			binary.LittleEndian.PutUint32(frame[12:16], 0)
			binary.LittleEndian.PutUint32(frame[16:20], 0)
			binary.LittleEndian.PutUint32(frame[20:24], samplesPerBuffer)
			binary.LittleEndian.PutUint32(frame[24:28], uint32(bufBytes))
			if err := writeAll(conn, frame); err != nil {
				return
			}
			time.Sleep(time.Millisecond)
		}
		return
	}

	var fhdr [28]byte
	buf := make([]byte, bufBytes)
	for !s.isClosed() {
		if _, err := io.ReadFull(conn, fhdr[:]); err != nil {
			return
		}
		n := binary.LittleEndian.Uint32(fhdr[24:28])
		if int(n) > bufBytes {
			return
		}
		if n > 0 {
			if _, err := io.ReadFull(conn, buf[:n]); err != nil {
				return
			}
		}
	}
}
