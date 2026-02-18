package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/network/mhfpacket"
)

// Temporary function to just return no results for a MSG_MHF_ENUMERATE* packet
func stubEnumerateNoResults(s *Session, ackHandle uint32) {
	enumBf := byteframe.NewByteFrame()
	enumBf.WriteUint32(0) // Entry count (count for quests, rankings, events, etc.)

	doAckBufSucceed(s, ackHandle, enumBf.Data())
}

func doAckEarthSucceed(s *Session, ackHandle uint32, data []*byteframe.ByteFrame) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(s.server.erupeConfig.EarthID))
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(uint32(len(data)))
	for i := range data {
		bf.WriteBytes(data[i].Data())
	}
	doAckBufSucceed(s, ackHandle, bf.Data())
}

func doAckBufSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckBufFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        1,
		AckData:          data,
	})
}

func doAckSimpleSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckSimpleFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        1,
		AckData:          data,
	})
}

func updateRights(s *Session) {
	rightsInt := uint32(2)
	_ = s.server.db.QueryRow("SELECT rights FROM users u INNER JOIN characters c ON u.id = c.user_id WHERE c.id = $1", s.charID).Scan(&rightsInt)
	s.courses, rightsInt = mhfcourse.GetCourseStruct(rightsInt)
	update := &mhfpacket.MsgSysUpdateRight{
		ClientRespAckHandle: 0,
		Bitfield:            rightsInt,
		Rights:              s.courses,
		UnkSize:             0,
	}
	s.QueueSendMHFNonBlocking(update)
}
