package channelserver

import (
	"sync"
	"testing"
)

func TestObjectStruct(t *testing.T) {
	obj := &Object{
		id:          12345,
		ownerCharID: 67890,
		x:           100.5,
		y:           50.25,
		z:           -10.0,
	}

	if obj.id != 12345 {
		t.Errorf("Object id = %d, want 12345", obj.id)
	}
	if obj.ownerCharID != 67890 {
		t.Errorf("Object ownerCharID = %d, want 67890", obj.ownerCharID)
	}
	if obj.x != 100.5 {
		t.Errorf("Object x = %f, want 100.5", obj.x)
	}
	if obj.y != 50.25 {
		t.Errorf("Object y = %f, want 50.25", obj.y)
	}
	if obj.z != -10.0 {
		t.Errorf("Object z = %f, want -10.0", obj.z)
	}
}

func TestObjectRWMutex(t *testing.T) {
	obj := &Object{
		id:          1,
		ownerCharID: 100,
		x:           0,
		y:           0,
		z:           0,
	}

	// Test read lock
	obj.RLock()
	_ = obj.x
	obj.RUnlock()

	// Test write lock
	obj.Lock()
	obj.x = 100.0
	obj.Unlock()

	if obj.x != 100.0 {
		t.Errorf("Object x = %f, want 100.0 after write", obj.x)
	}
}

func TestObjectConcurrentAccess(t *testing.T) {
	obj := &Object{
		id:          1,
		ownerCharID: 100,
		x:           0,
		y:           0,
		z:           0,
	}

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val float32) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				obj.Lock()
				obj.x = val
				obj.y = val
				obj.z = val
				obj.Unlock()
			}
		}(float32(i))
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				obj.RLock()
				_ = obj.x
				_ = obj.y
				_ = obj.z
				obj.RUnlock()
			}
		}()
	}

	wg.Wait()
}

func TestStageBinaryKeyStruct(t *testing.T) {
	key1 := stageBinaryKey{id0: 1, id1: 2}
	key2 := stageBinaryKey{id0: 1, id1: 3}
	key3 := stageBinaryKey{id0: 1, id1: 2}

	// Different keys
	if key1 == key2 {
		t.Error("key1 and key2 should be different")
	}

	// Same keys
	if key1 != key3 {
		t.Error("key1 and key3 should be equal")
	}
}

func TestStageBinaryKeyAsMapKey(t *testing.T) {
	data := make(map[stageBinaryKey][]byte)

	key1 := stageBinaryKey{id0: 0, id1: 0}
	key2 := stageBinaryKey{id0: 0, id1: 1}
	key3 := stageBinaryKey{id0: 1, id1: 0}

	data[key1] = []byte{0x01}
	data[key2] = []byte{0x02}
	data[key3] = []byte{0x03}

	if len(data) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(data))
	}

	if data[key1][0] != 0x01 {
		t.Errorf("data[key1] = 0x%02X, want 0x01", data[key1][0])
	}
	if data[key2][0] != 0x02 {
		t.Errorf("data[key2] = 0x%02X, want 0x02", data[key2][0])
	}
	if data[key3][0] != 0x03 {
		t.Errorf("data[key3] = 0x%02X, want 0x03", data[key3][0])
	}
}

func TestNewStageDefaults(t *testing.T) {
	stage := NewStage("test_stage_001")

	if stage.id != "test_stage_001" {
		t.Errorf("stage.id = %s, want test_stage_001", stage.id)
	}
	if stage.maxPlayers != 4 {
		t.Errorf("stage.maxPlayers = %d, want 4 (default)", stage.maxPlayers)
	}
	if stage.objectIndex != 0 {
		t.Errorf("stage.objectIndex = %d, want 0", stage.objectIndex)
	}
	if stage.clients == nil {
		t.Error("stage.clients should be initialized")
	}
	if stage.reservedClientSlots == nil {
		t.Error("stage.reservedClientSlots should be initialized")
	}
	if stage.objects == nil {
		t.Error("stage.objects should be initialized")
	}
	if stage.rawBinaryData == nil {
		t.Error("stage.rawBinaryData should be initialized")
	}
	if stage.host != nil {
		t.Error("stage.host should be nil initially")
	}
	if stage.password != "" {
		t.Errorf("stage.password should be empty, got %s", stage.password)
	}
}

func TestStageNextObjectID(t *testing.T) {
	stage := NewStage("test")

	// First ID should be 1 (index 0 is skipped)
	id1 := stage.NextObjectID()
	if stage.objectIndex != 1 {
		t.Errorf("objectIndex after first call = %d, want 1", stage.objectIndex)
	}

	// Get several IDs and ensure they increment
	id2 := stage.NextObjectID()
	id3 := stage.NextObjectID()

	if id1 == id2 || id2 == id3 {
		t.Error("Object IDs should be unique")
	}
}

