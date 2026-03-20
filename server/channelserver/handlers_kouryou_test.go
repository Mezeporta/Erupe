package channelserver

import (
	"encoding/binary"
	"errors"
	"testing"

	"erupe-ce/network/mhfpacket"
)

// parseAckBufData extracts AckData from a serialized MsgSysAck buffer response.
// Wire format: opcode(2) + ackHandle(4) + isBuffer(1) + errorCode(1) + dataLen(2) + data(N)
func parseAckBufData(t *testing.T, raw []byte) (ackHandle uint32, errorCode uint8, ackData []byte) {
	t.Helper()
	if len(raw) < 10 {
		t.Fatalf("raw packet too short: %d bytes", len(raw))
	}
	ackHandle = binary.BigEndian.Uint32(raw[2:6])
	isBuffer := raw[6]
	errorCode = raw[7]
	if isBuffer == 0 {
		t.Fatal("Expected buffer response, got simple ack")
	}
	dataLen := binary.BigEndian.Uint16(raw[8:10])
	if int(dataLen) > len(raw)-10 {
		t.Fatalf("data len %d exceeds remaining bytes %d", dataLen, len(raw)-10)
	}
	ackData = raw[10 : 10+dataLen]
	return
}

func TestHandleMsgMhfGetKouryouPoint(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["kouryou_point"] = 500
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetKouryouPoint{AckHandle: 100}
	handleMsgMhfGetKouryouPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
		if len(ackData) < 4 {
			t.Fatal("AckData too short")
		}
		points := binary.BigEndian.Uint32(ackData[:4])
		if points != 500 {
			t.Errorf("points = %d, want 500", points)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetKouryouPoint_Error(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.readErr = errors.New("db error")
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetKouryouPoint{AckHandle: 100}
	handleMsgMhfGetKouryouPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		points := binary.BigEndian.Uint32(ackData[:4])
		if points != 0 {
			t.Errorf("points = %d, want 0 on error", points)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAddKouryouPoint(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["kouryou_point"] = 100
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     200,
		KouryouPoints: 50,
	}
	handleMsgMhfAddKouryouPoint(session, pkt)

	if charRepo.ints["kouryou_point"] != 150 {
		t.Errorf("kouryou_point = %d, want 150", charRepo.ints["kouryou_point"])
	}

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		points := binary.BigEndian.Uint32(ackData[:4])
		if points != 150 {
			t.Errorf("response points = %d, want 150", points)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAddKouryouPoint_Error(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.adjustErr = errors.New("db error")
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     200,
		KouryouPoints: 50,
	}
	handleMsgMhfAddKouryouPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		points := binary.BigEndian.Uint32(ackData[:4])
		if points != 0 {
			t.Errorf("response points = %d, want 0 on error", points)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfExchangeKouryouPoint(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["kouryou_point"] = 10000
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeKouryouPoint{
		AckHandle:     300,
		KouryouPoints: 10000,
	}
	handleMsgMhfExchangeKouryouPoint(session, pkt)

	if charRepo.ints["kouryou_point"] != 0 {
		t.Errorf("kouryou_point = %d, want 0 after exchange", charRepo.ints["kouryou_point"])
	}

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		points := binary.BigEndian.Uint32(ackData[:4])
		if points != 0 {
			t.Errorf("response points = %d, want 0", points)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfExchangeKouryouPoint_Error(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.adjustErr = errors.New("db error")
	server.charRepo = charRepo
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeKouryouPoint{
		AckHandle:     300,
		KouryouPoints: 5000,
	}
	handleMsgMhfExchangeKouryouPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should still respond on error")
		}
	default:
		t.Fatal("No response queued")
	}
}
