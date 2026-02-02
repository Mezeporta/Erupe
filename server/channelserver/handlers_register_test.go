package channelserver

import (
	"testing"
)

func TestHandleMsgSysNotifyRegister(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysNotifyRegister panicked: %v", r)
		}
	}()

	handleMsgSysNotifyRegister(session, nil)
}

func TestGetRaviSemaphore_None(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	result := getRaviSemaphore(server)

	if result != nil {
		t.Error("Expected nil when no raviente semaphore exists")
	}
}

func TestGetRaviSemaphore_Found(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	// Create a raviente semaphore (matches prefix hs_l0u3B5 and suffix 3)
	sema := NewSemaphore(server, "hs_l0u3B53", 32)
	server.semaphore["hs_l0u3B53"] = sema

	result := getRaviSemaphore(server)

	if result == nil {
		t.Error("Expected to find raviente semaphore")
	}
	if result.id_semaphore != "hs_l0u3B53" {
		t.Errorf("Wrong semaphore returned: %s", result.id_semaphore)
	}
}

func TestGetRaviSemaphore_WrongSuffix(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	// Create a semaphore with wrong suffix
	sema := NewSemaphore(server, "hs_l0u3B51", 32)
	server.semaphore["hs_l0u3B51"] = sema

	result := getRaviSemaphore(server)

	if result != nil {
		t.Error("Should not match semaphore with wrong suffix")
	}
}
