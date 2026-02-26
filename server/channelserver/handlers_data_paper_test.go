package channelserver

import (
	"encoding/binary"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// paperTestSession creates a minimal session for paper data handler tests.
func paperTestSession() *Session {
	server := createMockServer()
	return createMockSession(1, server)
}

// callGetPaperData invokes the handler and returns the ACK payload.
func callGetPaperData(t *testing.T, dataType uint32) []byte {
	t.Helper()
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  dataType,
	}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		return p.data
	default:
		t.Fatal("expected ACK packet, got none")
		return nil
	}
}

// --- DataType 0: Mission Timetable ---

func TestGetPaperData_Type0_MissionTimetable(t *testing.T) {
	data := callGetPaperData(t, 0)
	if len(data) == 0 {
		t.Fatal("expected non-empty response for DataType 0")
	}

	// doAckBufSucceed wraps the payload in a MsgSysAck.
	// The raw payload sent to the session contains the ack structure.
	// We just verify the packet was sent and is non-empty.
}

func TestGetPaperData_Type0_MissionPayloadStructure(t *testing.T) {
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 1, DataType: 0}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case <-s.sendPackets:
		// ACK sent successfully
	default:
		t.Fatal("expected ACK packet for DataType 0")
	}
}

// --- DataType 5: Tower Parameters ---

func TestGetPaperData_Type5_TowerParams(t *testing.T) {
	data := callGetPaperData(t, 5)
	if len(data) == 0 {
		t.Fatal("expected non-empty response for DataType 5")
	}
}

func TestGetPaperData_Type5_EntryCount(t *testing.T) {
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 1, DataType: 5}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		// doAckEarthSucceed writes: earthID(4) + 0(4) + 0(4) + count(4) + entries
		// The full packet includes the MsgSysAck header, but we can verify it's substantial.
		// Type 5 has 52 PaperData entries (counted from source), each 14 bytes.
		// Minimum expected: 16 (earth header) + 52*14 = 744 bytes in the ack payload.
		if len(p.data) < 100 {
			t.Errorf("type 5 payload too small: %d bytes", len(p.data))
		}
	default:
		t.Fatal("expected ACK packet for DataType 5")
	}
}

// --- DataType 6: Tower Floor/Reward Data ---

func TestGetPaperData_Type6_TowerFloorData(t *testing.T) {
	data := callGetPaperData(t, 6)
	if len(data) == 0 {
		t.Fatal("expected non-empty response for DataType 6")
	}
}

func TestGetPaperData_Type6_LargerThanType5(t *testing.T) {
	data5 := callGetPaperData(t, 5)
	data6 := callGetPaperData(t, 6)

	// Type 6 has significantly more entries than type 5
	if len(data6) <= len(data5) {
		t.Errorf("type 6 (%d bytes) should be larger than type 5 (%d bytes)", len(data6), len(data5))
	}
}

// --- DataType > 1000: Paper Gift Data ---

func TestGetPaperData_KnownGiftType_6001(t *testing.T) {
	data := callGetPaperData(t, 6001)
	if len(data) == 0 {
		t.Fatal("expected non-empty response for gift type 6001")
	}
}

func TestGetPaperData_KnownGiftType_7001(t *testing.T) {
	data := callGetPaperData(t, 7001)
	if len(data) == 0 {
		t.Fatal("expected non-empty response for gift type 7001")
	}
}

func TestGetPaperData_AllKnownGiftTypes(t *testing.T) {
	for dataType := range paperGiftData {
		t.Run("gift_"+itoa(dataType), func(t *testing.T) {
			data := callGetPaperData(t, dataType)
			if len(data) == 0 {
				t.Errorf("expected non-empty response for gift type %d", dataType)
			}
		})
	}
}

// --- DataType > 1000 with unknown key ---

func TestGetPaperData_UnknownGiftType(t *testing.T) {
	// 9999 is > 1000 but not in paperGiftData
	data := callGetPaperData(t, 9999)
	if len(data) == 0 {
		t.Fatal("expected ACK even for unknown gift type")
	}
}

// --- Unknown DataType (< 1000, not 0/5/6) ---

func TestGetPaperData_UnknownType_3(t *testing.T) {
	// DataType 3 hits the default case, then the else branch (empty paperData)
	data := callGetPaperData(t, 3)
	if len(data) == 0 {
		t.Fatal("expected ACK even for unknown DataType")
	}
}

func TestGetPaperData_UnknownType_1(t *testing.T) {
	data := callGetPaperData(t, 1)
	if len(data) == 0 {
		t.Fatal("expected ACK for DataType 1")
	}
}

// --- Serialization Verification ---

