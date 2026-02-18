package channelserver

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
	"erupe-ce/network/mhfpacket"
)

// ============================================================================
// RENGOKU (HUNTING ROAD) INTEGRATION TESTS
// Tests for GitHub issue #85: Hunting Road skill data not saving
//
// The bug: Road skills are reset upon login. Points spent remain invested
// but skills are not equipped, forcing users to use a reset item.
//
// These tests verify the save/load round-trip integrity for rengoku data
// to determine if the server-side persistence is the root cause.
// ============================================================================

// buildRengokuTestPayload creates a realistic rengoku save data payload.
// The structure is based on the default empty response in handleMsgMhfLoadRengokuData
// and pcap analysis. Fields are annotated with known offsets.
//
// Layout (based on load handler default + save handler score extraction):
//
//	Offset 0-3:   uint32 unknown (progression flags?)
//	Offset 4-7:   uint32 unknown
//	Offset 8-9:   uint16 unknown
//	Offset 10-13: uint32 unknown
//	Offset 14-15: uint16 unknown
//	Offset 16-17: uint16 unknown
//	Offset 18-21: uint32 unknown
//	Offset 22-25: uint32 unknown (added based on pcaps)
//	Offset 26:    uint8  count1 (3 entries of uint16)
//	Offset 27-32: 3x uint16 — possibly skill slot IDs or flags
//	Offset 33-44: 3x uint32 — unknown (12 bytes)
//	Offset 45:    uint8  count2 (3 entries of uint32)
//	Offset 46-57: 3x uint32 — possibly equipped skill data
//	Offset 58:    uint8  count3 (3 entries of uint32)
//	Offset 59-70: 3x uint32 — possibly skill point allocations
//	Offset 71-74: uint32 maxStageMp (extracted by save handler)
//	Offset 75-78: uint32 maxScoreMp (extracted by save handler)
//	Offset 79-82: 4 bytes skipped (seek +4 in save handler)
//	Offset 83-86: uint32 maxStageSp (extracted by save handler)
//	Offset 87-90: uint32 maxScoreSp (extracted by save handler)
//	Offset 91+:   remaining score/progression data
func buildRengokuTestPayload(
	maxStageMp, maxScoreMp, maxStageSp, maxScoreSp uint32,
	skillSlots [3]uint16,
	equippedSkills [3]uint32,
	skillPoints [3]uint32,
) []byte {
	bf := byteframe.NewByteFrame()

	// Header region (offsets 0-25): progression flags, etc.
	bf.WriteUint32(0x00000001) // 0-3: some flag indicating data exists
	bf.WriteUint32(0)          // 4-7
	bf.WriteUint16(0)          // 8-9
	bf.WriteUint32(0)          // 10-13
	bf.WriteUint16(0)          // 14-15
	bf.WriteUint16(0)          // 16-17
	bf.WriteUint32(0)          // 18-21
	bf.WriteUint32(0)          // 22-25: extra 4 bytes from pcaps

	// Skill slots region (offsets 26-32)
	bf.WriteUint8(3)
	for _, slot := range skillSlots {
		bf.WriteUint16(slot)
	}

	// Unknown uint32 region (offsets 33-44)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)

	// Equipped skills region (offsets 45-57)
	bf.WriteUint8(3)
	for _, skill := range equippedSkills {
		bf.WriteUint32(skill)
	}

	// Skill points region (offsets 58-70)
	bf.WriteUint8(3)
	for _, pts := range skillPoints {
		bf.WriteUint32(pts)
	}

	// Score region (offsets 71-90) — extracted by save handler
	bf.WriteUint32(maxStageMp)
	bf.WriteUint32(maxScoreMp)
	bf.WriteUint32(0) // 4 bytes skipped by save handler (seek +4)
	bf.WriteUint32(maxStageSp)
	bf.WriteUint32(maxScoreSp)

	// Trailing data
	bf.WriteUint32(0)

	return bf.Data()
}

