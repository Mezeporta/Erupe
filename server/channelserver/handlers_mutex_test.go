package channelserver

import (
	"testing"
)

// Test that all mutex handlers don't panic (they are empty implementations)

func TestHandleMsgSysCreateMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysCreateMutex panicked: %v", r)
		}
	}()

	handleMsgSysCreateMutex(session, nil)
}

func TestHandleMsgSysCreateOpenMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysCreateOpenMutex panicked: %v", r)
		}
	}()

	handleMsgSysCreateOpenMutex(session, nil)
}

func TestHandleMsgSysDeleteMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDeleteMutex panicked: %v", r)
		}
	}()

	handleMsgSysDeleteMutex(session, nil)
}

func TestHandleMsgSysOpenMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysOpenMutex panicked: %v", r)
		}
	}()

	handleMsgSysOpenMutex(session, nil)
}

func TestHandleMsgSysCloseMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysCloseMutex panicked: %v", r)
		}
	}()

	handleMsgSysCloseMutex(session, nil)
}
