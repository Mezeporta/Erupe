package conn

import (
	"io"
	"net"
	"testing"
)

// TestCryptConnRoundTrip verifies that encrypting and decrypting a packet
// through a pair of CryptConn instances produces the original data.
func TestCryptConnRoundTrip(t *testing.T) {
	// Create an in-process TCP pipe.
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	sender := NewCryptConn(client)
	receiver := NewCryptConn(server)

	testCases := [][]byte{
		{0x00, 0x14, 0x00, 0x00, 0x00, 0x01}, // Minimal login-like packet
		{0xDE, 0xAD, 0xBE, 0xEF},
		make([]byte, 256), // Larger packet
	}

	for i, original := range testCases {
		// Send in a goroutine to avoid blocking.
		errCh := make(chan error, 1)
		go func() {
			errCh <- sender.SendPacket(original)
		}()

		received, err := receiver.ReadPacket()
		if err != nil {
			t.Fatalf("case %d: ReadPacket error: %v", i, err)
		}

		if err := <-errCh; err != nil {
			t.Fatalf("case %d: SendPacket error: %v", i, err)
		}

		if len(received) != len(original) {
			t.Fatalf("case %d: length mismatch: got %d, want %d", i, len(received), len(original))
		}
		for j := range original {
			if received[j] != original[j] {
				t.Fatalf("case %d: byte %d mismatch: got 0x%02X, want 0x%02X", i, j, received[j], original[j])
			}
		}
	}
}

// TestCryptPacketHeaderRoundTrip verifies header encode/decode.
func TestCryptPacketHeaderRoundTrip(t *testing.T) {
	original := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0x03,
		PacketNum:               42,
		DataSize:                100,
		PrevPacketCombinedCheck: 0x1234,
		Check0:                 0xAAAA,
		Check1:                 0xBBBB,
		Check2:                 0xCCCC,
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if len(encoded) != CryptPacketHeaderLength {
		t.Fatalf("encoded length: got %d, want %d", len(encoded), CryptPacketHeaderLength)
	}

	decoded, err := NewCryptPacketHeader(encoded)
	if err != nil {
		t.Fatalf("NewCryptPacketHeader error: %v", err)
	}

	if *decoded != *original {
		t.Fatalf("header mismatch:\ngot  %+v\nwant %+v", *decoded, *original)
	}
}

// TestMultiPacketSequence verifies that key rotation stays in sync across
// multiple sequential packets.
func TestMultiPacketSequence(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	sender := NewCryptConn(client)
	receiver := NewCryptConn(server)

	for i := 0; i < 10; i++ {
		data := []byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}

		errCh := make(chan error, 1)
		go func() {
			errCh <- sender.SendPacket(data)
		}()

		received, err := receiver.ReadPacket()
		if err != nil {
			t.Fatalf("packet %d: ReadPacket error: %v", i, err)
		}

		if err := <-errCh; err != nil {
			t.Fatalf("packet %d: SendPacket error: %v", i, err)
		}

		for j := range data {
			if received[j] != data[j] {
				t.Fatalf("packet %d byte %d: got 0x%02X, want 0x%02X", i, j, received[j], data[j])
			}
		}
	}
}

// TestDialWithInit verifies that DialWithInit sends 8 NULL bytes on connect.
func TestDialWithInit(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	done := make(chan []byte, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 8)
		_, _ = io.ReadFull(conn, buf)
		done <- buf
	}()

	c, err := DialWithInit(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	initBytes := <-done
	for i, b := range initBytes {
		if b != 0 {
			t.Fatalf("init byte %d: got 0x%02X, want 0x00", i, b)
		}
	}
}
