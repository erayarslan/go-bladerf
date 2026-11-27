package mock

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
)

func TestMockServerHelloAndClose(t *testing.T) {
	srv, err := NewServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer func() {
		if err := srv.Close(); err != nil {
			t.Errorf("second Close returned error: %v", err)
		}
	}()

	conn, err := net.Dial("tcp", srv.Addr())
	if err != nil {
		t.Fatalf("net.Dial failed: %v", err)
	}
	defer conn.Close()

	var req [16]byte
	binary.LittleEndian.PutUint32(req[0:4], protoMagic)
	binary.LittleEndian.PutUint16(req[4:6], opHello)
	if _, err := conn.Write(req[:]); err != nil {
		t.Fatalf("write HELLO failed: %v", err)
	}

	var hdr [16]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		t.Fatalf("read HELLO header failed: %v", err)
	}
	if binary.LittleEndian.Uint32(hdr[0:4]) != protoMagic {
		t.Fatalf("unexpected magic: %#x", binary.LittleEndian.Uint32(hdr[0:4]))
	}
	payloadLen := binary.LittleEndian.Uint32(hdr[12:16])
	if payloadLen != 246 {
		t.Fatalf("unexpected HELLO payload len: %d, want 246", payloadLen)
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		t.Fatalf("read HELLO payload failed: %v", err)
	}
	if got := cString(payload[0:16]); got != "bladerf2" {
		t.Errorf("board_name = %q, want %q", got, "bladerf2")
	}
	if got := cString(payload[82:115]); got != "bladeRF 2.0" {
		t.Errorf("product = %q, want %q", got, "bladeRF 2.0")
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