// extractAckData parses a serialized packet from the session send channel
// and returns the AckData payload. The packet format is:
// 2 bytes opcode + MsgSysAck.Build() output.
func extractAckData(t *testing.T, s *Session) []byte {
	t.Helper()
	select {
	case p := <-s.sendPackets:
		if len(p.data) < 2 {
			t.Fatal("Packet too short to contain opcode")
		}
		// Skip 2-byte opcode header, parse as MsgSysAck
		bf := byteframe.NewByteFrameFromBytes(p.data[2:])
		ack := &mhfpacket.MsgSysAck{}
		if err := ack.Parse(bf, &clientctx.ClientContext{}); err != nil {
			t.Fatalf("Failed to parse ACK packet: %v", err)
		}
		if ack.ErrorCode != 0 {
			t.Fatalf("ACK returned error code %d", ack.ErrorCode)
		}
		return ack.AckData
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for ACK packet")
		return nil
	}
}

// drainAck consumes one packet from the send channel (used after save operations).
func drainAck(t *testing.T, s *Session) {
	t.Helper()
	select {
	case <-s.sendPackets:
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for ACK packet")
	}
}

// TestRengokuData_SaveLoadRoundTrip verifies that rengoku data saved by
// handleMsgMhfSaveRengokuData is returned byte-for-byte identical by
// handleMsgMhfLoadRengokuData. This is the core test for issue #85.
func TestRengokuData_SaveLoadRoundTrip(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_test_user")
	charID := CreateTestCharacter(t, db, userID, "RengokuChar")

	server := createTestServerWithDB(t, db)

	session := createTestSessionForServerWithChar(server, charID, "RengokuChar")

	// Build a realistic payload with non-zero skill data
	payload := buildRengokuTestPayload(
		15, 18519,   // MP: 15 stages, 18519 points
		4, 381,      // SP: 4 stages, 381 points
		[3]uint16{0x0012, 0x0034, 0x0056}, // skill slot IDs
		[3]uint32{0x00110001, 0x00220002, 0x00330003}, // equipped skills
		[3]uint32{100, 200, 300}, // skill points invested
	)

	// === SAVE ===
	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      1001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	// === LOAD ===
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 1002,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	// === VERIFY BYTE-FOR-BYTE EQUALITY ===
	if !bytes.Equal(payload, loadedData) {
		t.Errorf("Round-trip mismatch: saved %d bytes, loaded %d bytes", len(payload), len(loadedData))
		// Find first differing byte for diagnostics
		minLen := len(payload)
		if len(loadedData) < minLen {
			minLen = len(loadedData)
		}
		for i := 0; i < minLen; i++ {
			if payload[i] != loadedData[i] {
				t.Errorf("First difference at offset %d: saved 0x%02X, loaded 0x%02X", i, payload[i], loadedData[i])
				break
			}
		}
	} else {
		t.Logf("Round-trip OK: %d bytes saved and loaded identically", len(payload))
	}
}

// TestRengokuData_SaveLoadRoundTrip_AcrossSessions tests that rengoku data
// persists across session boundaries (simulating logout/login).
func TestRengokuData_SaveLoadRoundTrip_AcrossSessions(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_session_user")
	charID := CreateTestCharacter(t, db, userID, "RengokuChar2")

	server := createTestServerWithDB(t, db)

	// === SESSION 1: Save data, then logout ===
	session1 := createTestSessionForServerWithChar(server, charID, "RengokuChar2")

	payload := buildRengokuTestPayload(
		80, 342295,  // MP: deep run
		38, 54634,   // SP: deep run
		[3]uint16{0x00AA, 0x00BB, 0x00CC},
		[3]uint32{0xDEAD0001, 0xBEEF0002, 0xCAFE0003},
		[3]uint32{500, 750, 1000},
	)

	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      2001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}
	handleMsgMhfSaveRengokuData(session1, savePkt)
	drainAck(t, session1)

	// Logout session 1
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// === SESSION 2: Load data in new session ===
	session2 := createTestSessionForServerWithChar(server, charID, "RengokuChar2")

	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 2002,
	}
	handleMsgMhfLoadRengokuData(session2, loadPkt)
	loadedData := extractAckData(t, session2)

	if !bytes.Equal(payload, loadedData) {
		t.Errorf("Cross-session round-trip mismatch: saved %d bytes, loaded %d bytes", len(payload), len(loadedData))
		minLen := len(payload)
		if len(loadedData) < minLen {
			minLen = len(loadedData)
		}
		for i := 0; i < minLen; i++ {
			if payload[i] != loadedData[i] {
				t.Errorf("First difference at offset %d: saved 0x%02X, loaded 0x%02X", i, payload[i], loadedData[i])
				break
			}
		}
	} else {
		t.Logf("Cross-session round-trip OK: %d bytes persisted correctly", len(payload))
	}

	logoutPlayer(session2)
}