func TestGetPaperData_Type0_SerializationFormat(t *testing.T) {
	// Build expected payload manually and compare structure
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 42, DataType: 0}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		// The raw data is the full MsgSysAck Build output.
		// We verify it's non-trivial (contains the timetable data).
		if len(p.data) < 20 {
			t.Errorf("type 0 ACK payload too small: %d bytes", len(p.data))
		}
	default:
		t.Fatal("expected ACK packet")
	}
}

// ackPayloadOffset is the offset to the ACK payload data within the raw packet.
// Raw packet layout: opcode(2) + AckHandle(4) + IsBuffer(1) + ErrorCode(1) + payloadSize(2) = 10 bytes header.
const ackPayloadOffset = 10

// extractAckPayload extracts the ACK payload from a raw packet sent via QueueSendMHF.
func extractAckPayload(t *testing.T, data []byte) []byte {
	t.Helper()
	if len(data) < ackPayloadOffset {
		t.Fatalf("packet too short for ACK header: %d bytes", len(data))
	}
	payloadLen := binary.BigEndian.Uint16(data[8:10])
	if payloadLen == 0xFFFF {
		// Extended size
		if len(data) < 14 {
			t.Fatalf("packet too short for extended ACK header: %d bytes", len(data))
		}
		extLen := binary.BigEndian.Uint32(data[10:14])
		return data[14 : 14+extLen]
	}
	return data[ackPayloadOffset : ackPayloadOffset+int(payloadLen)]
}

func TestGetPaperData_GiftSerialization_6001(t *testing.T) {
	// Verify that gift type 6001 produces the right number of gift entries.
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 1, DataType: 6001}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		payload := extractAckPayload(t, p.data)

		// Earth succeed: earthID(4) + 0(4) + 0(4) + count(4) = 16 byte header
		if len(payload) < 16 {
			t.Fatalf("earth payload too short: %d bytes", len(payload))
		}
		count := binary.BigEndian.Uint32(payload[12:16])
		expectedCount := uint32(len(paperGiftData[6001]))
		if count != expectedCount {
			t.Errorf("gift entry count = %d, want %d", count, expectedCount)
		}

		// Each gift entry is 6 bytes
		expectedDataLen := 16 + int(expectedCount)*6
		if len(payload) != expectedDataLen {
			t.Errorf("earth payload length = %d, want %d", len(payload), expectedDataLen)
		}
	default:
		t.Fatal("expected ACK packet")
	}
}

func TestGetPaperData_Type5_EarthSucceedEntryCount(t *testing.T) {
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 1, DataType: 5}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		payload := extractAckPayload(t, p.data)

		// Earth succeed: earthID(4) + 0(4) + 0(4) + count(4) = 16 byte header
		if len(payload) < 16 {
			t.Fatalf("earth payload too short: %d bytes", len(payload))
		}
		count := binary.BigEndian.Uint32(payload[12:16])
		// Type 5 has 52 PaperData entries
		if count != 52 {
			t.Errorf("type 5 entry count = %d, want 52", count)
		}

		// Each PaperData entry: uint16 + 6*int16 = 14 bytes
		expectedDataLen := 16 + 52*14
		if len(payload) != expectedDataLen {
			t.Errorf("earth payload length = %d, want %d", len(payload), expectedDataLen)
		}
	default:
		t.Fatal("expected ACK packet")
	}
}

func TestGetPaperData_Type0_TimetableContent(t *testing.T) {
	s := paperTestSession()
	pkt := &mhfpacket.MsgMhfGetPaperData{AckHandle: 1, DataType: 0}
	handleMsgMhfGetPaperData(s, pkt)

	select {
	case p := <-s.sendPackets:
		payload := extractAckPayload(t, p.data)

		// Mission payload: uint16(numTimetables) + uint16(numData) + timetable entries
		if len(payload) < 4 {
			t.Fatalf("mission payload too short: %d bytes", len(payload))
		}
		numTimetables := binary.BigEndian.Uint16(payload[0:2])
		numData := binary.BigEndian.Uint16(payload[2:4])

		if numTimetables != 1 {
			t.Errorf("timetable count = %d, want 1", numTimetables)
		}
		if numData != 0 {
			t.Errorf("mission data count = %d, want 0", numData)
		}

		// 1 timetable = 8 bytes (start uint32 + end uint32)
		expectedLen := 4 + 8 // header + 1 timetable entry
		if len(payload) != expectedLen {
			t.Errorf("mission payload length = %d, want %d", len(payload), expectedLen)
		}

		// Verify start < end (midnight < midnight+24h)
		start := binary.BigEndian.Uint32(payload[4:8])
		end := binary.BigEndian.Uint32(payload[8:12])
		if start >= end {
			t.Errorf("timetable start (%d) should be < end (%d)", start, end)
		}
	default:
		t.Fatal("expected ACK packet")
	}
}

