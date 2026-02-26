package channelserver

import (
	"errors"
	"testing"
)

func TestLoadCharacterData_Success(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["test_col"] = []byte{0xAA, 0xBB, 0xCC}
	server.charRepo = charRepo
	session := createMockSession(1, server)

	loadCharacterData(session, 100, "test_col", nil)

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestLoadCharacterData_EmptyUsesDefault(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	session := createMockSession(1, server)

	defaultData := []byte{0x01, 0x02, 0x03}
	loadCharacterData(session, 100, "missing_col", defaultData)

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestLoadCharacterData_Error(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.loadColumnErr = errors.New("db error")
	server.charRepo = charRepo
	session := createMockSession(1, server)

	defaultData := []byte{0xFF}
	loadCharacterData(session, 100, "test_col", defaultData)

	// Should still send a response (with default data)
	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data even on error")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestSaveCharacterData_Success(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	session := createMockSession(1, server)

	data := []byte{0x01, 0x02, 0x03}
	saveCharacterData(session, 100, "test_col", data, 100)

	// Should save and ack
	if saved := charRepo.columns["test_col"]; saved == nil {
		t.Error("Data should be saved to repo")
	}

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestSaveCharacterData_TooLarge(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	session := createMockSession(1, server)

	data := make([]byte, 200)
	saveCharacterData(session, 100, "test_col", data, 50)

	// Should fail with ack
	if _, ok := charRepo.columns["test_col"]; ok {
		t.Error("Data should NOT be saved when too large")
	}

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data")
		}
	default:
		t.Fatal("Should queue a fail ack")
	}
}

func TestSaveCharacterData_SaveError(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.saveErr = errors.New("save failed")
	server.charRepo = charRepo
	session := createMockSession(1, server)

	data := []byte{0x01}
	saveCharacterData(session, 100, "test_col", data, 100)

	// Should still queue a fail ack
	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response packet should have data")
		}
	default:
		t.Fatal("Should queue a fail ack on save error")
	}
}

func TestSaveCharacterData_NoMaxSize(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	session := createMockSession(1, server)

	data := make([]byte, 5000)
	saveCharacterData(session, 100, "test_col", data, 0)

	// maxSize=0 means no limit
	if saved := charRepo.columns["test_col"]; saved == nil {
		t.Error("Data should be saved when maxSize is 0 (no limit)")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Should queue success ack")
	}
}

func TestDoAckEarthSucceed(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.EarthID = 42
	session := createMockSession(1, server)

	doAckEarthSucceed(session, 100, nil)

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Response should have data")
		}
	default:
		t.Fatal("Should queue a packet")
	}
}

func TestUpdateRights(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{}
	userRepo.rights = 30
	server.userRepo = userRepo
	session := createMockSession(1, server)

	updateRights(session)

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Should queue MsgSysUpdateRight")
		}
	default:
		t.Fatal("updateRights should queue a packet")
	}
}

// Tests consolidated from handlers_coverage3_test.go

func TestStubHelpers(t *testing.T) {
	server := createMockServer()

	t.Run("stubEnumerateNoResults", func(t *testing.T) {
		session := createMockSession(1, server)
		stubEnumerateNoResults(session, 1)
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckBufSucceed", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckBufSucceed(session, 1, []byte{0x01, 0x02, 0x03})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckBufFail", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckBufFail(session, 1, []byte{0x01, 0x02, 0x03})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckSimpleSucceed", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckSimpleSucceed(session, 1, []byte{0x00, 0x00, 0x00, 0x00})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckSimpleFail", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckSimpleFail(session, 1, []byte{0x00, 0x00, 0x00, 0x00})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}

func TestUpdateRights_Error(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{rightsErr: errors.New("db error")}
	server.userRepo = userRepo
	session := createMockSession(1, server)

	// Should not panic, falls back to rights=2
	updateRights(session)

	select {
	case pkt := <-session.sendPackets:
		if pkt.data == nil {
			t.Fatal("Should queue MsgSysUpdateRight even on error")
		}
	default:
		t.Fatal("updateRights should queue a packet even on error")
	}
}