// TestRengokuData_ScoreExtraction verifies that the save handler correctly
// extracts stage/score metadata into the rengoku_score table.
func TestRengokuData_ScoreExtraction(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_score_user")
	charID := CreateTestCharacter(t, db, userID, "ScoreChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "ScoreChar")

	maxStageMp := uint32(15)
	maxScoreMp := uint32(18519)
	maxStageSp := uint32(4)
	maxScoreSp := uint32(381)

	payload := buildRengokuTestPayload(
		maxStageMp, maxScoreMp, maxStageSp, maxScoreSp,
		[3]uint16{}, [3]uint32{}, [3]uint32{},
	)

	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      3001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	// Verify rengoku_score table
	var gotStageMp, gotScoreMp, gotStageSp, gotScoreSp uint32
	err := db.QueryRow(
		"SELECT max_stages_mp, max_points_mp, max_stages_sp, max_points_sp FROM rengoku_score WHERE character_id=$1",
		charID,
	).Scan(&gotStageMp, &gotScoreMp, &gotStageSp, &gotScoreSp)
	if err != nil {
		t.Fatalf("Failed to query rengoku_score: %v", err)
	}

	if gotStageMp != maxStageMp {
		t.Errorf("max_stages_mp: got %d, want %d", gotStageMp, maxStageMp)
	}
	if gotScoreMp != maxScoreMp {
		t.Errorf("max_points_mp: got %d, want %d", gotScoreMp, maxScoreMp)
	}
	if gotStageSp != maxStageSp {
		t.Errorf("max_stages_sp: got %d, want %d", gotStageSp, maxStageSp)
	}
	if gotScoreSp != maxScoreSp {
		t.Errorf("max_points_sp: got %d, want %d", gotScoreSp, maxScoreSp)
	}

	t.Logf("Score extraction OK: MP(%d stages, %d pts) SP(%d stages, %d pts)",
		gotStageMp, gotScoreMp, gotStageSp, gotScoreSp)
}

// TestRengokuData_SkillRegionPreserved verifies that the "skill" portion of
// the rengoku blob (offsets 26-70) survives the round-trip intact.
// This directly targets issue #85: skills reset but points remain.
func TestRengokuData_SkillRegionPreserved(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_skill_user")
	charID := CreateTestCharacter(t, db, userID, "SkillChar")

	server := createTestServerWithDB(t, db)

	// === SESSION 1: Save with non-zero skill data ===
	session1 := createTestSessionForServerWithChar(server, charID, "SkillChar")

	skillSlots := [3]uint16{0x1234, 0x5678, 0x9ABC}
	equippedSkills := [3]uint32{0xAAAA1111, 0xBBBB2222, 0xCCCC3333}
	skillPoints := [3]uint32{999, 888, 777}

	payload := buildRengokuTestPayload(
		10, 5000, 5, 1000,
		skillSlots, equippedSkills, skillPoints,
	)

	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      4001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}
	handleMsgMhfSaveRengokuData(session1, savePkt)
	drainAck(t, session1)
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// === SESSION 2: Load and verify skill region ===
	session2 := createTestSessionForServerWithChar(server, charID, "SkillChar")

	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 4002,
	}
	handleMsgMhfLoadRengokuData(session2, loadPkt)
	loadedData := extractAckData(t, session2)

	// Parse skill region from loaded data
	bf := byteframe.NewByteFrameFromBytes(loadedData)
	_, _ = bf.Seek(26, 0) // Skip to skill slots region

	count1 := bf.ReadUint8()
	if count1 != 3 {
		t.Fatalf("Skill slot count: got %d, want 3", count1)
	}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint16()
		if got != skillSlots[i] {
			t.Errorf("Skill slot %d: got 0x%04X, want 0x%04X", i, got, skillSlots[i])
		}
	}

	// Skip 12 bytes of unknown uint32s
	_, _ = bf.Seek(12, 1)

	count2 := bf.ReadUint8()
	if count2 != 3 {
		t.Fatalf("Equipped skill count: got %d, want 3", count2)
	}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint32()
		if got != equippedSkills[i] {
			t.Errorf("Equipped skill %d: got 0x%08X, want 0x%08X", i, got, equippedSkills[i])
		}
	}

	count3 := bf.ReadUint8()
	if count3 != 3 {
		t.Fatalf("Skill points count: got %d, want 3", count3)
	}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint32()
		if got != skillPoints[i] {
			t.Errorf("Skill points %d: got %d, want %d", i, got, skillPoints[i])
		}
	}

	t.Log("Skill region preserved across sessions")
	logoutPlayer(session2)
}

