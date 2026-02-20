package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
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

// loadCharacterData loads a column from the characters table and sends it as
// a buffered ack response. If the data is empty/nil, defaultData is sent instead.
func loadCharacterData(s *Session, ackHandle uint32, column string, defaultData []byte) {
	data, err := s.server.charRepo.LoadColumn(s.charID, column)
	if err != nil {
		s.logger.Error("Failed to load "+column, zap.Error(err))
	}
	if len(data) == 0 && defaultData != nil {
		data = defaultData
	}
	doAckBufSucceed(s, ackHandle, data)
}

// saveCharacterData saves data to a column in the characters table with size
// validation, optional save dump, and a simple ack response.
func saveCharacterData(s *Session, ackHandle uint32, column string, data []byte, maxSize int) {
	if maxSize > 0 && len(data) > maxSize {
		s.logger.Warn("Payload too large for "+column, zap.Int("len", len(data)), zap.Int("max", maxSize))
		doAckSimpleFail(s, ackHandle, make([]byte, 4))
		return
	}
	dumpSaveData(s, data, column)
	err := s.server.charRepo.SaveColumn(s.charID, column, data)
	if err != nil {
		s.logger.Error("Failed to save "+column, zap.Error(err))
		doAckSimpleFail(s, ackHandle, make([]byte, 4))
		return
	}
	doAckSimpleSucceed(s, ackHandle, make([]byte, 4))
}

// readCharacterInt reads a single integer column from the characters table.
// Returns 0 for NULL columns via COALESCE.
func readCharacterInt(s *Session, column string) (int, error) {
	return s.server.charRepo.ReadInt(s.charID, column)
}

// adjustCharacterInt atomically adds delta to an integer column and returns the new value.
// Handles NULL columns via COALESCE (NULL + delta = delta).
func adjustCharacterInt(s *Session, column string, delta int) (int, error) {
	return s.server.charRepo.AdjustInt(s.charID, column, delta)
}

func updateRights(s *Session) {
	rightsInt, err := s.server.userRepo.GetRights(s.userID)
	if err != nil {
		rightsInt = 2
	}
	s.courses, rightsInt = mhfcourse.GetCourseStruct(rightsInt, s.server.erupeConfig.DefaultCourses)
	update := &mhfpacket.MsgSysUpdateRight{
		ClientRespAckHandle: 0,
		Bitfield:            rightsInt,
		Rights:              s.courses,
		TokenLength:         0,
	}
	s.QueueSendMHFNonBlocking(update)
}
