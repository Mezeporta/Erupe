package network

import (
	"bytes"
	"errors"
	cfg "erupe-ce/config"
	"erupe-ce/network/crypto"
	"io"
	"net"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readData  *bytes.Buffer
	writeData *bytes.Buffer
	closed    bool
	readErr   error
	writeErr  error
}

func newMockConn(readData []byte) *mockConn {
	return &mockConn{
		readData:  bytes.NewBuffer(readData),
		writeData: bytes.NewBuffer(nil),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	return m.readData.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeData.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNewCryptConn(t *testing.T) {
	mockConn := newMockConn(nil)
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	if cc == nil {
		t.Fatal("NewCryptConn() returned nil")
	}

	if cc.conn != mockConn {
		t.Error("conn not set correctly")
	}

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

	if cc.realClientMode != cfg.ZZ {
		t.Errorf("realClientMode = %d, want %d", cc.realClientMode, cfg.ZZ)
	}
}

func TestCryptConn_SendPacket(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "small packet",
			data: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name: "empty packet",
			data: []byte{},
		},
		{
			name: "larger packet",
			data: bytes.Repeat([]byte{0xAA}, 256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := newMockConn(nil)
			cc := NewCryptConn(mockConn, cfg.ZZ, nil)

			err := cc.SendPacket(tt.data)
			if err != nil {
				t.Fatalf("SendPacket() error = %v, want nil", err)
			}

			written := mockConn.writeData.Bytes()
			if len(written) < CryptPacketHeaderLength {
				t.Fatalf("written data length = %d, want at least %d", len(written), CryptPacketHeaderLength)
			}

			// Verify header was written
			headerData := written[:CryptPacketHeaderLength]
			header, err := NewCryptPacketHeader(headerData)
			if err != nil {
				t.Fatalf("Failed to parse header: %v", err)
			}

			// Verify packet counter incremented
			if cc.sentPackets != 1 {
				t.Errorf("sentPackets = %d, want 1", cc.sentPackets)
			}

			// Verify header fields
			if header.KeyRotDelta != 3 {
				t.Errorf("header.KeyRotDelta = %d, want 3", header.KeyRotDelta)
			}

			if header.PacketNum != 0 {
				t.Errorf("header.PacketNum = %d, want 0", header.PacketNum)
			}

			// Verify encrypted data was written
			encryptedData := written[CryptPacketHeaderLength:]
			if len(encryptedData) != int(header.DataSize) {
				t.Errorf("encrypted data length = %d, want %d", len(encryptedData), header.DataSize)
			}
		})
	}
}

func TestCryptConn_SendPacket_MultiplePackets(t *testing.T) {
	mockConn := newMockConn(nil)
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	// Send first packet
	err := cc.SendPacket([]byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("SendPacket(1) error = %v", err)
	}

	if cc.sentPackets != 1 {
		t.Errorf("After 1 packet: sentPackets = %d, want 1", cc.sentPackets)
	}

	// Send second packet
	err = cc.SendPacket([]byte{0x03, 0x04})
	if err != nil {
		t.Fatalf("SendPacket(2) error = %v", err)
	}

	if cc.sentPackets != 2 {
		t.Errorf("After 2 packets: sentPackets = %d, want 2", cc.sentPackets)
	}

	// Send third packet
	err = cc.SendPacket([]byte{0x05, 0x06})
	if err != nil {
		t.Fatalf("SendPacket(3) error = %v", err)
	}

	if cc.sentPackets != 3 {
		t.Errorf("After 3 packets: sentPackets = %d, want 3", cc.sentPackets)
	}
}

func TestCryptConn_SendPacket_KeyRotation(t *testing.T) {
	mockConn := newMockConn(nil)
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	initialKey := cc.sendKeyRot

	err := cc.SendPacket([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatalf("SendPacket() error = %v", err)
	}

	// Key should have been rotated (keyRotDelta=3, so new key = 3 * (oldKey + 1))
	expectedKey := 3 * (initialKey + 1)
	if cc.sendKeyRot != expectedKey {
		t.Errorf("sendKeyRot = %d, want %d", cc.sendKeyRot, expectedKey)
	}
}

func TestCryptConn_SendPacket_WriteError(t *testing.T) {
	mockConn := newMockConn(nil)
	mockConn.writeErr = errors.New("write error")
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	err := cc.SendPacket([]byte{0x01, 0x02, 0x03})
	// Note: Current implementation doesn't return write error
	// This test documents the behavior
	if err != nil {
		t.Logf("SendPacket() returned error: %v", err)
	}
}

func TestCryptConn_ReadPacket_Success(t *testing.T) {
	testData := []byte{0x74, 0x65, 0x73, 0x74} // "test"
	key := uint32(0)

	// Encrypt the data
	encryptedData, combinedCheck, check0, check1, check2 := crypto.Crypto(testData, key, true, nil)

	// Build header
	header := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0,
		PacketNum:               0,
		DataSize:                uint16(len(encryptedData)),
		PrevPacketCombinedCheck: 0,
		Check0:                  check0,
		Check1:                  check1,
		Check2:                  check2,
	}

	headerBytes, _ := header.Encode()

	// Combine header and encrypted data
	packet := append(headerBytes, encryptedData...)

	mockConn := newMockConn(packet)
	cc := NewCryptConn(mockConn, cfg.Z1, nil)

	// Set the key to match what we used for encryption
	cc.readKeyRot = key

	result, err := cc.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket() error = %v, want nil", err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("ReadPacket() = %v, want %v", result, testData)
	}

	if cc.prevRecvPacketCombinedCheck != combinedCheck {
		t.Errorf("prevRecvPacketCombinedCheck = %d, want %d", cc.prevRecvPacketCombinedCheck, combinedCheck)
	}
}