// TestRengokuData_OverwritePreservesNewData verifies that saving new rengoku
// data overwrites the old data completely (no stale data leaking through).
func TestRengokuData_OverwritePreservesNewData(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_overwrite_user")
	charID := CreateTestCharacter(t, db, userID, "OverwriteChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "OverwriteChar")

	// First save: skills equipped
	payload1 := buildRengokuTestPayload(
		10, 5000, 5, 1000,
		[3]uint16{0x1111, 0x2222, 0x3333},
		[3]uint32{0xAAAAAAAA, 0xBBBBBBBB, 0xCCCCCCCC},
		[3]uint32{100, 200, 300},
	)
	savePkt1 := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      5001,
		DataSize:       uint32(len(payload1)),
		RawDataPayload: payload1,
	}
	handleMsgMhfSaveRengokuData(session, savePkt1)
	drainAck(t, session)

	// Second save: different skills (simulating skill reset + re-equip)
	payload2 := buildRengokuTestPayload(
		12, 7000, 6, 2000,
		[3]uint16{0x4444, 0x5555, 0x6666},
		[3]uint32{0xDDDDDDDD, 0xEEEEEEEE, 0xFFFFFFFF},
		[3]uint32{400, 500, 600},
	)
	savePkt2 := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      5002,
		DataSize:       uint32(len(payload2)),
		RawDataPayload: payload2,
	}
	handleMsgMhfSaveRengokuData(session, savePkt2)
	drainAck(t, session)

	// Load and verify we get payload2, not payload1
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 5003,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	if !bytes.Equal(payload2, loadedData) {
		t.Error("Overwrite failed: loaded data does not match second save")
		if bytes.Equal(payload1, loadedData) {
			t.Error("Loaded data matches FIRST save — overwrite did not take effect")
		}
	} else {
		t.Log("Overwrite OK: second save correctly replaced first")
	}
}

