package channelserver

import (
	"errors"
	"testing"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

func TestHandleMsgMhfLoadPlateData(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["platedata"] = []byte{0x01, 0x02, 0x03}
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadPlateData{AckHandle: 100}
	handleMsgMhfLoadPlateData(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfLoadPlateData_Empty(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	// No platedata column set — loadCharacterData uses nil default
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadPlateData{AckHandle: 100}
	handleMsgMhfLoadPlateData(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSavePlateData_OversizedPayload(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSavePlateData{
		AckHandle:      100,
		RawDataPayload: make([]byte, plateDataMaxPayload+1),
		IsDataDiff:     false,
	}
	handleMsgMhfSavePlateData(session, pkt)

	// Should still get ACK
	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	// Data should NOT have been saved
	if charRepo.columns["platedata"] != nil {
		t.Error("Expected platedata to NOT be saved when oversized")
	}
}

func TestHandleMsgMhfSavePlateData_FullSave(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	payload := []byte{0x10, 0x20, 0x30, 0x40}
	pkt := &mhfpacket.MsgMhfSavePlateData{
		AckHandle:      100,
		RawDataPayload: payload,
		IsDataDiff:     false,
	}
	handleMsgMhfSavePlateData(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	saved := charRepo.columns["platedata"]
	if saved == nil {
		t.Fatal("Expected platedata to be saved")
	}
	if len(saved) != len(payload) {
		t.Errorf("Expected saved data length %d, got %d", len(payload), len(saved))
	}
}

func TestHandleMsgMhfSavePlateData_DiffPath_LoadError(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	charRepo.loadColumnErr = errors.New("load failed")
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSavePlateData{
		AckHandle:      100,
		RawDataPayload: []byte{0x01},
		IsDataDiff:     true,
	}
	handleMsgMhfSavePlateData(session, pkt)

	select {
	case <-session.sendPackets:
		// returns ACK even on error
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSavePlateData_DiffPath_SaveError(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	// Provide compressed data so decompress works
	original := make([]byte, 100)
	compressed, _ := nullcomp.Compress(original)
	charRepo.columns["platedata"] = compressed
	charRepo.saveErr = errors.New("save failed")
	server.charRepo = charRepo

	session := createMockSession(1, server)

	// Build a valid diff payload: matchCount=2 (offset becomes 1), diffCount=2 (means 1 byte), then 1 data byte
	diffPayload := []byte{2, 2, 0xAA}
	pkt := &mhfpacket.MsgMhfSavePlateData{
		AckHandle:      100,
		RawDataPayload: diffPayload,
		IsDataDiff:     true,
	}
	handleMsgMhfSavePlateData(session, pkt)

	select {
	case <-session.sendPackets:
		// returns ACK even on save error
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfLoadPlateBox(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["platebox"] = []byte{0xAA, 0xBB}
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadPlateBox{AckHandle: 100}
	handleMsgMhfLoadPlateBox(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSavePlateBox_OversizedPayload(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSavePlateBox{
		AckHandle:      100,
		RawDataPayload: make([]byte, plateBoxMaxPayload+1),
		IsDataDiff:     false,
	}
	handleMsgMhfSavePlateBox(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	if charRepo.columns["platebox"] != nil {
		t.Error("Expected platebox to NOT be saved when oversized")
	}
}

func TestHandleMsgMhfSavePlateBox_FullSave(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	payload := []byte{0xCC, 0xDD}
	pkt := &mhfpacket.MsgMhfSavePlateBox{
		AckHandle:      100,
		RawDataPayload: payload,
		IsDataDiff:     false,
	}
	handleMsgMhfSavePlateBox(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	if charRepo.columns["platebox"] == nil {
		t.Fatal("Expected platebox to be saved")
	}
}

func TestHandleMsgMhfSavePlateBox_DiffPath(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	// Provide compressed data
	original := make([]byte, 100)
	compressed, _ := nullcomp.Compress(original)
	charRepo.columns["platebox"] = compressed
	server.charRepo = charRepo

	session := createMockSession(1, server)

	// Valid diff: matchCount=2 (offset becomes 1), diffCount=2 (1 byte), data byte
	diffPayload := []byte{2, 2, 0xBB}
	pkt := &mhfpacket.MsgMhfSavePlateBox{
		AckHandle:      100,
		RawDataPayload: diffPayload,
		IsDataDiff:     true,
	}
	handleMsgMhfSavePlateBox(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfLoadPlateMyset(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadPlateMyset{AckHandle: 100}
	handleMsgMhfLoadPlateMyset(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSavePlateMyset_OversizedPayload(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSavePlateMyset{
		AckHandle:      100,
		RawDataPayload: make([]byte, plateMysetMaxPayload+1),
	}
	handleMsgMhfSavePlateMyset(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	if charRepo.columns["platemyset"] != nil {
		t.Error("Expected platemyset to NOT be saved when oversized")
	}
}

func TestHandleMsgMhfSavePlateMyset_Success(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(1, server)

	payload := make([]byte, plateMysetDefaultLen)
	payload[0] = 0xFF
	pkt := &mhfpacket.MsgMhfSavePlateMyset{
		AckHandle:      100,
		RawDataPayload: payload,
	}
	handleMsgMhfSavePlateMyset(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	if charRepo.columns["platemyset"] == nil {
		t.Fatal("Expected platemyset to be saved")
	}
	if charRepo.columns["platemyset"][0] != 0xFF {
		t.Error("Expected first byte to be 0xFF")
	}
}

func TestHandleMsgMhfSavePlateData_CacheInvalidation(t *testing.T) {
	server := createMockServer()
	server.userBinary = NewUserBinaryStore()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	session := createMockSession(42, server)

	// Pre-populate the cache
	server.userBinary.Set(42, 2, []byte{0x01})
	server.userBinary.Set(42, 3, []byte{0x02})

	pkt := &mhfpacket.MsgMhfSavePlateData{
		AckHandle:      100,
		RawDataPayload: []byte{0x10},
		IsDataDiff:     false,
	}
	handleMsgMhfSavePlateData(session, pkt)

	// Verify cache was invalidated
	if data := server.userBinary.GetCopy(42, 2); len(data) > 0 {
		t.Error("Expected user binary type 2 to be invalidated")
	}
	if data := server.userBinary.GetCopy(42, 3); len(data) > 0 {
		t.Error("Expected user binary type 3 to be invalidated")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}
