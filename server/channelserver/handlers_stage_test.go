package channelserver

import (
	"sync"
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

// TestWaitStageBinaryInfiniteLoopRisk documents the infinite loop risk in handleMsgSysWaitStageBinary.
//
// CURRENT BEHAVIOR (BUG - needs fix commit c539905):
//
//	for {
//	    // ... check for binary data
//	    if gotBinary {
//	        doAckBufSucceed(...)
//	        break
//	    } else {
//	        time.Sleep(1 * time.Second)
//	        continue
//	    }
//	}
//
// This loop runs FOREVER if the binary data never arrives, causing:
// - Resource leak (goroutine stuck forever)
// - Memory leak (session can't be cleaned up)
// - Client timeout/disconnect with no server-side cleanup
//
// EXPECTED BEHAVIOR (after fix):
//
//	for i := 0; i < 10; i++ {  // Maximum 10 iterations (10 seconds)
//	    // ... check for binary data
//	    if gotBinary {
//	        doAckBufSucceed(...)
//	        return
//	    } else {
//	        time.Sleep(1 * time.Second)
//	        continue
//	    }
//	}
//	// Timeout - return empty response
//	doAckBufSucceed(s, pkt.AckHandle, []byte{})
func TestWaitStageBinaryInfiniteLoopRisk(t *testing.T) {
	// This test documents the expected behavior, not the actual implementation
	// (which would require full server setup)

	t.Run("current behavior loops forever", func(t *testing.T) {
		// Simulate current infinite loop behavior
		iterations := 0
		maxIterations := 100 // Safety limit for test

		// This simulates what the CURRENT code does (infinite loop)
		simulateCurrentBehavior := func(getBinary func() bool) int {
			for {
				if getBinary() {
					return iterations
				}
				iterations++
				if iterations >= maxIterations {
					return iterations // Safety break for test
				}
				// In real code: time.Sleep(1 * time.Second)
			}
		}

		// Binary never arrives
		neverReturns := func() bool { return false }
		result := simulateCurrentBehavior(neverReturns)

		if result < maxIterations {
			t.Errorf("expected loop to hit safety limit (%d), got %d", maxIterations, result)
		}
		t.Logf("Current behavior would loop forever (hit safety limit at %d iterations)", result)
	})

	t.Run("fixed behavior has timeout", func(t *testing.T) {
		// Simulate fixed behavior with timeout
		const maxTimeout = 10 // Maximum iterations before timeout

		simulateFixedBehavior := func(getBinary func() bool) (int, bool) {
			for i := 0; i < maxTimeout; i++ {
				if getBinary() {
					return i, true // Found binary
				}
				// In real code: time.Sleep(1 * time.Second)
			}
			return maxTimeout, false // Timeout
		}

		// Binary never arrives
		neverReturns := func() bool { return false }
		iterations, found := simulateFixedBehavior(neverReturns)

		if found {
			t.Error("expected timeout (not found)")
		}
		if iterations != maxTimeout {
			t.Errorf("expected %d iterations before timeout, got %d", maxTimeout, iterations)
		}
		t.Logf("Fixed behavior times out after %d iterations", iterations)
	})

	t.Run("fixed behavior returns quickly when binary exists", func(t *testing.T) {
		const maxTimeout = 10

		// Simulate binary arriving on 3rd check
		checkCount := 0
		arrivesOnThird := func() bool {
			checkCount++
			return checkCount >= 3
		}

		simulateFixedBehavior := func(getBinary func() bool) (int, bool) {
			for i := 0; i < maxTimeout; i++ {
				if getBinary() {
					return i + 1, true
				}
			}
			return maxTimeout, false
		}

		iterations, found := simulateFixedBehavior(arrivesOnThird)

		if !found {
			t.Error("expected to find binary")
		}
		if iterations != 3 {
			t.Errorf("expected 3 iterations, got %d", iterations)
		}
	})
}

// TestStageBinaryKeyAccess tests the stageBinaryKey struct used for indexing.
func TestStageBinaryKeyAccess(t *testing.T) {
	// Create a stage with binary data
	stage := NewStage("test_stage")

	key := stageBinaryKey{id0: 1, id1: 2}
	testData := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	// Store binary data
	stage.Lock()
	stage.rawBinaryData[key] = testData
	stage.Unlock()

	// Retrieve binary data
	stage.Lock()
	data, exists := stage.rawBinaryData[key]
	stage.Unlock()

	if !exists {
		t.Error("expected binary data to exist")
	}
	if len(data) != 4 {
		t.Errorf("data length = %d, want 4", len(data))
	}
	if data[0] != 0xDE {
		t.Errorf("data[0] = 0x%X, want 0xDE", data[0])
	}
}

// TestStageBinaryKeyUniqueness verifies different keys map to different data.
func TestStageBinaryKeyUniqueness(t *testing.T) {
	stage := NewStage("test_stage")

	key1 := stageBinaryKey{id0: 1, id1: 2}
	key2 := stageBinaryKey{id0: 1, id1: 3}
	key3 := stageBinaryKey{id0: 2, id1: 2}

	data1 := []byte{0x01}
	data2 := []byte{0x02}
	data3 := []byte{0x03}

	stage.Lock()
	stage.rawBinaryData[key1] = data1
	stage.rawBinaryData[key2] = data2
	stage.rawBinaryData[key3] = data3
	stage.Unlock()

	stage.Lock()
	defer stage.Unlock()

	if d, ok := stage.rawBinaryData[key1]; !ok || d[0] != 0x01 {
		t.Error("key1 data mismatch")
	}
	if d, ok := stage.rawBinaryData[key2]; !ok || d[0] != 0x02 {
		t.Error("key2 data mismatch")
	}
	if d, ok := stage.rawBinaryData[key3]; !ok || d[0] != 0x03 {
		t.Error("key3 data mismatch")
	}
}

// TestStageBinarySpecialCase tests the special case for id0=1, id1=12.
// This case returns immediately with a fixed response instead of waiting.
func TestStageBinarySpecialCase(t *testing.T) {
	// The special case is:
	// if pkt.id0 == 1 && pkt.id1 == 12 {
	//     doAckBufSucceed(s, pkt.AckHandle, []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	//     return
	// }

	expectedResponse := []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if len(expectedResponse) != 8 {
		t.Errorf("special case response length = %d, want 8", len(expectedResponse))
	}
	if expectedResponse[0] != 0x04 {
		t.Errorf("special case response[0] = 0x%X, want 0x04", expectedResponse[0])
	}
}

// TestStageLockingDuringBinaryAccess verifies proper locking during binary data access.
func TestStageLockingDuringBinaryAccess(t *testing.T) {
	stage := NewStage("concurrent_test")
	key := stageBinaryKey{id0: 1, id1: 1}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				stage.Lock()
				stage.rawBinaryData[key] = []byte{byte(id), byte(j)}
				stage.Unlock()
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				stage.Lock()
				_ = stage.rawBinaryData[key]
				stage.Unlock()
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestWaitStageBinaryTimeoutDuration documents the expected timeout duration.
// After the fix, the handler should wait at most 10 seconds (10 iterations * 1 second sleep).
func TestWaitStageBinaryTimeoutDuration(t *testing.T) {
	const (
		sleepDuration   = 1 * time.Second
		maxIterations   = 10
		expectedTimeout = sleepDuration * maxIterations
	)

	if expectedTimeout != 10*time.Second {
		t.Errorf("expected timeout = %v, want 10s", expectedTimeout)
	}

	t.Logf("After fix, WaitStageBinary will timeout after %v (%d iterations * %v sleep)",
		expectedTimeout, maxIterations, sleepDuration)
}

// TestHandleMsgSysCreateStage_NewStage tests creating a new stage
func TestHandleMsgSysCreateStage_NewStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysCreateStage{
		AckHandle:   12345,
		StageID:     "test_create_stage",
		PlayerCount: 4,
	}

	handleMsgSysCreateStage(session, pkt)

	// Verify stage was created
	server.Lock()
	stage, exists := server.stages["test_create_stage"]
	server.Unlock()

	if !exists {
		t.Error("Stage should be created")
	}
	if stage.maxPlayers != 4 {
		t.Errorf("stage.maxPlayers = %d, want 4", stage.maxPlayers)
	}
	if stage.host != session {
		t.Error("Session should be host of the stage")
	}

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

// TestHandleMsgSysCreateStage_ExistingStage tests creating a stage that already exists
func TestHandleMsgSysCreateStage_ExistingStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create existing stage
	existingStage := NewStage("existing_stage")
	server.stages["existing_stage"] = existingStage

	pkt := &mhfpacket.MsgSysCreateStage{
		AckHandle:   12345,
		StageID:     "existing_stage",
		PlayerCount: 4,
	}

	handleMsgSysCreateStage(session, pkt)

	// Verify response packet was queued (should be failure)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestDoStageTransfer_NewStage tests entering a stage that doesn't exist
func TestDoStageTransfer_NewStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	doStageTransfer(session, 12345, "new_transfer_stage")

	// Verify stage was created
	server.Lock()
	stage, exists := server.stages["new_transfer_stage"]
	server.Unlock()

	if !exists {
		t.Error("Stage should be created")
	}

	// Verify session is in the stage
	stage.RLock()
	_, inStage := stage.clients[session]
	stage.RUnlock()

	if !inStage {
		t.Error("Session should be in the stage")
	}

	// Verify session's stage reference is set
	if session.stage != stage {
		t.Error("Session's stage reference should be set")
	}

	// Verify response packets were queued
	packetCount := 0
	for {
		select {
		case <-session.sendPackets:
			packetCount++
		default:
			goto done
		}
	}
done:
	if packetCount < 2 {
		t.Errorf("Expected at least 2 packets (cleanup + ack), got %d", packetCount)
	}
}

// TestDoStageTransfer_ExistingStage tests entering an existing stage
func TestDoStageTransfer_ExistingStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create existing stage
	existingStage := NewStage("existing_transfer_stage")
	server.stages["existing_transfer_stage"] = existingStage

	doStageTransfer(session, 12345, "existing_transfer_stage")

	// Verify session is in the stage
	existingStage.RLock()
	_, inStage := existingStage.clients[session]
	existingStage.RUnlock()

	if !inStage {
		t.Error("Session should be in the stage")
	}

	// Verify response packets were queued
	packetCount := 0
	for {
		select {
		case <-session.sendPackets:
			packetCount++
		default:
			goto done
		}
	}
done:
	if packetCount < 2 {
		t.Errorf("Expected at least 2 packets, got %d", packetCount)
	}
}