// TestRengokuData_DefaultResponseStructure verifies the default (empty)
// response matches the expected client structure when no data exists.
func TestRengokuData_DefaultResponseStructure(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_default_user")
	charID := CreateTestCharacter(t, db, userID, "DefaultChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "DefaultChar")

	// Load without any prior save
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 6001,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	data := extractAckData(t, session)

	// Expected size: 4+4+2+4+2+2+4+4 + 1+6 + 12 + 1+12 + 1+12 + 24 = 95 bytes
	// Manually compute from the handler:
	expected := byteframe.NewByteFrame()
	expected.WriteUint32(0) // 4
	expected.WriteUint32(0) // 4
	expected.WriteUint16(0) // 2
	expected.WriteUint32(0) // 4
	expected.WriteUint16(0) // 2
	expected.WriteUint16(0) // 2
	expected.WriteUint32(0) // 4
	expected.WriteUint32(0) // 4  (pcap extra)

	expected.WriteUint8(3)  // count
	expected.WriteUint16(0) // 3x uint16
	expected.WriteUint16(0)
	expected.WriteUint16(0)

	expected.WriteUint32(0) // 3x uint32
	expected.WriteUint32(0)
	expected.WriteUint32(0)

	expected.WriteUint8(3)  // count
	expected.WriteUint32(0) // 3x uint32
	expected.WriteUint32(0)
	expected.WriteUint32(0)

	expected.WriteUint8(3)  // count
	expected.WriteUint32(0) // 3x uint32
	expected.WriteUint32(0)
	expected.WriteUint32(0)

	expected.WriteUint32(0) // 6x uint32
	expected.WriteUint32(0)
	expected.WriteUint32(0)
	expected.WriteUint32(0)
	expected.WriteUint32(0)
	expected.WriteUint32(0)

	expectedData := expected.Data()

	if !bytes.Equal(data, expectedData) {
		t.Errorf("Default response mismatch: got %d bytes, want %d bytes", len(data), len(expectedData))
		t.Errorf("Got:    %X", data)
		t.Errorf("Expect: %X", expectedData)
	} else {
		t.Logf("Default response OK: %d bytes", len(data))
	}
}

// TestRengokuData_SaveOnDBError verifies that save handler sends ACK even on DB failure.
// Note: requires a test DB because the handler accesses server.db directly without
// nil checks. This test uses a valid DB connection then drops the table to simulate error.
func TestRengokuData_SaveOnDBError(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_err_user")
	charID := CreateTestCharacter(t, db, userID, "ErrChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "ErrChar")

	// Drop the rengoku_score table to trigger error in score extraction
	_, _ = db.Exec("DROP TABLE IF EXISTS rengoku_score")

	payload := make([]byte, 100)
	binary.BigEndian.PutUint32(payload[71:75], 10) // maxStageMp

	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      7001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}

	// Should not panic, should send ACK even on score table error
	handleMsgMhfSaveRengokuData(session, savePkt)

	select {
	case <-session.sendPackets:
		t.Log("ACK sent despite rengoku_score table error")
	case <-time.After(2 * time.Second):
		t.Error("No ACK sent on DB error — client would hang")
	}
}

// TestRengokuData_LoadOnDBError verifies that load handler sends default data on DB failure.
func TestRengokuData_LoadOnDBError(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	// Use a charID that doesn't exist to trigger "no rows" error
	session := createTestSessionForServerWithChar(server, 999999, "GhostChar")

	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 8001,
	}

	// Should not panic, should send default response
	handleMsgMhfLoadRengokuData(session, loadPkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Empty response on DB error")
		} else {
			t.Log("Default response sent on missing character")
		}
	case <-time.After(2 * time.Second):
		t.Error("No response sent on DB error — client would hang")
	}
}

// TestRengokuData_MultipleSavesSameSession verifies that multiple saves in
// the same session always persist the latest data.
func TestRengokuData_MultipleSavesSameSession(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_multi_user")
	charID := CreateTestCharacter(t, db, userID, "MultiChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "MultiChar")

	// Simulate Road progression: save after each floor
	for floor := uint32(1); floor <= 5; floor++ {
		payload := buildRengokuTestPayload(
			floor, floor*1000, 0, 0,
			[3]uint16{uint16(floor), uint16(floor * 10), uint16(floor * 100)},
			[3]uint32{floor, floor * 2, floor * 3},
			[3]uint32{floor * 100, floor * 200, floor * 300},
		)

		savePkt := &mhfpacket.MsgMhfSaveRengokuData{
			AckHandle:      9000 + floor,
			DataSize:       uint32(len(payload)),
			RawDataPayload: payload,
		}
		handleMsgMhfSaveRengokuData(session, savePkt)
		drainAck(t, session)
	}

	// Load should return the last save (floor 5)
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 9999,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	// Build expected final payload
	expectedPayload := buildRengokuTestPayload(
		5, 5000, 0, 0,
		[3]uint16{5, 50, 500},
		[3]uint32{5, 10, 15},
		[3]uint32{500, 1000, 1500},
	)

	if !bytes.Equal(expectedPayload, loadedData) {
		t.Error("After 5 saves, loaded data does not match the final save")
	} else {
		t.Log("Multiple saves OK: final state persisted correctly")
	}

	// Verify rengoku_score has the latest scores
	var gotStage, gotScore uint32
	err := db.QueryRow(
		"SELECT max_stages_mp, max_points_mp FROM rengoku_score WHERE character_id=$1",
		charID,
	).Scan(&gotStage, &gotScore)
	if err != nil {
		t.Fatalf("Failed to query rengoku_score: %v", err)
	}
	if gotStage != 5 || gotScore != 5000 {
		t.Errorf("Score not updated: got stage=%d score=%d, want stage=5 score=5000", gotStage, gotScore)
	}
}

// ============================================================================
// PROTECTION LOGIC UNIT TESTS (Issue #85 fix)
// Tests for rengokuSkillsZeroed, rengokuHasPoints, rengokuMergeSkills,
// and the race condition detection in handleMsgMhfSaveRengokuData.
// ============================================================================

// TestRengokuSkillsZeroed verifies the zeroed-skill detection function.
func TestRengokuSkillsZeroed(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect bool
	}{
		{"nil data", nil, true},
		{"too short", make([]byte, 0x30), true},
		{"all zeroed", make([]byte, 0x47), true},
		{"skill slot nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x1B] = 0x12
			return d
		}(), false},
		{"equipped skill nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x2E] = 0x01
			return d
		}(), false},
		{"last skill slot byte nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x20] = 0xFF
			return d
		}(), false},
		{"last equipped byte nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x39] = 0xFF
			return d
		}(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rengokuSkillsZeroed(tt.data)
			if got != tt.expect {
				t.Errorf("rengokuSkillsZeroed() = %v, want %v", got, tt.expect)
			}
		})
	}
}

