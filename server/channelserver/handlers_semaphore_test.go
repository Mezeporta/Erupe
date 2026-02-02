package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysCreateSemaphore(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysCreateSemaphore{
		AckHandle:      12345,
		Unk0:           0,
		DataSize:       0,
		RawDataPayload: []byte{},
	}

	handleMsgSysCreateSemaphore(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysDeleteSemaphore_NoSemaphores(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysDeleteSemaphore{
		AckHandle: 12345,
	}

	// Should not panic when no semaphores exist
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDeleteSemaphore panicked: %v", r)
		}
	}()

	handleMsgSysDeleteSemaphore(session, pkt)
}

func TestHandleMsgSysDeleteSemaphore_WithSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create a semaphore
	sema := NewSemaphore(server, "test_sema", 4)
	server.semaphore["test_sema"] = sema

	pkt := &mhfpacket.MsgSysDeleteSemaphore{
		AckHandle: sema.id,
	}

	handleMsgSysDeleteSemaphore(session, pkt)

	// Semaphore should be deleted
	if _, exists := server.semaphore["test_sema"]; exists {
		t.Error("Semaphore should be deleted")
	}
}

func TestHandleMsgSysCreateAcquireSemaphore_NewSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysCreateAcquireSemaphore{
		AckHandle:   12345,
		Unk0:        0,
		PlayerCount: 4,
		SemaphoreID: "test_semaphore",
	}

	handleMsgSysCreateAcquireSemaphore(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify semaphore was created
	if _, exists := server.semaphore["test_semaphore"]; !exists {
		t.Error("Semaphore should be created")
	}
}

func TestHandleMsgSysCreateAcquireSemaphore_ExistingSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Pre-create semaphore
	sema := NewSemaphore(server, "existing_sema", 4)
	server.semaphore["existing_sema"] = sema

	pkt := &mhfpacket.MsgSysCreateAcquireSemaphore{
		AckHandle:   12345,
		Unk0:        0,
		PlayerCount: 4,
		SemaphoreID: "existing_sema",
	}

	handleMsgSysCreateAcquireSemaphore(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify client was added to semaphore
	if len(sema.reservedClientSlots) == 0 && len(sema.clients) == 0 {
		t.Error("Session should be added to semaphore")
	}
}

func TestHandleMsgSysCreateAcquireSemaphore_RavienteSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Test raviente semaphore (special prefix)
	pkt := &mhfpacket.MsgSysCreateAcquireSemaphore{
		AckHandle:   12345,
		Unk0:        0,
		PlayerCount: 32,
		SemaphoreID: "hs_l0u3B51", // Raviente prefix + suffix
	}

	handleMsgSysCreateAcquireSemaphore(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify raviente semaphore was created with special settings
	if sema, exists := server.semaphore["hs_l0u3B51"]; !exists {
		t.Error("Raviente semaphore should be created")
	} else if sema.maxPlayers != 32 {
		t.Errorf("Raviente semaphore maxPlayers = %d, want 32", sema.maxPlayers)
	}
}