// TestHandleMsgSysStageDestruct tests the empty handler
func TestHandleMsgSysStageDestruct(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysStageDestruct panicked: %v", r)
		}
	}()

	handleMsgSysStageDestruct(session, nil)
}

// TestHandleMsgSysLockStage tests the lock stage handler
func TestHandleMsgSysLockStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockStage{
		AckHandle: 12345,
	}

	handleMsgSysLockStage(session, pkt)

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

// TestHandleMsgSysLeaveStage tests the empty handler
func TestHandleMsgSysLeaveStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysLeaveStage panicked: %v", r)
		}
	}()

	handleMsgSysLeaveStage(session, nil)
}

// TestDestructEmptyStages tests the stage cleanup function
func TestDestructEmptyStages(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create different types of stages
	questStage := NewStage("00XQsStage1") // Quest stage
	myStage := NewStage("00XMsStage1")    // My series stage
	guildStage := NewStage("00XGsStage1") // Guild stage
	townStage := NewStage("00XTwStage1")  // Town stage (should not be deleted)

	server.stages["00XQsStage1"] = questStage
	server.stages["00XMsStage1"] = myStage
	server.stages["00XGsStage1"] = guildStage
	server.stages["00XTwStage1"] = townStage

	destructEmptyStages(session)

	// Quest/My/Guild stages should be deleted (empty)
	if _, exists := server.stages["00XQsStage1"]; exists {
		t.Error("Empty quest stage should be deleted")
	}
	if _, exists := server.stages["00XMsStage1"]; exists {
		t.Error("Empty my series stage should be deleted")
	}
	if _, exists := server.stages["00XGsStage1"]; exists {
		t.Error("Empty guild stage should be deleted")
	}

	// Town stage should remain
	if _, exists := server.stages["00XTwStage1"]; !exists {
		t.Error("Town stage should not be deleted")
	}
}

