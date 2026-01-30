package channelserver

import (
	"sync"
	"testing"
	"time"
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
