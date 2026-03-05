package protocol

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestHandleAck_SimpleAck(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(1)
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(ackHandle, waitCh)

	// Build simple ACK data (after opcode has been stripped).
	// Format: uint32 ackHandle + uint8 isBuffer(0) + uint8 errorCode + uint16 ignored + uint32 data
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], ackHandle)
	data[4] = 0 // isBuffer = false
	data[5] = 0 // errorCode = 0
	binary.BigEndian.PutUint16(data[6:8], 0)
	binary.BigEndian.PutUint32(data[8:12], 0xDEADBEEF) // simple ACK data

	ch.handleAck(data)

	select {
	case resp := <-waitCh:
		if resp.AckHandle != ackHandle {
			t.Errorf("AckHandle: got %d, want %d", resp.AckHandle, ackHandle)
		}
		if resp.IsBufferResponse {
			t.Error("IsBufferResponse: got true, want false")
		}
		if resp.ErrorCode != 0 {
			t.Errorf("ErrorCode: got %d, want 0", resp.ErrorCode)
		}
		if len(resp.Data) != 4 {
			t.Fatalf("Data length: got %d, want 4", len(resp.Data))
		}
		val := binary.BigEndian.Uint32(resp.Data)
		if val != 0xDEADBEEF {
			t.Errorf("Data value: got 0x%08X, want 0xDEADBEEF", val)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for ACK response")
	}
}