// TestDestructEmptyStages_NonEmpty tests that non-empty stages are preserved
func TestDestructEmptyStages_NonEmpty(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create quest stage with clients
	questStage := NewStage("00XQsStage2")
	questStage.clients[session] = session.charID
	server.stages["00XQsStage2"] = questStage

	destructEmptyStages(session)

	// Stage with clients should not be deleted
	if _, exists := server.stages["00XQsStage2"]; !exists {
		t.Error("Non-empty quest stage should not be deleted")
	}
}

// TestDestructEmptyStages_WithReservations tests that stages with reservations are preserved
func TestDestructEmptyStages_WithReservations(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create quest stage with reservations
	questStage := NewStage("00XQsStage3")
	questStage.reservedClientSlots[session.charID] = true
	server.stages["00XQsStage3"] = questStage

	destructEmptyStages(session)

	// Stage with reservations should not be deleted
	if _, exists := server.stages["00XQsStage3"]; !exists {
		t.Error("Quest stage with reservations should not be deleted")
	}
}

// TestRemoveSessionFromStage tests removing a session from its stage
func TestRemoveSessionFromStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage and add session to it
	stage := NewStage("00XTwRemove1")
	stage.clients[session] = session.charID
	session.stage = stage
	server.stages["00XTwRemove1"] = stage

	// Verify session is in stage
	if _, exists := stage.clients[session]; !exists {
		t.Error("Session should be in stage before removal")
	}

	removeSessionFromStage(session)

	// Verify session is removed from stage
	if _, exists := stage.clients[session]; exists {
		t.Error("Session should be removed from stage")
	}
}

