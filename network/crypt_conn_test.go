package network

import (
	"bytes"
	"net"
	"testing"
)

func TestNewCryptConn(t *testing.T) {
	// NewCryptConn with nil should not panic
	cc := NewCryptConn(nil)

	if cc == nil {
		t.Fatal("NewCryptConn() returned nil")
	}

	// Verify default key rotation values
	if cc.readKeyRot != 995117 {
		t.Errorf("readKeyRot = %d, want 995117", cc.readKeyRot)
	}
	if cc.sendKeyRot != 995117 {
		t.Errorf("sendKeyRot = %d, want 995117", cc.sendKeyRot)
	}
	if cc.sentPackets != 0 {
		t.Errorf("sentPackets = %d, want 0", cc.sentPackets)
	}
	if cc.prevRecvPacketCombinedCheck != 0 {
		t.Errorf("prevRecvPacketCombinedCheck = %d, want 0", cc.prevRecvPacketCombinedCheck)
	}
	if cc.prevSendPacketCombinedCheck != 0 {
		t.Errorf("prevSendPacketCombinedCheck = %d, want 0", cc.prevSendPacketCombinedCheck)
	}
}

func TestCryptConnInitialState(t *testing.T) {
	cc := &CryptConn{}

	// Zero value should have all zeros
	if cc.readKeyRot != 0 {
		t.Errorf("zero value readKeyRot = %d, want 0", cc.readKeyRot)
	}
	if cc.sendKeyRot != 0 {
		t.Errorf("zero value sendKeyRot = %d, want 0", cc.sendKeyRot)
	}
	if cc.conn != nil {
		t.Error("zero value conn should be nil")
	}
}

func TestCryptConnDefaultKeyRotation(t *testing.T) {
	// The magic number 995117 is the default key rotation value
	const defaultKeyRot = 995117

	cc := NewCryptConn(nil)

	if cc.readKeyRot != defaultKeyRot {
		t.Errorf("default readKeyRot = %d, want %d", cc.readKeyRot, defaultKeyRot)
	}
	if cc.sendKeyRot != defaultKeyRot {
		t.Errorf("default sendKeyRot = %d, want %d", cc.sendKeyRot, defaultKeyRot)
	}
}

func TestCryptConnStructFields(t *testing.T) {
	cc := &CryptConn{
		readKeyRot:                  123456,
		sendKeyRot:                  654321,
		sentPackets:                 10,
		prevRecvPacketCombinedCheck: 0x1234,
		prevSendPacketCombinedCheck: 0x5678,
	}

	if cc.readKeyRot != 123456 {
		t.Errorf("readKeyRot = %d, want 123456", cc.readKeyRot)
	}
	if cc.sendKeyRot != 654321 {
		t.Errorf("sendKeyRot = %d, want 654321", cc.sendKeyRot)
	}
	if cc.sentPackets != 10 {
		t.Errorf("sentPackets = %d, want 10", cc.sentPackets)
	}
	if cc.prevRecvPacketCombinedCheck != 0x1234 {
		t.Errorf("prevRecvPacketCombinedCheck = 0x%X, want 0x1234", cc.prevRecvPacketCombinedCheck)
	}
	if cc.prevSendPacketCombinedCheck != 0x5678 {
		t.Errorf("prevSendPacketCombinedCheck = 0x%X, want 0x5678", cc.prevSendPacketCombinedCheck)
	}
}

func TestCryptConnKeyRotationType(t *testing.T) {
	// Verify key rotation uses uint32
	cc := NewCryptConn(nil)

	// Simulate key rotation
	keyRotDelta := byte(3)
	cc.sendKeyRot = (uint32(keyRotDelta) * (cc.sendKeyRot + 1))

	// Should not overflow or behave unexpectedly
	if cc.sendKeyRot == 0 {
		t.Error("sendKeyRot should not be 0 after rotation")
	}
}

func TestCryptConnSentPacketsCounter(t *testing.T) {
	cc := NewCryptConn(nil)

	if cc.sentPackets != 0 {
		t.Errorf("initial sentPackets = %d, want 0", cc.sentPackets)
	}

	// Simulate incrementing sent packets
	cc.sentPackets++
	if cc.sentPackets != 1 {
		t.Errorf("sentPackets after increment = %d, want 1", cc.sentPackets)
	}

	// Verify it's int32
	cc.sentPackets = 0x7FFFFFFF // Max int32
	if cc.sentPackets != 0x7FFFFFFF {
		t.Errorf("sentPackets max value = %d, want %d", cc.sentPackets, 0x7FFFFFFF)
	}
}

