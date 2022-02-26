package channelserver

import (
	"fmt"
	"time"

	"github.com/Solenataris/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

func handleMsgSysCreateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateStage)

	s.server.stagesLock.Lock()
		stage := NewStage(pkt.StageID)
		stage.maxPlayers = uint16(pkt.PlayerCount)
		s.server.stages[stage.id] = stage
	s.server.stagesLock.Unlock()
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysStageDestruct(s *Session, p mhfpacket.MHFPacket) {}

func doStageTransfer(s *Session, ackHandle uint32, stageID string) {
	// Remove this session from old stage clients list and put myself in the new one.
	s.server.stagesLock.Lock()
	newStage, gotNewStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	if s.stage != nil {
		removeSessionFromStage(s)
	}

	// Add the new stage.
	if gotNewStage {
		newStage.Lock()
		newStage.clients[s] = s.charID
		newStage.Unlock()
	} else {
		// Fix stages
		s.logger.Info("Fix Map Appliqued")
		s.server.stagesLock.Lock()
		s.server.stages[stageID] = NewStage(stageID)
		newStage = s.server.stages[stageID]
		s.server.stagesLock.Unlock()
		newStage.Lock()
		newStage.clients[s] = s.charID
		newStage.Unlock()
	}

	// Save our new stage ID and pointer to the new stage itself.
	s.Lock()
	s.stageID = string(stageID)
	s.stage = newStage
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	s.QueueSendMHF(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	doAckSimpleSucceed(s, ackHandle, []byte{0x00, 0x00, 0x00, 0x00})

	// Notify existing stage clients that this new client has entered.
	s.logger.Info("Sending MsgSysInsertUser")
	if s.stage != nil { // avoids lock up when using bed for dream quests
		s.stage.BroadcastMHF(&mhfpacket.MsgSysInsertUser{
			CharID: s.charID,
		}, s)

		// It seems to be acceptable to recast all MSG_SYS_SET_USER_BINARY messages so far,
		// players are still notified when a new player has joined the stage.
		// These extra messages may not be needed
		//s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		//	CharID:     s.charID,
		//	BinaryType: 1,
		//}, s)
		//s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		//	CharID:     s.charID,
		//	BinaryType: 2,
		//}, s)
		//s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		//	CharID:     s.charID,
		//	BinaryType: 3,
		//}, s)

		//Notify the entree client about all of the existing clients in the stage.
		s.logger.Info("Notifying entree about existing stage clients")
		s.stage.RLock()
		clientNotif := byteframe.NewByteFrame()
		for session := range s.stage.clients {
			var cur mhfpacket.MHFPacket
				cur = &mhfpacket.MsgSysInsertUser{
					CharID: session.charID,
				}
				clientNotif.WriteUint16(uint16(cur.Opcode()))
				cur.Build(clientNotif, session.clientContext)

				cur = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.charID,
					BinaryType: 1,
				}
				clientNotif.WriteUint16(uint16(cur.Opcode()))
				cur.Build(clientNotif, session.clientContext)

				cur = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.charID,
					BinaryType: 2,
				}
				clientNotif.WriteUint16(uint16(cur.Opcode()))
				cur.Build(clientNotif, session.clientContext)

				cur = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.charID,
					BinaryType: 3,
				}
				clientNotif.WriteUint16(uint16(cur.Opcode()))
				cur.Build(clientNotif, session.clientContext)
		}
		s.stage.RUnlock()
		clientNotif.WriteUint16(0x0010) // End it.
		s.QueueSend(clientNotif.Data())

		// Notify the client to duplicate the existing objects.
		s.logger.Info("Notifying entree about existing stage objects")
		clientDupObjNotif := byteframe.NewByteFrame()
		s.stage.RLock()
		for _, obj := range s.stage.objects {
				cur := &mhfpacket.MsgSysDuplicateObject{
					ObjID:       obj.id,
					X:           obj.x,
					Y:           obj.y,
					Z:           obj.z,
					Unk0:        0,
					OwnerCharID: obj.ownerCharID,
				}
				clientDupObjNotif.WriteUint16(uint16(cur.Opcode()))
				cur.Build(clientDupObjNotif, s.clientContext)
		}
		s.stage.RUnlock()
		clientDupObjNotif.WriteUint16(0x0010) // End it.
		s.QueueSend(clientDupObjNotif.Data())
	}
}