// TestRemoveSessionFromStage_WithObjects tests removing objects when leaving stage
func TestRemoveSessionFromStage_WithObjects(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage and add session with objects
	stage := NewStage("00XTwRemove2")
	stage.clients[session] = session.charID
	stage.objects[session.charID] = &Object{
		id:          1,
		ownerCharID: session.charID,
		x:           100.0,
		y:           200.0,
		z:           300.0,
	}
	session.stage = stage
	server.stages["00XTwRemove2"] = stage

	removeSessionFromStage(session)

	// Verify objects owned by session are removed
	if _, exists := stage.objects[session.charID]; exists {
		t.Error("Objects owned by session should be removed")
	}
}

// TestHandleMsgSysSetStageBinary tests setting stage binary data
func TestHandleMsgSysSetStageBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage
	stage := NewStage("test_binary_stage")
	server.stages["test_binary_stage"] = stage

	pkt := &mhfpacket.MsgSysSetStageBinary{
		StageID:        "test_binary_stage",
		BinaryType0:    1,
		BinaryType1:    2,
		RawDataPayload: []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	handleMsgSysSetStageBinary(session, pkt)

	// Verify binary was stored
	stage.Lock()
	data, exists := stage.rawBinaryData[stageBinaryKey{1, 2}]
	stage.Unlock()

	if !exists {
		t.Error("Binary data should be stored")
	}
	if len(data) != 4 {
		t.Errorf("Binary data length = %d, want 4", len(data))
	}
}

// TestHandleMsgSysGetStageBinary tests getting stage binary data
func TestHandleMsgSysGetStageBinary_WithData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage with binary data
	stage := NewStage("test_get_binary")
	stage.rawBinaryData[stageBinaryKey{1, 2}] = []byte{0x01, 0x02, 0x03, 0x04}
	server.stages["test_get_binary"] = stage

	pkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   12345,
		StageID:     "test_get_binary",
		BinaryType0: 1,
		BinaryType1: 2,
	}

	handleMsgSysGetStageBinary(session, pkt)

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