// TestRengokuHasPoints verifies the point-allocation detection function.
func TestRengokuHasPoints(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect bool
	}{
		{"nil data", nil, false},
		{"too short", make([]byte, 0x40), false},
		{"all zeroed", make([]byte, 0x47), false},
		{"first point nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x3B] = 0x01
			return d
		}(), true},
		{"last point nonzero", func() []byte {
			d := make([]byte, 0x47)
			d[0x46] = 0x01
			return d
		}(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rengokuHasPoints(tt.data)
			if got != tt.expect {
				t.Errorf("rengokuHasPoints() = %v, want %v", got, tt.expect)
			}
		})
	}
}

// TestRengokuMergeSkills verifies skill data is copied from src to dst.
func TestRengokuMergeSkills(t *testing.T) {
	dst := make([]byte, 0x47)
	src := make([]byte, 0x47)

	// Fill src skill regions with identifiable data
	for i := 0x1B; i <= 0x20; i++ {
		src[i] = byte(i)
	}
	for i := 0x2E; i <= 0x39; i++ {
		src[i] = byte(i)
	}
	// Put some data in dst points region that should NOT be touched
	dst[0x3B] = 0xFF

	rengokuMergeSkills(dst, src)

	// Verify skill slots were copied
	for i := 0x1B; i <= 0x20; i++ {
		if dst[i] != byte(i) {
			t.Errorf("offset 0x%02X: got 0x%02X, want 0x%02X", i, dst[i], byte(i))
		}
	}
	// Verify equipped skills were copied
	for i := 0x2E; i <= 0x39; i++ {
		if dst[i] != byte(i) {
			t.Errorf("offset 0x%02X: got 0x%02X, want 0x%02X", i, dst[i], byte(i))
		}
	}
	// Verify points region was NOT touched
	if dst[0x3B] != 0xFF {
		t.Errorf("Points region modified: got 0x%02X, want 0xFF", dst[0x3B])
	}
}