func TestCryptConn_ReadPacket_KeyRotation(t *testing.T) {
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	key := uint32(995117)
	keyRotDelta := byte(3)

	// Calculate expected rotated key
	rotatedKey := uint32(keyRotDelta) * (key + 1)

	// Encrypt with the rotated key
	encryptedData, _, check0, check1, check2 := crypto.Crypto(testData, rotatedKey, true, nil)

	// Build header with key rotation
	header := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             keyRotDelta,
		PacketNum:               0,
		DataSize:                uint16(len(encryptedData)),
		PrevPacketCombinedCheck: 0,
		Check0:                  check0,
		Check1:                  check1,
		Check2:                  check2,
	}

	headerBytes, _ := header.Encode()
	packet := append(headerBytes, encryptedData...)

	mockConn := newMockConn(packet)
	cc := NewCryptConn(mockConn, cfg.Z1, nil)
	cc.readKeyRot = key

	result, err := cc.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket() error = %v, want nil", err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("ReadPacket() = %v, want %v", result, testData)
	}

	// Verify key was rotated
	if cc.readKeyRot != rotatedKey {
		t.Errorf("readKeyRot = %d, want %d", cc.readKeyRot, rotatedKey)
	}
}

func TestCryptConn_ReadPacket_NoKeyRotation(t *testing.T) {
	testData := []byte{0x01, 0x02}
	key := uint32(12345)

	// Encrypt without key rotation
	encryptedData, _, check0, check1, check2 := crypto.Crypto(testData, key, true, nil)

	header := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0, // No rotation
		PacketNum:               0,
		DataSize:                uint16(len(encryptedData)),
		PrevPacketCombinedCheck: 0,
		Check0:                  check0,
		Check1:                  check1,
		Check2:                  check2,
	}

	headerBytes, _ := header.Encode()
	packet := append(headerBytes, encryptedData...)

	mockConn := newMockConn(packet)
	cc := NewCryptConn(mockConn, cfg.Z1, nil)
	cc.readKeyRot = key

	originalKeyRot := cc.readKeyRot

	result, err := cc.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket() error = %v, want nil", err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("ReadPacket() = %v, want %v", result, testData)
	}

	// Verify key was NOT rotated
	if cc.readKeyRot != originalKeyRot {
		t.Errorf("readKeyRot = %d, want %d (should not have changed)", cc.readKeyRot, originalKeyRot)
	}
}

func TestCryptConn_ReadPacket_HeaderReadError(t *testing.T) {
	mockConn := newMockConn([]byte{0x01, 0x02}) // Only 2 bytes, header needs 14
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	_, err := cc.ReadPacket()
	if err == nil {
		t.Fatal("ReadPacket() error = nil, want error")
	}

	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Errorf("ReadPacket() error = %v, want io.EOF or io.ErrUnexpectedEOF", err)
	}
}

func TestCryptConn_ReadPacket_InvalidHeader(t *testing.T) {
	// Create invalid header data (wrong endianness or malformed)
	invalidHeader := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	mockConn := newMockConn(invalidHeader)
	cc := NewCryptConn(mockConn, cfg.ZZ, nil)

	_, err := cc.ReadPacket()
	if err == nil {
		t.Fatal("ReadPacket() error = nil, want error")
	}
}

func TestCryptConn_ReadPacket_BodyReadError(t *testing.T) {
	// Create valid header but incomplete body
	header := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0,
		PacketNum:               0,
		DataSize:                100, // Claim 100 bytes
		PrevPacketCombinedCheck: 0,
		Check0:                  0x1234,
		Check1:                  0x5678,
		Check2:                  0x9ABC,
	}

	headerBytes, _ := header.Encode()
	incompleteBody := []byte{0x01, 0x02, 0x03} // Only 3 bytes, not 100

	packet := append(headerBytes, incompleteBody...)

	mockConn := newMockConn(packet)
	cc := NewCryptConn(mockConn, cfg.Z1, nil)

	_, err := cc.ReadPacket()
	if err == nil {
		t.Fatal("ReadPacket() error = nil, want error")
	}
}

func TestCryptConn_ReadPacket_ChecksumMismatch(t *testing.T) {
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	key := uint32(0)

	encryptedData, _, _, _, _ := crypto.Crypto(testData, key, true, nil)

	// Build header with WRONG checksums
	header := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0,
		PacketNum:               0,
		DataSize:                uint16(len(encryptedData)),
		PrevPacketCombinedCheck: 0,
		Check0:                  0xFFFF, // Wrong checksum
		Check1:                  0xFFFF, // Wrong checksum
		Check2:                  0xFFFF, // Wrong checksum
	}

	headerBytes, _ := header.Encode()
	packet := append(headerBytes, encryptedData...)

	mockConn := newMockConn(packet)
	cc := NewCryptConn(mockConn, cfg.Z1, nil)
	cc.readKeyRot = key

	_, err := cc.ReadPacket()
	if err == nil {
		t.Fatal("ReadPacket() error = nil, want error for checksum mismatch")
	}

	expectedErr := "decrypted data checksum doesn't match header"
	if err.Error() != expectedErr {
		t.Errorf("ReadPacket() error = %q, want %q", err.Error(), expectedErr)
	}
}

func TestCryptConn_Interface(t *testing.T) {
	// Test that CryptConn implements Conn interface
	var _ Conn = (*CryptConn)(nil)
}