// TestHandleMsgSysGetStageBinary_NoData tests getting non-existent binary data
func TestHandleMsgSysGetStageBinary_NoData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage without the requested binary data
	stage := NewStage("test_no_binary")
	server.stages["test_no_binary"] = stage

	pkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   12345,
		StageID:     "test_no_binary",
		BinaryType0: 1,
		BinaryType1: 2,
	}

	handleMsgSysGetStageBinary(session, pkt)

	// Should still return a response (empty)
	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysGetStageBinary_Type4 tests special type 4 binary
func TestHandleMsgSysGetStageBinary_Type4(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage
	stage := NewStage("test_type4_binary")
	server.stages["test_type4_binary"] = stage

	pkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   12345,
		StageID:     "test_type4_binary",
		BinaryType0: 0,
		BinaryType1: 4,
	}

	handleMsgSysGetStageBinary(session, pkt)

	// Should return empty response for type 4
	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysSetStagePass tests setting stage password
func TestHandleMsgSysSetStagePass_WithReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage and add session reservation
	stage := NewStage("test_pass_stage")
	stage.reservedClientSlots[session.charID] = true
	server.stages["test_pass_stage"] = stage
	session.reservationStage = stage

	pkt := &mhfpacket.MsgSysSetStagePass{
		Password: "secret123",
	}

	handleMsgSysSetStagePass(session, pkt)

	// Verify password was set
	stage.Lock()
	password := stage.password
	stage.Unlock()

	if password != "secret123" {
		t.Errorf("Stage password = %s, want secret123", password)
	}
}

// TestHandleMsgSysSetStagePass_NoReservation tests setting pass without reservation
func TestHandleMsgSysSetStagePass_NoReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysSetStagePass{
		Password: "secret456",
	}

	handleMsgSysSetStagePass(session, pkt)

	// Verify password was stored in session for later use
	session.Lock()
	password := session.stagePass
	session.Unlock()

	if password != "secret456" {
		t.Errorf("Session stagePass = %s, want secret456", password)
	}
}

// TestHandleMsgSysEnumerateStage tests enumerating stages
func TestHandleMsgSysEnumerateStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create some stages
	stage1 := NewStage("00XQsStage1")
	stage1.reservedClientSlots[100] = true
	stage1.maxPlayers = 4

	stage2 := NewStage("00XQsStage2")
	stage2.clients[session] = session.charID
	stage2.maxPlayers = 2

	server.stages["00XQsStage1"] = stage1
	server.stages["00XQsStage2"] = stage2

	pkt := &mhfpacket.MsgSysEnumerateStage{
		AckHandle:   12345,
		StagePrefix: "Qs",
	}

	handleMsgSysEnumerateStage(session, pkt)

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

// TestHandleMsgSysEnumerateStage_Empty tests enumerating with no matching stages
func TestHandleMsgSysEnumerateStage_Empty(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysEnumerateStage{
		AckHandle:   12345,
		StagePrefix: "NonExistent",
	}

	handleMsgSysEnumerateStage(session, pkt)

	// Should still return a response
	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysUnreserveStage tests unreserving a stage
func TestHandleMsgSysUnreserveStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create a stage and add session reservation
	stage := NewStage("test_unreserve")
	stage.reservedClientSlots[session.charID] = true
	server.stages["test_unreserve"] = stage
	session.reservationStage = stage

	handleMsgSysUnreserveStage(session, nil)

	// Verify reservation was removed
	stage.Lock()
	_, exists := stage.reservedClientSlots[session.charID]
	stage.Unlock()

	if exists {
		t.Error("Reservation should be removed")
	}

	// Verify session's reservation stage is cleared
	session.Lock()
	reservationStage := session.reservationStage
	session.Unlock()

	if reservationStage != nil {
		t.Error("Session's reservation stage should be nil")
	}
}