func TestStageNextObjectID_WrapAround(t *testing.T) {
	stage := NewStage("test")
	stage.objectIndex = 125

	// Get ID at 126
	stage.NextObjectID()
	if stage.objectIndex != 126 {
		t.Errorf("objectIndex = %d, want 126", stage.objectIndex)
	}

	// Next should wrap to 1 (skipping 0 and 127)
	stage.NextObjectID()
	if stage.objectIndex != 1 {
		t.Errorf("objectIndex = %d, want 1 (wrapped around)", stage.objectIndex)
	}
}

func TestStageIsQuest(t *testing.T) {
	stage := NewStage("test")

	// Initially not a quest
	if stage.isQuest() {
		t.Error("New stage should not be a quest")
	}

	// Add reserved slot
	stage.reservedClientSlots[100] = true

	if !stage.isQuest() {
		t.Error("Stage with reserved slots should be a quest")
	}
}

func TestStageReservedClientSlots(t *testing.T) {
	stage := NewStage("test")

	// Reserve some slots
	stage.reservedClientSlots[100] = true
	stage.reservedClientSlots[200] = false // ready status doesn't matter for presence
	stage.reservedClientSlots[300] = true

	if len(stage.reservedClientSlots) != 3 {
		t.Errorf("reservedClientSlots count = %d, want 3", len(stage.reservedClientSlots))
	}

	// Check ready status
	if !stage.reservedClientSlots[100] {
		t.Error("charID 100 should be ready")
	}
	if stage.reservedClientSlots[200] {
		t.Error("charID 200 should not be ready")
	}
}

func TestStageRawBinaryData(t *testing.T) {
	stage := NewStage("test")

	key := stageBinaryKey{id0: 5, id1: 10}
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	stage.rawBinaryData[key] = data

	retrieved := stage.rawBinaryData[key]
	if len(retrieved) != 4 {
		t.Fatalf("retrieved data len = %d, want 4", len(retrieved))
	}
	if retrieved[0] != 0xDE || retrieved[3] != 0xEF {
		t.Error("retrieved data doesn't match stored data")
	}
}

func TestStageObjects(t *testing.T) {
	stage := NewStage("test")

	obj := &Object{
		id:          stage.NextObjectID(),
		ownerCharID: 12345,
		x:           100.0,
		y:           200.0,
		z:           300.0,
	}

	stage.objects[obj.id] = obj

	if len(stage.objects) != 1 {
		t.Errorf("objects count = %d, want 1", len(stage.objects))
	}

	retrieved := stage.objects[obj.id]
	if retrieved.ownerCharID != 12345 {
		t.Errorf("retrieved object ownerCharID = %d, want 12345", retrieved.ownerCharID)
	}
}

func TestStageHost(t *testing.T) {
	server := createMockServer()
	stage := NewStage("test")

	// Set host
	host := createMockSession(100, server)
	stage.host = host

	if stage.host != host {
		t.Error("stage host not set correctly")
	}
	if stage.host.charID != 100 {
		t.Errorf("stage host charID = %d, want 100", stage.host.charID)
	}
}

func TestStagePassword(t *testing.T) {
	stage := NewStage("test")

	// Set password
	stage.password = "secret123"

	if stage.password != "secret123" {
		t.Errorf("stage password = %s, want secret123", stage.password)
	}
}

func TestStageMaxPlayers(t *testing.T) {
	stage := NewStage("test")

	// Change max players
	stage.maxPlayers = 16

	if stage.maxPlayers != 16 {
		t.Errorf("stage maxPlayers = %d, want 16", stage.maxPlayers)
	}
}

func TestStageConcurrentClientAccess(t *testing.T) {
	server := createMockServer()
	stage := NewStage("test")

	var wg sync.WaitGroup

	// Concurrent client additions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				session := createMockSession(uint32(id*100+j), server)
				stage.Lock()
				stage.clients[session] = session.charID
				stage.Unlock()

				stage.Lock()
				delete(stage.clients, session)
				stage.Unlock()
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				stage.RLock()
				_ = len(stage.clients)
				stage.RUnlock()
			}
		}()
	}

	wg.Wait()
}

func TestStageBroadcastMHF_EmptyStage(t *testing.T) {
	stage := NewStage("test")
	pkt := &mockPacket{opcode: 0x1234}

	// Should not panic with empty stage
	stage.BroadcastMHF(pkt, nil)
}

func TestStageBroadcastMHF_SkipsNilClientContext(t *testing.T) {
	server := createMockServer()
	stage := NewStage("test")

	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)
	session2.clientContext = nil // Nil context should be skipped

	stage.clients[session1] = session1.charID
	stage.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x1234}

	// Should not panic
	stage.BroadcastMHF(pkt, nil)

	// Only session1 should receive
	select {
	case <-session1.sendPackets:
		// Good
	default:
		t.Error("session1 should receive broadcast")
	}
}