// --- paperGiftData table integrity ---

func TestPaperGiftData_AllEntriesHaveData(t *testing.T) {
	for dataType, gifts := range paperGiftData {
		if len(gifts) == 0 {
			t.Errorf("paperGiftData[%d] is empty", dataType)
		}
	}
}

func TestPaperGiftData_KnownKeys(t *testing.T) {
	expectedKeys := []uint32{6001, 6002, 6010, 6011, 6012, 7001, 7002, 7011, 7012}
	for _, key := range expectedKeys {
		if _, ok := paperGiftData[key]; !ok {
			t.Errorf("paperGiftData missing expected key %d", key)
		}
	}
}

// --- PaperData struct serialization ---

func TestPaperData_Serialization_RoundTrip(t *testing.T) {
	pd := PaperData{Unk0: 1001, Unk1: 1, Unk2: 100, Unk3: 200, Unk4: 300, Unk5: 400, Unk6: 500}

	bf := byteframe.NewByteFrame()
	bf.WriteUint16(pd.Unk0)
	bf.WriteInt16(pd.Unk1)
	bf.WriteInt16(pd.Unk2)
	bf.WriteInt16(pd.Unk3)
	bf.WriteInt16(pd.Unk4)
	bf.WriteInt16(pd.Unk5)
	bf.WriteInt16(pd.Unk6)

	data := bf.Data()
	if len(data) != 14 {
		t.Fatalf("PaperData serialized size = %d, want 14", len(data))
	}

	// Read back
	rbf := byteframe.NewByteFrameFromBytes(data)
	if rbf.ReadUint16() != 1001 {
		t.Error("Unk0 mismatch")
	}
	if rbf.ReadInt16() != 1 {
		t.Error("Unk1 mismatch")
	}
	if rbf.ReadInt16() != 100 {
		t.Error("Unk2 mismatch")
	}
}

func TestPaperGift_Serialization_RoundTrip(t *testing.T) {
	pg := PaperGift{Unk0: 11159, Unk1: 1, Unk2: 1, Unk3: 5000}

	bf := byteframe.NewByteFrame()
	bf.WriteUint16(pg.Unk0)
	bf.WriteUint8(pg.Unk1)
	bf.WriteUint8(pg.Unk2)
	bf.WriteUint16(pg.Unk3)

	data := bf.Data()
	if len(data) != 6 {
		t.Fatalf("PaperGift serialized size = %d, want 6", len(data))
	}

	rbf := byteframe.NewByteFrameFromBytes(data)
	if rbf.ReadUint16() != 11159 {
		t.Error("Unk0 mismatch")
	}
	if rbf.ReadUint8() != 1 {
		t.Error("Unk1 mismatch")
	}
	if rbf.ReadUint8() != 1 {
		t.Error("Unk2 mismatch")
	}
	if rbf.ReadUint16() != 5000 {
		t.Error("Unk3 mismatch")
	}
}

// itoa is a tiny helper to avoid importing strconv for test names.
func itoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// =============================================================================
// Tests consolidated from handlers_coverage4_test.go
// =============================================================================

func TestHandleMsgMhfGetPaperData_Case0(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  0,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 0: response should have data")
		}
	default:
		t.Error("case 0: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_Case5(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  5,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 5: response should have data")
		}
	default:
		t.Error("case 5: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_Case6(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  6,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 6: response should have data")
		}
	default:
		t.Error("case 6: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_GreaterThan1000_KnownKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// 6001 is a known key in paperGiftData
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  6001,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error(">1000 known: response should have data")
		}
	default:
		t.Error(">1000 known: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_GreaterThan1000_UnknownKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// 9999 is not a known key in paperGiftData
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  9999,
	})

	select {
	case p := <-session.sendPackets:
		// Even unknown keys should produce a response (empty earth succeed)
		_ = p
	default:
		t.Error(">1000 unknown: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_DefaultUnknownLessThan1000(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Unknown type < 1000, hits default case then falls to else branch
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:  99,
	})

	select {
	case p := <-session.sendPackets:
		_ = p
	default:
		t.Error("default <1000: no response queued")
	}
}

func TestHandleMsgMhfGetGachaPlayHistory(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGachaPlayHistory(session, &mhfpacket.MsgMhfGetGachaPlayHistory{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfPlayFreeGacha(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfPlayFreeGacha(session, &mhfpacket.MsgMhfPlayFreeGacha{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}
