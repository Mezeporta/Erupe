package channelserver

import (
	"fmt"
	"strings"
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

func handleMsgSysCreateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateStage)
	stage := NewStage(pkt.StageID)
	stage.host = s
	stage.maxPlayers = uint16(pkt.PlayerCount)
	if s.server.stages.StoreIfAbsent(pkt.StageID, stage) {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	} else {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysStageDestruct(s *Session, p mhfpacket.MHFPacket) {}

func doStageTransfer(s *Session, ackHandle uint32, stageID string) {
	stage, created := s.server.stages.GetOrCreate(stageID)

	stage.Lock()
	if created {
		stage.host = s
	}
	stage.clients[s] = s.charID
	stage.Unlock()

	// Ensure this session no longer belongs to reservations.
	if s.stage != nil {
		removeSessionFromStage(s)
	}

	// Save our new stage pointer.
	s.Lock()
	s.stage = stage
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	// Use blocking send to ensure this critical cleanup packet is not dropped.
	s.QueueSendMHF(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	doAckSimpleSucceed(s, ackHandle, []byte{0x00, 0x00, 0x00, 0x00})

	newNotif := byteframe.NewByteFrame()

	// Cast existing user data to new user
	if !s.loaded {
		s.loaded = true

		// Lock server to safely iterate over sessions map
		// We need to copy the session list first to avoid holding the lock during packet building
		s.server.Lock()
		var sessionList []*Session
		for _, session := range s.server.sessions {
			if s == session || !session.loaded {
				continue
			}
			sessionList = append(sessionList, session)
		}
		s.server.Unlock()

		// Build packets for each session without holding the lock
		var temp mhfpacket.MHFPacket
		for _, session := range sessionList {
			temp = &mhfpacket.MsgSysInsertUser{CharID: session.charID}
			newNotif.WriteUint16(uint16(temp.Opcode()))
			_ = temp.Build(newNotif, s.clientContext)
			for i := 0; i < 3; i++ {
				temp = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.charID,
					BinaryType: uint8(i + 1),
				}
				newNotif.WriteUint16(uint16(temp.Opcode()))
				_ = temp.Build(newNotif, s.clientContext)
			}
		}
	}

	if s.stage != nil { // avoids lock up when using bed for dream quests
		// Notify the client to duplicate the existing objects.
		s.logger.Info(fmt.Sprintf("Sending existing stage objects to %s", s.Name))

		// Lock stage to safely iterate over objects map
		// We need to copy the objects list first to avoid holding the lock during packet building
		s.stage.RLock()
		var objectList []*Object
		for _, obj := range s.stage.objects {
			if obj.ownerCharID == s.charID {
				continue
			}
			objectList = append(objectList, obj)
		}
		s.stage.RUnlock()

		// Build packets for each object without holding the lock
		var temp mhfpacket.MHFPacket
		for _, obj := range objectList {
			temp = &mhfpacket.MsgSysDuplicateObject{
				ObjID:       obj.id,
				X:           obj.x,
				Y:           obj.y,
				Z:           obj.z,
				Unk0:        0,
				OwnerCharID: obj.ownerCharID,
			}
			newNotif.WriteUint16(uint16(temp.Opcode()))
			_ = temp.Build(newNotif, s.clientContext)
		}
	}

	// FIX: Always send stage transfer packet, even if empty.
	// The client expects this packet to complete the zone change, regardless of content.
	// Previously, if newNotif was empty (no users, no objects), no packet was sent,
	// causing the client to timeout after 60 seconds.
	s.QueueSend(newNotif.Data())
}

func destructEmptyStages(s *Session) {
	s.server.stages.Range(func(id string, stage *Stage) bool {
		// Destroy empty Quest/My series/Guild stages.
		if id[3:5] == "Qs" || id[3:5] == "Ms" || id[3:5] == "Gs" || id[3:5] == "Ls" {
			stage.Lock()
			isEmpty := len(stage.reservedClientSlots) == 0 && len(stage.clients) == 0
			stage.Unlock()

			if isEmpty {
				s.server.stages.Delete(id)
				s.logger.Debug("Destructed stage", zap.String("stage.id", id))
			}
		}
		return true
	})
}

func removeSessionFromStage(s *Session) {
	// Acquire stage lock to protect concurrent access to clients and objects maps
	// This prevents race conditions when multiple goroutines access these maps
	s.stage.Lock()

	// Remove client from old stage.
	delete(s.stage.clients, s)

	// Delete old stage objects owned by the client.
	// We must copy the objects to delete to avoid modifying the map while iterating
	var objectsToDelete []*Object
	for _, object := range s.stage.objects {
		if object.ownerCharID == s.charID {
			objectsToDelete = append(objectsToDelete, object)
		}
	}

	// Delete from map while still holding lock
	for _, object := range objectsToDelete {
		delete(s.stage.objects, object.ownerCharID)
	}

	// CRITICAL FIX: Unlock BEFORE broadcasting to avoid deadlock
	// BroadcastMHF also tries to lock the stage, so we must release our lock first
	s.stage.Unlock()

	// Now broadcast the deletions (without holding the lock)
	for _, object := range objectsToDelete {
		s.stage.BroadcastMHF(&mhfpacket.MsgSysDeleteObject{ObjID: object.id}, s)
	}

	destructEmptyStages(s)
	destructEmptySemaphores(s)
}

func isStageFull(s *Session, StageID string) bool {
	stage, exists := s.server.stages.Get(StageID)

	if exists {
		// Lock stage to safely check client counts
		// Read the values we need while holding RLock, then release immediately
		// to avoid deadlock with other functions that might hold server lock
		stage.RLock()
		reserved := len(stage.reservedClientSlots)
		clients := len(stage.clients)
		_, hasReservation := stage.reservedClientSlots[s.charID]
		maxPlayers := stage.maxPlayers
		stage.RUnlock()

		if hasReservation {
			return false
		}
		return reserved+clients >= int(maxPlayers)
	}
	return false
}

func handleMsgSysEnterStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)

	if isStageFull(s, pkt.StageID) {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	// Push our current stage ID to the movement stack before entering another one.
	if s.stage != nil {
		s.stage.Lock()
		s.stage.reservedClientSlots[s.charID] = false
		s.stage.Unlock()
		s.stageMoveStack.Push(s.stage.id)
	}

	if s.reservationStage != nil {
		s.reservationStage = nil
	}

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysBackStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysBackStage)

	// Transfer back to the saved stage ID before the previous move or enter.
	backStage, err := s.stageMoveStack.Pop()
	if backStage == "" || err != nil {
		backStage = "sl1Ns200p0a0u0"
	}

	if isStageFull(s, backStage) {
		s.stageMoveStack.Push(backStage)
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	if s.stage != nil {
		s.stage.Lock()
		delete(s.stage.reservedClientSlots, s.charID)
		s.stage.Unlock()
	}

	backStagePtr, exists := s.server.stages.Get(backStage)
	if exists {
		backStagePtr.Lock()
		delete(backStagePtr.reservedClientSlots, s.charID)
		backStagePtr.Unlock()
	}

	doStageTransfer(s, pkt.AckHandle, backStage)
}

func handleMsgSysMoveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysMoveStage)

	if isStageFull(s, pkt.StageID) {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysLeaveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockStage)
	stage, exists := s.server.stages.Get(pkt.StageID)
	if exists {
		stage.Lock()
		stage.locked = true
		stage.Unlock()
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgSysUnlockStage(s *Session, p mhfpacket.MHFPacket) {
	if s.reservationStage != nil {
		// Read reserved client slots under stage RLock
		s.reservationStage.RLock()
		var charIDs []uint32
		for charID := range s.reservationStage.reservedClientSlots {
			charIDs = append(charIDs, charID)
		}
		stageID := s.reservationStage.id
		s.reservationStage.RUnlock()

		for _, charID := range charIDs {
			session := s.server.FindSessionByCharID(charID)
			if session != nil {
				session.QueueSendMHFNonBlocking(&mhfpacket.MsgSysStageDestruct{})
			}
		}

		s.server.stages.Delete(stageID)
	}

	destructEmptyStages(s)
}

func handleMsgSysReserveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserveStage)
	stage, exists := s.server.stages.Get(pkt.StageID)
	if exists {
		stage.Lock()
		defer stage.Unlock()
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			switch pkt.Ready {
			case 1: // 0x01
				stage.reservedClientSlots[s.charID] = false
			case 17: // 0x11
				stage.reservedClientSlots[s.charID] = true
			}
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else if uint16(len(stage.reservedClientSlots)) < stage.maxPlayers {
			if stage.locked {
				doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
				return
			}
			if len(stage.password) > 0 {
				if stage.password != s.stagePass {
					doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
					return
				}
			}
			stage.reservedClientSlots[s.charID] = false
			// Save the reservation stage in the session for later use in MsgSysUnreserveStage.
			s.Lock()
			s.reservationStage = stage
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	} else {
		s.logger.Error("Failed to get stage", zap.String("StageID", pkt.StageID))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgSysUnreserveStage(s *Session, p mhfpacket.MHFPacket) {
	s.Lock()
	stage := s.reservationStage
	s.reservationStage = nil
	s.Unlock()
	if stage != nil {
		stage.Lock()
		delete(stage.reservedClientSlots, s.charID)
		stage.Unlock()
	}
}

func handleMsgSysSetStagePass(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStagePass)
	s.Lock()
	stage := s.reservationStage
	s.Unlock()
	if stage != nil {
		stage.Lock()
		// Will only exist if host.
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			stage.password = pkt.Password
		}
		stage.Unlock()
	} else {
		// Store for use on next ReserveStage.
		s.Lock()
		s.stagePass = pkt.Password
		s.Unlock()
	}
}

func handleMsgSysSetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStageBinary)
	stage, exists := s.server.stages.Get(pkt.StageID)
	if exists {
		stage.Lock()
		stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}] = pkt.RawDataPayload
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
	}
}

func handleMsgSysGetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetStageBinary)
	stage, exists := s.server.stages.Get(pkt.StageID)
	if exists {
		stage.Lock()
		if binaryData, exists := stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]; exists {
			doAckBufSucceed(s, pkt.AckHandle, binaryData)
		} else if pkt.BinaryType1 == 4 {
			// Server-generated binary used for guild room checks and lobby state.
			// Earlier clients (G1) crash on a completely empty response when parsing
			// this during lobby initialization, so return a minimal valid structure
			// with a zero entry count.
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		}
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
	s.logger.Debug("MsgSysGetStageBinary Done!")
}

func handleMsgSysWaitStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysWaitStageBinary)
	stage, exists := s.server.stages.Get(pkt.StageID)
	if exists {
		if pkt.BinaryType0 == 1 && pkt.BinaryType1 == 12 {
			// This might contain the hunter count, or max player count?
			doAckBufSucceed(s, pkt.AckHandle, []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			return
		}
		for i := 0; i < 10; i++ {
			s.logger.Debug("MsgSysWaitStageBinary before lock and get stage")
			stage.Lock()
			stageBinary, gotBinary := stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
			stage.Unlock()
			s.logger.Debug("MsgSysWaitStageBinary after lock and get stage")
			if gotBinary {
				doAckBufSucceed(s, pkt.AckHandle, stageBinary)
				return
			} else {
				s.logger.Debug("Waiting stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
				time.Sleep(1 * time.Second)
				continue
			}
		}
		s.logger.Warn("MsgSysWaitStageBinary stage binary timeout")
		doAckBufSucceed(s, pkt.AckHandle, []byte{})
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
	s.logger.Debug("MsgSysWaitStageBinary Done!")
}

func handleMsgSysEnumerateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateStage)

	// Build the response
	bf := byteframe.NewByteFrame()
	var joinable uint16
	bf.WriteUint16(0)
	s.server.stages.Range(func(sid string, stage *Stage) bool {
		stage.RLock()

		if len(stage.reservedClientSlots) == 0 && len(stage.clients) == 0 {
			stage.RUnlock()
			return true
		}
		if !strings.Contains(stage.id, pkt.StagePrefix) {
			stage.RUnlock()
			return true
		}
		joinable++

		bf.WriteUint16(uint16(len(stage.reservedClientSlots)))
		bf.WriteUint16(uint16(len(stage.clients)))
		if strings.HasPrefix(stage.id, "sl2Ls") {
			bf.WriteUint16(uint16(len(stage.clients) + len(stage.reservedClientSlots)))
		} else {
			bf.WriteUint16(uint16(len(stage.clients)))
		}
		bf.WriteUint16(stage.maxPlayers)
		var flags uint8
		if stage.locked {
			flags |= 1
		}
		if len(stage.password) > 0 {
			flags |= 2
		}
		bf.WriteUint8(flags)
		ps.Uint8(bf, sid, false)
		stage.RUnlock()
		return true
	})
	_, _ = bf.Seek(0, 0)
	bf.WriteUint16(joinable)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