func TestCryptConnCombinedCheckStorage(t *testing.T) {
	cc := NewCryptConn(nil)

	// Test combined check storage
	cc.prevRecvPacketCombinedCheck = 0xABCD
	cc.prevSendPacketCombinedCheck = 0xDCBA

	if cc.prevRecvPacketCombinedCheck != 0xABCD {
		t.Errorf("prevRecvPacketCombinedCheck = 0x%X, want 0xABCD", cc.prevRecvPacketCombinedCheck)
	}
	if cc.prevSendPacketCombinedCheck != 0xDCBA {
		t.Errorf("prevSendPacketCombinedCheck = 0x%X, want 0xDCBA", cc.prevSendPacketCombinedCheck)
	}
}

func TestCryptConnKeyRotationFormula(t *testing.T) {
	// Test the key rotation formula: (keyRotDelta * (keyRot + 1))
	tests := []struct {
		name        string
		initialKey  uint32
		keyRotDelta byte
		expectedKey uint32
	}{
		{"delta 1", 995117, 1, 995118},
		{"delta 3 default", 995117, 3, 2985354},
		{"delta 0", 995117, 0, 0},
		{"zero initial", 0, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newKey := uint32(tt.keyRotDelta) * (tt.initialKey + 1)
			if newKey != tt.expectedKey {
				t.Errorf("key rotation = %d, want %d", newKey, tt.expectedKey)
			}
		})
	}
}

func TestCryptPacketHeaderLengthConstant(t *testing.T) {
	// CryptPacketHeaderLength should always be 14
	if CryptPacketHeaderLength != 14 {
		t.Errorf("CryptPacketHeaderLength = %d, want 14", CryptPacketHeaderLength)
	}
}

func TestMultipleCryptConnInstances(t *testing.T) {
	// Multiple instances should be independent
	cc1 := NewCryptConn(nil)
	cc2 := NewCryptConn(nil)

	cc1.sendKeyRot = 12345
	cc2.sendKeyRot = 54321

	if cc1.sendKeyRot == cc2.sendKeyRot {
		t.Error("CryptConn instances should be independent")
	}
}

func TestCryptConnSendAndReadPacket(t *testing.T) {
	// Use net.Pipe to create an in-memory bidirectional connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sender := NewCryptConn(clientConn)
	receiver := NewCryptConn(serverConn)

	testData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	// Send in a goroutine since Pipe is synchronous
	errCh := make(chan error, 1)
	go func() {
		errCh <- sender.SendPacket(testData)
	}()

	// Read on the other end
	received, err := receiver.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket() error = %v", err)
	}

	if sendErr := <-errCh; sendErr != nil {
		t.Fatalf("SendPacket() error = %v", sendErr)
	}

	if !bytes.Equal(received, testData) {
		t.Errorf("ReadPacket() = %v, want %v", received, testData)
	}
}

func TestCryptConnMultiplePackets(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sender := NewCryptConn(clientConn)
	receiver := NewCryptConn(serverConn)

	packets := [][]byte{
		{0x01, 0x02, 0x03, 0x04},
		{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE},
		{0xFF},
		make([]byte, 64),
	}

	errCh := make(chan error, 1)
	go func() {
		for _, pkt := range packets {
			if err := sender.SendPacket(pkt); err != nil {
				errCh <- err
				return
			}
		}
		errCh <- nil
	}()

	for i, expected := range packets {
		received, err := receiver.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket() packet %d error = %v", i, err)
		}
		if !bytes.Equal(received, expected) {
			t.Errorf("Packet %d: got %v, want %v", i, received, expected)
		}
	}

	if sendErr := <-errCh; sendErr != nil {
		t.Fatalf("SendPacket() error = %v", sendErr)
	}
}

func TestCryptConnSendPacketStateUpdate(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sender := NewCryptConn(clientConn)

	// Consume the data on the other side
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := serverConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	if sender.sentPackets != 0 {
		t.Errorf("initial sentPackets = %d, want 0", sender.sentPackets)
	}

	err := sender.SendPacket([]byte{0x01, 0x02, 0x03, 0x04})
	if err != nil {
		t.Fatalf("SendPacket() error = %v", err)
	}

	if sender.sentPackets != 1 {
		t.Errorf("sentPackets after 1 send = %d, want 1", sender.sentPackets)
	}

	// Key rotation should have changed from default
	if sender.sendKeyRot == 995117 {
		t.Error("sendKeyRot should have changed after SendPacket")
	}

	if sender.prevSendPacketCombinedCheck == 0 {
		t.Error("prevSendPacketCombinedCheck should be set after SendPacket")
	}
}

func TestCryptConnReadPacketClosedConn(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	receiver := NewCryptConn(serverConn)

	// Close the writing end
	clientConn.Close()

	_, err := receiver.ReadPacket()
	if err == nil {
		t.Error("ReadPacket() on closed connection should return error")
	}
	serverConn.Close()
}