// TestRengokuData_RaceConditionMerge simulates the Sky Corridor race condition
// (issue #85): client sends a save with zeroed skills but nonzero points.
// The server should merge existing skill data into the save.
func TestRengokuData_RaceConditionMerge(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_race_user")
	charID := CreateTestCharacter(t, db, userID, "RaceChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "RaceChar")

	// Step 1: Save valid data with skills equipped
	validPayload := buildRengokuTestPayload(
		10, 5000, 5, 1000,
		[3]uint16{0x0012, 0x0034, 0x0056},
		[3]uint32{0x00110001, 0x00220002, 0x00330003},
		[3]uint32{100, 200, 300},
	)
	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      11001,
		DataSize:       uint32(len(validPayload)),
		RawDataPayload: validPayload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	// Step 2: Simulate race condition — zeroed skills, nonzero points
	racedPayload := buildRengokuTestPayload(
		12, 7000, 6, 2000,
		[3]uint16{0, 0, 0},       // zeroed skill slots (race condition)
		[3]uint32{0, 0, 0},       // zeroed equipped skills (race condition)
		[3]uint32{100, 200, 300}, // points still present
	)
	racePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      11002,
		DataSize:       uint32(len(racedPayload)),
		RawDataPayload: racedPayload,
	}
	handleMsgMhfSaveRengokuData(session, racePkt)
	drainAck(t, session)

	// Step 3: Load and verify skills were preserved from step 1
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 11003,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	// Parse skill region
	bf := byteframe.NewByteFrameFromBytes(loadedData)
	_, _ = bf.Seek(26, 0) // offset of count1

	count1 := bf.ReadUint8()
	if count1 != 3 {
		t.Fatalf("Skill slot count: got %d, want 3", count1)
	}
	expectedSlots := [3]uint16{0x0012, 0x0034, 0x0056}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint16()
		if got != expectedSlots[i] {
			t.Errorf("Skill slot %d: got 0x%04X, want 0x%04X (skill was NOT preserved)", i, got, expectedSlots[i])
		}
	}

	_, _ = bf.Seek(12, 1) // skip unknown u32 triple

	count2 := bf.ReadUint8()
	if count2 != 3 {
		t.Fatalf("Equipped skill count: got %d, want 3", count2)
	}
	expectedEquipped := [3]uint32{0x00110001, 0x00220002, 0x00330003}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint32()
		if got != expectedEquipped[i] {
			t.Errorf("Equipped skill %d: got 0x%08X, want 0x%08X (skill was NOT preserved)", i, got, expectedEquipped[i])
		}
	}

	// Points should reflect the raced save (updated to step 2 values)
	count3 := bf.ReadUint8()
	if count3 != 3 {
		t.Fatalf("Skill points count: got %d, want 3", count3)
	}
	expectedPoints := [3]uint32{100, 200, 300}
	for i := 0; i < 3; i++ {
		got := bf.ReadUint32()
		if got != expectedPoints[i] {
			t.Errorf("Skill points %d: got %d, want %d", i, got, expectedPoints[i])
		}
	}

	// Scores should be from the raced save (step 2 values, not step 1)
	_, _ = bf.Seek(71, 0)
	gotStageMp := bf.ReadUint32()
	gotScoreMp := bf.ReadUint32()
	if gotStageMp != 12 || gotScoreMp != 7000 {
		t.Errorf("Scores not updated from raced save: stageMp=%d scoreMp=%d, want 12/7000", gotStageMp, gotScoreMp)
	}

	t.Log("Race condition merge OK: skills preserved, scores and points updated")
}

// TestRengokuData_EmptySentinelRejection verifies that a save with sentinel=0
// does not overwrite valid existing data.
func TestRengokuData_EmptySentinelRejection(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_sentinel_user")
	charID := CreateTestCharacter(t, db, userID, "SentinelChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "SentinelChar")

	// Step 1: Save valid data (sentinel != 0)
	validPayload := buildRengokuTestPayload(
		10, 5000, 5, 1000,
		[3]uint16{0x0012, 0x0034, 0x0056},
		[3]uint32{0x00110001, 0x00220002, 0x00330003},
		[3]uint32{100, 200, 300},
	)
	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      12001,
		DataSize:       uint32(len(validPayload)),
		RawDataPayload: validPayload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	// Step 2: Try to save with sentinel=0 (empty data)
	emptyPayload := make([]byte, 95)
	// sentinel at offset 0-3 is already 0
	emptyPkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      12002,
		DataSize:       uint32(len(emptyPayload)),
		RawDataPayload: emptyPayload,
	}
	handleMsgMhfSaveRengokuData(session, emptyPkt)
	drainAck(t, session)

	// Step 3: Load and verify original data was preserved
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 12003,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	if !bytes.Equal(validPayload, loadedData) {
		t.Error("Empty sentinel save overwrote valid data!")
	} else {
		t.Log("Empty sentinel rejection OK: valid data preserved")
	}
}

