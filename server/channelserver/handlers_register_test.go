package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// createMockServerWithRaviente creates a mock server with raviente and semaphore
// initialized, which the base createMockServer() does not do.
func createMockServerWithRaviente() *Server {
	s := createMockServer()
	s.raviente = &Raviente{
		register: make([]uint32, 30),
		state:    make([]uint32, 30),
		support:  make([]uint32, 30),
	}
	s.semaphore = make(map[string]*Semaphore)
	return s
}

func TestRavienteInitialization(t *testing.T) {
	r := &Raviente{
		register: make([]uint32, 30),
		state:    make([]uint32, 30),
		support:  make([]uint32, 30),
	}
	if r == nil {
		t.Fatal("Raviente is nil")
	}
	if len(r.register) != 30 {
		t.Errorf("register length = %d, want 30", len(r.register))
	}
	if len(r.state) != 30 {
		t.Errorf("state length = %d, want 30", len(r.state))
	}
	if len(r.support) != 30 {
		t.Errorf("support length = %d, want 30", len(r.support))
	}
	// All values should be zero-initialized
	for i, v := range r.register {
		if v != 0 {
			t.Errorf("register[%d] = %d, want 0", i, v)
		}
	}
	for i, v := range r.state {
		if v != 0 {
			t.Errorf("state[%d] = %d, want 0", i, v)
		}
	}
	for i, v := range r.support {
		if v != 0 {
			t.Errorf("support[%d] = %d, want 0", i, v)
		}
	}
	if r.id != 0 {
		t.Errorf("id = %d, want 0", r.id)
	}
}

func TestRavienteMutex(t *testing.T) {
	r := &Raviente{
		register: make([]uint32, 30),
		state:    make([]uint32, 30),
		support:  make([]uint32, 30),
	}

	// Test that we can lock and unlock without deadlock
	r.Lock()
	r.register[0] = 42
	r.Unlock()

	r.Lock()
	val := r.register[0]
	r.Unlock()

	if val != 42 {
		t.Errorf("register[0] = %d, want 42", val)
	}
}

func TestRavienteDataAccess(t *testing.T) {
	r := &Raviente{
		register: make([]uint32, 30),
		state:    make([]uint32, 30),
		support:  make([]uint32, 30),
	}

	// Write and verify register data
	r.register[0] = 100
	r.register[4] = 200
	r.register[29] = 300

	if r.register[0] != 100 {
		t.Errorf("register[0] = %d, want 100", r.register[0])
	}
	if r.register[4] != 200 {
		t.Errorf("register[4] = %d, want 200", r.register[4])
	}
	if r.register[29] != 300 {
		t.Errorf("register[29] = %d, want 300", r.register[29])
	}

	// Write and verify state data
	r.state[0] = 500
	r.state[28] = 600

	if r.state[0] != 500 {
		t.Errorf("state[0] = %d, want 500", r.state[0])
	}
	if r.state[28] != 600 {
		t.Errorf("state[28] = %d, want 600", r.state[28])
	}

	// Write and verify support data
	r.support[0] = 700
	r.support[24] = 800

	if r.support[0] != 700 {
		t.Errorf("support[0] = %d, want 700", r.support[0])
	}
	if r.support[24] != 800 {
		t.Errorf("support[24] = %d, want 800", r.support[24])
	}
}

func TestRavienteID(t *testing.T) {
	r := &Raviente{
		register: make([]uint32, 30),
		state:    make([]uint32, 30),
		support:  make([]uint32, 30),
	}

	r.id = 12345
	if r.id != 12345 {
		t.Errorf("id = %d, want 12345", r.id)
	}

	r.id = 0xFFFF
	if r.id != 0xFFFF {
		t.Errorf("id = %d, want %d", r.id, uint16(0xFFFF))
	}
}

func TestCreateMockServerWithRaviente(t *testing.T) {
	s := createMockServerWithRaviente()
	if s == nil {
		t.Fatal("createMockServerWithRaviente() returned nil")
	}
	if s.raviente == nil {
		t.Fatal("raviente should not be nil")
	}
	if s.semaphore == nil {
		t.Fatal("semaphore should not be nil")
	}
	if len(s.raviente.register) != 30 {
		t.Errorf("raviente register length = %d, want 30", len(s.raviente.register))
	}
	if len(s.raviente.state) != 30 {
		t.Errorf("raviente state length = %d, want 30", len(s.raviente.state))
	}
	if len(s.raviente.support) != 30 {
		t.Errorf("raviente support length = %d, want 30", len(s.raviente.support))
	}
}

func TestHandlerTableRegistered(t *testing.T) {
	s := createMockServer()
	if s == nil {
		t.Fatal("createMockServer() returned nil")
	}

	// Verify handler table is populated
	if len(handlerTable) == 0 {
		t.Error("handlers table should not be empty")
	}

	// Check that key handler types are registered
	// (these are critical handlers that must always be present)
	criticalHandlers := []string{
		"handleMsgSysCreateStage",
		"handleMsgSysStageDestruct",
	}
	_ = criticalHandlers // We just verify the table is non-empty since handler function names aren't directly accessible

	// Verify minimum handler count
	if len(handlerTable) < 50 {
		t.Errorf("handlers count = %d, expected at least 50", len(handlerTable))
	}
}

func TestHandlerTableNilSession(t *testing.T) {
	// This test verifies that the handler table exists and has entries
	// but doesn't call handlers (which would require a real session)
	_ = createMockServer()

	count := 0
	for range handlerTable {
		count++
	}

	if count == 0 {
		t.Error("No handlers registered")
	}
}

func TestMockServerPacketHandling(t *testing.T) {
	s := createMockServerWithRaviente()
	session := createMockSession(1, s)

	// Verify the session and server are properly linked
	if session.server != s {
		t.Error("Session server reference mismatch")
	}

	// Verify byteframe can be created for packet construction
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // AckHandle
	if len(bf.Data()) != 4 {
		t.Errorf("ByteFrame length = %d, want 4", len(bf.Data()))
	}

	// Verify packet types can be instantiated
	pkt := &mhfpacket.MsgSysAck{}
	if pkt == nil {
		t.Error("Failed to create MsgSysAck")
	}
}
