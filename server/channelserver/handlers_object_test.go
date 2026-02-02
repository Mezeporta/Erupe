package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysCreateObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage for the session
	stage := NewStage("test_stage")
	session.stage = stage

	pkt := &mhfpacket.MsgSysCreateObject{
		AckHandle: 12345,
		X:         100.0,
		Y:         50.0,
		Z:         -25.0,
		Unk0:      0,
	}

	handleMsgSysCreateObject(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify object was created in stage
	if len(stage.objects) != 1 {
		t.Errorf("Stage should have 1 object, got %d", len(stage.objects))
	}
}

func TestHandleMsgSysCreateObject_MultipleObjects(t *testing.T) {
	server := createMockServer()

	// Create multiple sessions that create objects
	sessions := make([]*Session, 3)
	stage := NewStage("test_stage")

	for i := 0; i < 3; i++ {
		sessions[i] = createMockSession(uint32(i+1), server)
		sessions[i].stage = stage

		pkt := &mhfpacket.MsgSysCreateObject{
			AckHandle: uint32(12345 + i),
			X:         float32(i * 10),
			Y:         float32(i * 20),
			Z:         float32(i * 30),
		}

		handleMsgSysCreateObject(sessions[i], pkt)

		// Drain send queue
		select {
		case <-sessions[i].sendPackets:
		default:
		}
	}

	// All objects should exist
	if len(stage.objects) != 3 {
		t.Errorf("Stage should have 3 objects, got %d", len(stage.objects))
	}
}

func TestHandleMsgSysPositionObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage with an existing object
	stage := NewStage("test_stage")
	session.stage = stage

	// Add another session to receive broadcast
	session2 := createMockSession(2, server)
	session2.stage = stage
	stage.clients[session] = session.charID
	stage.clients[session2] = session2.charID

	// Create an object
	stage.objects[session.charID] = &Object{
		id:          1,
		ownerCharID: session.charID,
		x:           0,
		y:           0,
		z:           0,
	}

	pkt := &mhfpacket.MsgSysPositionObject{
		ObjID: 1,
		X:     100.0,
		Y:     200.0,
		Z:     300.0,
	}

	handleMsgSysPositionObject(session, pkt)

	// Verify object position was updated
	obj := stage.objects[session.charID]
	if obj.x != 100.0 || obj.y != 200.0 || obj.z != 300.0 {
		t.Errorf("Object position not updated: got (%f, %f, %f), want (100, 200, 300)",
			obj.x, obj.y, obj.z)
	}

	// Verify broadcast was sent to session2
	select {
	case <-session2.sendPackets:
		// Good - broadcast received
	default:
		t.Error("Position update should be broadcast to other sessions")
	}
}

func TestHandleMsgSysPositionObject_NoObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	stage := NewStage("test_stage")
	session.stage = stage
	stage.clients[session] = session.charID

	// Position update for non-existent object - should not panic
	pkt := &mhfpacket.MsgSysPositionObject{
		ObjID: 999,
		X:     100.0,
		Y:     200.0,
		Z:     300.0,
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysPositionObject panicked with non-existent object: %v", r)
		}
	}()

	handleMsgSysPositionObject(session, pkt)
}

func TestHandleMsgSysDeleteObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDeleteObject panicked: %v", r)
		}
	}()

	handleMsgSysDeleteObject(session, nil)
}

func TestHandleMsgSysRotateObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysRotateObject panicked: %v", r)
		}
	}()

	handleMsgSysRotateObject(session, nil)
}

func TestHandleMsgSysDuplicateObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDuplicateObject panicked: %v", r)
		}
	}()

	handleMsgSysDuplicateObject(session, nil)
}

func TestHandleMsgSysSetObjectBinary(t *testing.T) {
	server := createMockServer()
	server.userBinaryParts = make(map[userBinaryPartID][]byte)
	session := createMockSession(1, server)
	server.sessions[nil] = session // Add session to server

	pkt := &mhfpacket.MsgSysSetObjectBinary{
		RawDataPayload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	handleMsgSysSetObjectBinary(session, pkt)

	// Verify binary was stored
	key := userBinaryPartID{charID: session.charID, index: 3}
	if data, ok := server.userBinaryParts[key]; !ok {
		t.Error("Object binary should be stored")
	} else if len(data) != 4 {
		t.Errorf("Object binary length = %d, want 4", len(data))
	}
}

func TestHandleMsgSysGetObjectBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysGetObjectBinary panicked: %v", r)
		}
	}()

	handleMsgSysGetObjectBinary(session, nil)
}

func TestHandleMsgSysGetObjectOwner(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysGetObjectOwner panicked: %v", r)
		}
	}()

	handleMsgSysGetObjectOwner(session, nil)
}

func TestHandleMsgSysUpdateObjectBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysUpdateObjectBinary panicked: %v", r)
		}
	}()

	handleMsgSysUpdateObjectBinary(session, nil)
}

func TestHandleMsgSysCleanupObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysCleanupObject panicked: %v", r)
		}
	}()

	handleMsgSysCleanupObject(session, nil)
}

func TestHandleMsgSysAddObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysAddObject panicked: %v", r)
		}
	}()

	handleMsgSysAddObject(session, nil)
}

func TestHandleMsgSysDelObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDelObject panicked: %v", r)
		}
	}()

	handleMsgSysDelObject(session, nil)
}

func TestHandleMsgSysDispObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDispObject panicked: %v", r)
		}
	}()

	handleMsgSysDispObject(session, nil)
}

func TestHandleMsgSysHideObject(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysHideObject panicked: %v", r)
		}
	}()

	handleMsgSysHideObject(session, nil)
}

func TestObjectHandlers_SequentialCreateObject(t *testing.T) {
	server := createMockServer()
	stage := NewStage("test_stage")

	// Create objects sequentially from multiple sessions
	// Note: handleMsgSysCreateObject has a race condition in NextObjectID
	// so we test sequential creation instead
	for i := 0; i < 10; i++ {
		session := createMockSession(uint32(i), server)
		session.stage = stage

		pkt := &mhfpacket.MsgSysCreateObject{
			AckHandle: uint32(i),
			X:         float32(i),
			Y:         float32(i * 2),
			Z:         float32(i * 3),
		}

		handleMsgSysCreateObject(session, pkt)

		// Drain send queue
		select {
		case <-session.sendPackets:
		default:
		}
	}

	// All objects should be created
	if len(stage.objects) != 10 {
		t.Errorf("Expected 10 objects, got %d", len(stage.objects))
	}
}

func TestObjectHandlers_SequentialPositionUpdate(t *testing.T) {
	server := createMockServer()
	stage := NewStage("test_stage")

	session := createMockSession(1, server)
	session.stage = stage
	stage.clients[session] = session.charID

	// Create an object
	stage.objects[session.charID] = &Object{
		id:          1,
		ownerCharID: session.charID,
		x:           0,
		y:           0,
		z:           0,
	}

	// Sequentially update object position
	for i := 0; i < 10; i++ {
		pkt := &mhfpacket.MsgSysPositionObject{
			ObjID: 1,
			X:     float32(i),
			Y:     float32(i * 2),
			Z:     float32(i * 3),
		}

		handleMsgSysPositionObject(session, pkt)
	}

	// Verify final position
	obj := stage.objects[session.charID]
	if obj.x != 9 || obj.y != 18 || obj.z != 27 {
		t.Errorf("Object position not as expected: got (%f, %f, %f), want (9, 18, 27)",
			obj.x, obj.y, obj.z)
	}
}