// TestRengokuData_EmptySentinelAllowedWhenNoExisting verifies that a save
// with sentinel=0 is allowed when no valid data exists yet.
func TestRengokuData_EmptySentinelAllowedWhenNoExisting(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_sentinel_ok_user")
	charID := CreateTestCharacter(t, db, userID, "SentinelOKChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "SentinelOKChar")

	// Save with sentinel=0 when no existing data
	emptyPayload := make([]byte, 95)
	binary.BigEndian.PutUint32(emptyPayload[71:75], 0) // maxStageMp = 0
	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      13001,
		DataSize:       uint32(len(emptyPayload)),
		RawDataPayload: emptyPayload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	// Load and verify it was saved
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 13002,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	if !bytes.Equal(emptyPayload, loadedData) {
		t.Error("Empty sentinel save was rejected when no existing data")
	} else {
		t.Log("Empty sentinel allowed when no existing data")
	}
}

// TestRengokuData_NoMergeWhenSkillsPresent verifies that the merge logic
// does NOT activate when the incoming save has valid skill data.
func TestRengokuData_NoMergeWhenSkillsPresent(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_nomerge_user")
	charID := CreateTestCharacter(t, db, userID, "NoMergeChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "NoMergeChar")

	// Step 1: Save with skills A
	payload1 := buildRengokuTestPayload(
		10, 5000, 5, 1000,
		[3]uint16{0x0012, 0x0034, 0x0056},
		[3]uint32{0x00110001, 0x00220002, 0x00330003},
		[3]uint32{100, 200, 300},
	)
	savePkt1 := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      14001,
		DataSize:       uint32(len(payload1)),
		RawDataPayload: payload1,
	}
	handleMsgMhfSaveRengokuData(session, savePkt1)
	drainAck(t, session)

	// Step 2: Save with different skills B (not zeroed — should NOT merge)
	payload2 := buildRengokuTestPayload(
		12, 7000, 6, 2000,
		[3]uint16{0xAAAA, 0xBBBB, 0xCCCC},
		[3]uint32{0xDDDD0001, 0xEEEE0002, 0xFFFF0003},
		[3]uint32{400, 500, 600},
	)
	savePkt2 := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      14002,
		DataSize:       uint32(len(payload2)),
		RawDataPayload: payload2,
	}
	handleMsgMhfSaveRengokuData(session, savePkt2)
	drainAck(t, session)

	// Step 3: Load and verify we get payload2, not a merge
	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 14003,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	if !bytes.Equal(payload2, loadedData) {
		t.Error("Valid skill save was incorrectly merged with existing data")
	} else {
		t.Log("No merge when skills are present: correct behavior")
	}
}

// TestRengokuData_LargePayload tests round-trip with a larger-than-default payload.
// Some client versions may send more data than the default structure.
func TestRengokuData_LargePayload(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rengoku_large_user")
	charID := CreateTestCharacter(t, db, userID, "LargeChar")

	server := createTestServerWithDB(t, db)
	session := createTestSessionForServerWithChar(server, charID, "LargeChar")

	// Build a payload larger than the default structure
	// Real clients may send 200+ bytes with additional fields
	payload := make([]byte, 256)
	// Fill with identifiable pattern
	for i := range payload {
		payload[i] = byte(i)
	}
	// Ensure valid score region at offsets 71-90
	binary.BigEndian.PutUint32(payload[71:75], 20)    // maxStageMp
	binary.BigEndian.PutUint32(payload[75:79], 30000)  // maxScoreMp
	binary.BigEndian.PutUint32(payload[83:87], 10)     // maxStageSp
	binary.BigEndian.PutUint32(payload[87:91], 15000)  // maxScoreSp

	savePkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      10001,
		DataSize:       uint32(len(payload)),
		RawDataPayload: payload,
	}
	handleMsgMhfSaveRengokuData(session, savePkt)
	drainAck(t, session)

	loadPkt := &mhfpacket.MsgMhfLoadRengokuData{
		AckHandle: 10002,
	}
	handleMsgMhfLoadRengokuData(session, loadPkt)
	loadedData := extractAckData(t, session)

	if !bytes.Equal(payload, loadedData) {
		t.Errorf("Large payload round-trip failed: saved %d bytes, loaded %d bytes", len(payload), len(loadedData))
	} else {
		t.Logf("Large payload round-trip OK: %d bytes", len(payload))
	}
}