func removeSessionFromStage(s *Session) {
	s.stage.Lock()
	defer s.stage.Unlock()

	// Remove client from old stage.
	delete(s.stage.clients, s)
	delete(s.stage.reservedClientSlots, s.charID)

	// Delete old stage objects owned by the client.
	s.logger.Info("Sending MsgSysDeleteObject to old stage clients")
	for objID, stageObject := range s.stage.objects {
		if stageObject.ownerCharID == s.charID {
			// Broadcast the deletion to clients in the stage.
			s.stage.BroadcastMHF(&mhfpacket.MsgSysDeleteObject{
				ObjID: stageObject.id,
			}, s)
			// TODO(Andoryuuta): Should this be sent to the owner's client as well? it currently isn't.
			// Actually delete it form the objects map.
			delete(s.stage.objects, objID)
		}
}
	for objListID, stageObjectList := range s.stage.objectList {
		if stageObjectList.charid == s.charID {
			//Added to prevent duplicates from flooding ObjectMap and causing server hangs
			s.stage.objectList[objListID].status=false
			s.stage.objectList[objListID].charid=0
			}
		}
}


func handleMsgSysEnterStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)
	fmt.Printf("The Stage is %s\n",pkt.StageID)
	// Push our current stage ID to the movement stack before entering another one.
	s.Lock()
	s.stageMoveStack.Push(s.stageID)
	s.Unlock()

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysBackStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysBackStage)

	if s.stage != nil {
		removeSessionFromStage(s)
	}

	// Transfer back to the saved stage ID before the previous move or enter.
	s.Lock()
	backStage, err := s.stageMoveStack.Pop()
	s.Unlock()

	if err != nil {
		panic(err)
	}

	doStageTransfer(s, pkt.AckHandle, backStage)
}

func handleMsgSysMoveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysMoveStage)

	// Push our current stage ID to the movement stack before entering another one.
	s.Lock()
	s.stageMoveStack.Push(s.stageID)
	s.Unlock()

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysLeaveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockStage)
	// TODO(Andoryuuta): What does this packet _actually_ do?
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysUnlockStage(s *Session, p mhfpacket.MHFPacket) {
	s.reservationStage.RLock()
	defer s.reservationStage.RUnlock()

	destructMessage := &mhfpacket.MsgSysStageDestruct{}

	for charID := range s.reservationStage.reservedClientSlots {
		session := s.server.FindSessionByCharID(charID)
		session.QueueSendMHF(destructMessage)
	}

	s.server.Lock()
	defer s.server.Unlock()

	delete(s.server.stages, s.reservationStage.id)
}

func handleMsgSysReserveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserveStage)

	stageID := pkt.StageID
	fmt.Printf("Got reserve stage req, TargetCount:%v, StageID:%v\n", pkt.Unk0, stageID)

	// Try to get the stage
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	if !gotStage {
		s.logger.Fatal("Failed to get stage", zap.String("StageID", stageID))
	}

	// Try to reserve a slot, fail if full.
	stage.Lock()
	defer stage.Unlock()

	// Quick fix to allow readying up while party is full, more investigation needed
	// Reserve stage is also sent when a player is ready, probably need to parse the
	// request a little more thoroughly.
	if _, exists := stage.reservedClientSlots[s.charID]; exists {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	} else if uint16(len(stage.reservedClientSlots)) < stage.maxPlayers {
		// Add the charID to the stage's reservation map
		stage.reservedClientSlots[s.charID] = nil

		// Save the reservation stage in the session for later use in MsgSysUnreserveStage.
		s.Lock()
		s.reservationStage = stage
		s.Unlock()

		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	} else {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysUnreserveStage(s *Session, p mhfpacket.MHFPacket) {
	// Clear the saved reservation stage
	s.Lock()
	stage := s.reservationStage
	if stage != nil {
		s.reservationStage = nil
	}
	s.Unlock()

	// Remove the charID from the stage's reservation map
	if stage != nil {
		stage.Lock()
		_, exists := stage.reservedClientSlots[s.charID]
		if exists {
			delete(stage.reservedClientSlots, s.charID)
		}
		stage.Unlock()
	}
}

func handleMsgSysSetStagePass(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysWaitStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysWaitStageBinary)
	defer s.logger.Debug("MsgSysWaitStageBinary Done!")

	// Try to get the stage
	stageID := pkt.StageID
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// TODO(Andoryuuta): This is a hack for a binary part that none of the clients set, figure out what it represents.
	// In the packet captures, it seemingly comes out of nowhere, so presumably the server makes it.
	if pkt.BinaryType0 == 1 && pkt.BinaryType1 == 12 {
		// This might contain the hunter count, or max player count?
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}

	// If we got the stage, lock and try to get the data.
	var stageBinary []byte
	var gotBinary bool
	if gotStage {
		for {
			s.logger.Debug("MsgSysWaitStageBinary before lock and get stage")
			stage.Lock()
			stageBinary, gotBinary = stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
			stage.Unlock()
			s.logger.Debug("MsgSysWaitStageBinary after lock and get stage")

			if gotBinary {
				doAckBufSucceed(s, pkt.AckHandle, stageBinary)
				break
			} else {
				s.logger.Debug("Waiting stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))

				// Couldn't get binary, sleep for some time and try again.
				time.Sleep(2 * time.Second)
				continue
			}

			// TODO(Andoryuuta): Figure out what the game sends on timeout and implement it!
			/*
				if timeout {
					s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
					s.logger.Warn("Sending blank stage binary")
					doAckBufSucceed(s, pkt.AckHandle, []byte{})
					return
				}
			*/
		}
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}
}

func handleMsgSysSetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStageBinary)

	// Try to get the stage
	stageID := pkt.StageID
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// If we got the stage, lock and set the data.
	if gotStage {
		stage.Lock()
		stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}] = pkt.RawDataPayload
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}
	s.logger.Debug("handleMsgSysSetStageBinary Done!")
}

func handleMsgSysGetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetStageBinary)

	// Try to get the stage
	stageID := pkt.StageID
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// If we got the stage, lock and try to get the data.
	var stageBinary []byte
	var gotBinary bool
	if gotStage {
		stage.Lock()
		stageBinary, gotBinary = stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}

	if gotBinary {
		doAckBufSucceed(s, pkt.AckHandle, stageBinary)

	} else if pkt.BinaryType1 == 4 {
		// This particular type seems to be expecting data that isn't set
		// is it required before the party joining can be completed
		//s.QueueAck(pkt.AckHandle, []byte{0x01, 0x00, 0x00, 0x00, 0x10})

		// TODO(Andoryuuta): This doesn't fit a normal ack packet? where is this from?
		// This would be a buffered(0x01), non-error(0x00), with no data payload (size 0x00, 0x00) packet.
		// but for some reason has a 0x10 on the end that the client shouldn't parse?

		doAckBufSucceed(s, pkt.AckHandle, []byte{}) // Without the previous 0x10 suffix
	} else {
		s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
		s.logger.Warn("Sending blank stage binary")
		doAckBufSucceed(s, pkt.AckHandle, []byte{})
	}

	s.logger.Debug("MsgSysGetStageBinary Done!")
}

func handleMsgSysEnumerateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateStage)

	// Read-lock the server stage map.
	s.server.stagesLock.RLock()
	defer s.server.stagesLock.RUnlock()

	// Build the response
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(uint16(len(s.server.stages)))
	for sid, stage := range s.server.stages {
		stage.RLock()
		defer stage.RUnlock()
		if len(stage.reservedClientSlots)+len(stage.clients) == 0 {
			continue
		}

		resp.WriteUint16(uint16(len(stage.reservedClientSlots))) // Current players.
		resp.WriteUint16(0)                                      // Unknown value

		var hasDeparted uint16
		if stage.hasDeparted {
			hasDeparted = 1
		}
		resp.WriteUint16(hasDeparted)           // HasDeparted.
		resp.WriteUint16(stage.maxPlayers)      // Max players.
		resp.WriteBool(len(stage.password) > 0) // Password protected.
		resp.WriteUint8(uint8(len(sid)))
		resp.WriteBytes([]byte(sid))
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