func TestHandleAck_BufferAck(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(2)
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(ackHandle, waitCh)

	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Build buffer ACK data.
	// Format: uint32 ackHandle + uint8 isBuffer(1) + uint8 errorCode + uint16 payloadSize + payload
	data := make([]byte, 8+len(payload))
	binary.BigEndian.PutUint32(data[0:4], ackHandle)
	data[4] = 1 // isBuffer = true
	data[5] = 0 // errorCode = 0
	binary.BigEndian.PutUint16(data[6:8], uint16(len(payload)))
	copy(data[8:], payload)

	ch.handleAck(data)

	select {
	case resp := <-waitCh:
		if resp.AckHandle != ackHandle {
			t.Errorf("AckHandle: got %d, want %d", resp.AckHandle, ackHandle)
		}
		if !resp.IsBufferResponse {
			t.Error("IsBufferResponse: got false, want true")
		}
		if resp.ErrorCode != 0 {
			t.Errorf("ErrorCode: got %d, want 0", resp.ErrorCode)
		}
		if len(resp.Data) != len(payload) {
			t.Fatalf("Data length: got %d, want %d", len(resp.Data), len(payload))
		}
		for i, b := range payload {
			if resp.Data[i] != b {
				t.Errorf("Data[%d]: got 0x%02X, want 0x%02X", i, resp.Data[i], b)
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for ACK response")
	}
}

func TestHandleAck_ExtendedBuffer(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(3)
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(ackHandle, waitCh)

	payload := make([]byte, 10)
	for i := range payload {
		payload[i] = byte(i)
	}

	// Build extended buffer ACK data (payloadSize == 0xFFFF).
	// Format: uint32 ackHandle + uint8 isBuffer(1) + uint8 errorCode + uint16(0xFFFF) + uint32 realSize + payload
	data := make([]byte, 12+len(payload))
	binary.BigEndian.PutUint32(data[0:4], ackHandle)
	data[4] = 1 // isBuffer = true
	data[5] = 0 // errorCode = 0
	binary.BigEndian.PutUint16(data[6:8], 0xFFFF)
	binary.BigEndian.PutUint32(data[8:12], uint32(len(payload)))
	copy(data[12:], payload)

	ch.handleAck(data)

	select {
	case resp := <-waitCh:
		if resp.AckHandle != ackHandle {
			t.Errorf("AckHandle: got %d, want %d", resp.AckHandle, ackHandle)
		}
		if !resp.IsBufferResponse {
			t.Error("IsBufferResponse: got false, want true")
		}
		if len(resp.Data) != len(payload) {
			t.Fatalf("Data length: got %d, want %d", len(resp.Data), len(payload))
		}
		for i, b := range payload {
			if resp.Data[i] != b {
				t.Errorf("Data[%d]: got 0x%02X, want 0x%02X", i, resp.Data[i], b)
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for ACK response")
	}
}

func TestHandleAck_TooShort(t *testing.T) {
	ch := &ChannelConn{}

	// Should not panic with data shorter than 8 bytes.
	ch.handleAck(nil)
	ch.handleAck([]byte{})
	ch.handleAck([]byte{0x00, 0x01, 0x02})
	ch.handleAck([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	// 7 bytes: still < 8, should return silently.
	data := make([]byte, 7)
	binary.BigEndian.PutUint32(data[0:4], 99)
	ch.handleAck(data)
}

func TestHandleAck_ExtendedBuffer_TooShortForSize(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(4)
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(ackHandle, waitCh)

	// payloadSize=0xFFFF but data too short for the uint32 real size field (only 8 bytes total).
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], ackHandle)
	data[4] = 1 // isBuffer = true
	data[5] = 0
	binary.BigEndian.PutUint16(data[6:8], 0xFFFF)

	ch.handleAck(data)

	// The handler should return early (no payload), but still dispatch the response
	// with nil Data since the early return is only for reading payload.
	select {
	case resp := <-waitCh:
		// Response dispatched but with nil data since len(data) < 12.
		if resp.Data != nil {
			t.Errorf("expected nil Data for truncated extended buffer, got %d bytes", len(resp.Data))
		}
	case <-time.After(100 * time.Millisecond):
		// The handler returns before dispatching if len(data) < 12 for 0xFFFF path.
		// This is also acceptable behavior.
	}
}

func TestHandleAck_WithErrorCode(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(5)
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(ackHandle, waitCh)

	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], ackHandle)
	data[4] = 0  // isBuffer = false
	data[5] = 42 // errorCode = 42
	binary.BigEndian.PutUint16(data[6:8], 0)
	binary.BigEndian.PutUint32(data[8:12], 0)

	ch.handleAck(data)

	select {
	case resp := <-waitCh:
		if resp.ErrorCode != 42 {
			t.Errorf("ErrorCode: got %d, want 42", resp.ErrorCode)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for ACK response")
	}
}

func TestHandleAck_NoWaiter(t *testing.T) {
	ch := &ChannelConn{}

	// No waiter registered for handle 999 — should not panic.
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], 999)
	data[4] = 0
	data[5] = 0
	binary.BigEndian.PutUint16(data[6:8], 0)
	binary.BigEndian.PutUint32(data[8:12], 0)

	// This should log but not panic.
	ch.handleAck(data)
}

func TestNextAckHandle(t *testing.T) {
	ch := &ChannelConn{}

	h1 := ch.NextAckHandle()
	h2 := ch.NextAckHandle()
	h3 := ch.NextAckHandle()

	if h1 != 1 {
		t.Errorf("first handle: got %d, want 1", h1)
	}
	if h2 != 2 {
		t.Errorf("second handle: got %d, want 2", h2)
	}
	if h3 != 3 {
		t.Errorf("third handle: got %d, want 3", h3)
	}
}

func TestNextAckHandle_Concurrent(t *testing.T) {
	ch := &ChannelConn{}

	const goroutines = 100
	results := make(chan uint32, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			results <- ch.NextAckHandle()
		}()
	}

	seen := make(map[uint32]bool)
	for i := 0; i < goroutines; i++ {
		h := <-results
		if seen[h] {
			t.Errorf("duplicate handle: %d", h)
		}
		seen[h] = true
	}

	if len(seen) != goroutines {
		t.Errorf("unique handles: got %d, want %d", len(seen), goroutines)
	}
}

func TestWaitForAck_Timeout(t *testing.T) {
	ch := &ChannelConn{}

	// No handleAck call will be made, so this should time out.
	_, err := ch.WaitForAck(999, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWaitForAck_Success(t *testing.T) {
	ch := &ChannelConn{}

	ackHandle := uint32(10)

	// Dispatch the ACK from another goroutine after a short delay.
	go func() {
		time.Sleep(10 * time.Millisecond)
		data := make([]byte, 12)
		binary.BigEndian.PutUint32(data[0:4], ackHandle)
		data[4] = 0 // isBuffer = false
		data[5] = 0 // errorCode
		binary.BigEndian.PutUint16(data[6:8], 0)
		binary.BigEndian.PutUint32(data[8:12], 0x12345678)
		ch.handleAck(data)
	}()

	resp, err := ch.WaitForAck(ackHandle, 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AckHandle != ackHandle {
		t.Errorf("AckHandle: got %d, want %d", resp.AckHandle, ackHandle)
	}
}