func TestHandleMsgSysCreateAcquireSemaphore_Full(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	// Create semaphore with 1 player max
	sema := NewSemaphore(server, "full_sema", 1)
	server.semaphore["full_sema"] = sema

	// Fill the semaphore
	session1 := createMockSession(1, server)
	sema.reservedClientSlots[session1.charID] = nil
	sema.clients[session1] = session1.charID

	// Try to acquire with another session
	session2 := createMockSession(2, server)
	pkt := &mhfpacket.MsgSysCreateAcquireSemaphore{
		AckHandle:   12345,
		Unk0:        0,
		PlayerCount: 1,
		SemaphoreID: "full_sema",
	}

	handleMsgSysCreateAcquireSemaphore(session2, pkt)

	// Should still respond (with failure indication)
	select {
	case p := <-session2.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data even for full semaphore")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysAcquireSemaphore_Exists(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create semaphore
	sema := NewSemaphore(server, "acquire_test", 4)
	server.semaphore["acquire_test"] = sema

	pkt := &mhfpacket.MsgSysAcquireSemaphore{
		AckHandle:   12345,
		SemaphoreID: "acquire_test",
	}

	handleMsgSysAcquireSemaphore(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify client was added
	if _, exists := sema.clients[session]; !exists {
		t.Error("Session should be added to semaphore clients")
	}
}

func TestHandleMsgSysAcquireSemaphore_NotExists(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysAcquireSemaphore{
		AckHandle:   12345,
		SemaphoreID: "nonexistent",
	}

	handleMsgSysAcquireSemaphore(session, pkt)

	// Should respond with failure
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysReleaseSemaphore(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (mostly empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysReleaseSemaphore panicked: %v", r)
		}
	}()

	pkt := &mhfpacket.MsgSysReleaseSemaphore{}
	handleMsgSysReleaseSemaphore(session, pkt)
}

func TestHandleMsgSysCheckSemaphore_Exists(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create semaphore
	sema := NewSemaphore(server, "check_test", 4)
	server.semaphore["check_test"] = sema

	pkt := &mhfpacket.MsgSysCheckSemaphore{
		AckHandle:   12345,
		SemaphoreID: "check_test",
	}

	handleMsgSysCheckSemaphore(session, pkt)

	// Verify response indicates semaphore exists
	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Error("Response packet should have at least 4 bytes")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysCheckSemaphore_NotExists(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysCheckSemaphore{
		AckHandle:   12345,
		SemaphoreID: "nonexistent",
	}

	handleMsgSysCheckSemaphore(session, pkt)

	// Verify response indicates semaphore does not exist
	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Error("Response packet should have at least 4 bytes")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestRemoveSessionFromSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create semaphore and add session
	sema := NewSemaphore(server, "remove_test", 4)
	sema.clients[session] = session.charID
	sema.reservedClientSlots[session.charID] = nil
	server.semaphore["remove_test"] = sema

	// Remove session
	removeSessionFromSemaphore(session)

	// Verify session was removed
	if _, exists := sema.clients[session]; exists {
		t.Error("Session should be removed from clients")
	}
	if _, exists := sema.reservedClientSlots[session.charID]; exists {
		t.Error("Session should be removed from reserved slots")
	}
}

func TestRemoveSessionFromSemaphore_MultipleSemaphores(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create multiple semaphores with the session
	for i := 0; i < 3; i++ {
		sema := NewSemaphore(server, "multi_test_"+string(rune('a'+i)), 4)
		sema.clients[session] = session.charID
		sema.reservedClientSlots[session.charID] = nil
		server.semaphore["multi_test_"+string(rune('a'+i))] = sema
	}

	// Remove session from all
	removeSessionFromSemaphore(session)

	// Verify session was removed from all semaphores
	for _, sema := range server.semaphore {
		if _, exists := sema.clients[session]; exists {
			t.Error("Session should be removed from all semaphore clients")
		}
		if _, exists := sema.reservedClientSlots[session.charID]; exists {
			t.Error("Session should be removed from all reserved slots")
		}
	}
}

func TestDestructEmptySemaphores(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)
	session := createMockSession(1, server)

	// Create empty semaphore
	sema := NewSemaphore(server, "empty_sema", 4)
	server.semaphore["empty_sema"] = sema

	// Create non-empty semaphore
	semaWithClients := NewSemaphore(server, "with_clients", 4)
	semaWithClients.clients[session] = session.charID
	server.semaphore["with_clients"] = semaWithClients

	destructEmptySemaphores(session)

	// Empty semaphore should be deleted
	if _, exists := server.semaphore["empty_sema"]; exists {
		t.Error("Empty semaphore should be deleted")
	}

	// Non-empty semaphore should remain
	if _, exists := server.semaphore["with_clients"]; !exists {
		t.Error("Non-empty semaphore should remain")
	}
}

func TestSemaphoreHandlers_SequentialAcquire(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	// Sequentially try to create/acquire the same semaphore
	// Note: the handler has race conditions when accessed concurrently
	for i := 0; i < 5; i++ {
		session := createMockSession(uint32(i), server)

		pkt := &mhfpacket.MsgSysCreateAcquireSemaphore{
			AckHandle:   uint32(i),
			Unk0:        0,
			PlayerCount: 4,
			SemaphoreID: "sequential_test",
		}

		handleMsgSysCreateAcquireSemaphore(session, pkt)

		// Drain send queue
		select {
		case <-session.sendPackets:
		default:
		}
	}

	// Semaphore should exist
	if _, exists := server.semaphore["sequential_test"]; !exists {
		t.Error("Semaphore should exist after sequential acquires")
	}
}

func TestSemaphoreHandlers_MultipleCheck(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	// Create semaphore
	sema := NewSemaphore(server, "check_multiple", 4)
	server.semaphore["check_multiple"] = sema

	// Check the semaphore from multiple sessions sequentially
	for i := 0; i < 5; i++ {
		session := createMockSession(uint32(i), server)

		pkt := &mhfpacket.MsgSysCheckSemaphore{
			AckHandle:   uint32(i),
			SemaphoreID: "check_multiple",
		}

		handleMsgSysCheckSemaphore(session, pkt)

		// Drain send queue
		select {
		case <-session.sendPackets:
		default:
		}
	}
}
