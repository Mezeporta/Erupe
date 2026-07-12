package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

func handleMsgSysCreateObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateObject)

	s.stage.Lock()
	newObj := &Object{
		id:          s.getObjectId(),
		ownerCharID: s.charID,
		x:           pkt.X,
		y:           pkt.Y,
		z:           pkt.Z,
	}
	s.stage.objects[s.charID] = newObj
	s.stage.Unlock()

	// Response to our requesting client.
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(newObj.id) // New local obj handle.
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
	// Duplicate the object creation to all sessions in the same stage.
	dupObjUpdate := &mhfpacket.MsgSysDuplicateObject{
		ObjID:       newObj.id,
		X:           newObj.x,
		Y:           newObj.y,
		Z:           newObj.z,
		OwnerCharID: newObj.ownerCharID,
	}

	s.logger.Info("Broadcasting new object", zap.String("name", s.Name), zap.Uint32("objectID", newObj.id))
	s.stage.BroadcastMHF(dupObjUpdate, s)
}

// handleMsgSysDeleteObject removes the sender's own synced stage object and
// relays the deletion to the rest of the stage, mirroring the same
// remove-then-broadcast pattern already used server-side when a client
// leaves a stage (see removeSessionFromStage). A client may only delete the
// object it owns -- the requested ObjID must match the object on record for
// the sender's charID, or the request is dropped.
func handleMsgSysDeleteObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysDeleteObject)

	s.stage.Lock()
	object, ok := s.stage.objects[s.charID]
	if ok && object.id == pkt.ObjID {
		delete(s.stage.objects, s.charID)
	} else {
		ok = false
	}
	s.stage.Unlock()

	if ok {
		s.stage.BroadcastMHF(pkt, s)
	}
}

func handleMsgSysPositionObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPositionObject)
	if s.server.erupeConfig.DebugOptions.LogInboundMessages {
		s.logger.Debug("Object position update",
			zap.String("name", s.Name),
			zap.Uint32("objectID", pkt.ObjID),
			zap.Float32("x", pkt.X),
			zap.Float32("y", pkt.Y),
			zap.Float32("z", pkt.Z),
		)
	}
	s.stage.Lock()
	object, ok := s.stage.objects[s.charID]
	if ok {
		object.x = pkt.X
		object.y = pkt.Y
		object.z = pkt.Z
	}
	s.stage.Unlock()
	// One of the few packets we can just re-broadcast directly.
	s.stage.BroadcastMHF(pkt, s)
}

// handleMsgSysRotateObject mirrors handleMsgSysPositionObject's pattern:
// update the sender's own synced stage object and re-broadcast the same
// packet to the rest of the stage so other clients turn the model to match.
func handleMsgSysRotateObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRotateObject)

	s.stage.Lock()
	object, ok := s.stage.objects[s.charID]
	if ok {
		object.rotation = pkt.Rotation
	}
	s.stage.Unlock()

	s.stage.BroadcastMHF(pkt, s)
}

func handleMsgSysDuplicateObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysSetObjectBinary(s *Session, p mhfpacket.MHFPacket) {
	_ = p.(*mhfpacket.MsgSysSetObjectBinary)
	/* This causes issues with PS3 as this actually sends with endiness!
	for _, session := range s.server.sessions {
		if session.charID == s.charID {
			s.server.userBinary.Set(s.charID, 3, pkt.RawDataPayload)
			msg := &mhfpacket.MsgSysNotifyUserBinary{
				CharID:     s.charID,
				BinaryType: 3,
			}
			s.server.BroadcastMHF(msg, s)
		}
	}
	*/
}

// handleMsgSysGetObjectBinary answers a request for another stage object's
// synced binary state. Erupe doesn't persist per-object binary payloads yet
// (see handleMsgSysSetObjectBinary's PS3-endianness caveat above), so this
// always acks a zero-length result -- the same "not found" shape the PC
// client itself falls back to when its local lookup misses (decompiled
// pkt_handler_MSG_SYS_GET_OBJECT_BINARY replies with a zero-length payload
// rather than an error ack), so a real client handles this gracefully.
func handleMsgSysGetObjectBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetObjectBinary)
	doAckBufSucceed(s, pkt.AckHandle, []byte{})
}

// handleMsgSysGetObjectOwner answers a request for the owning character of a
// stage object, resolved from the same s.stage.objects map handleMsgSysCreateObject
// populates and handleMsgSysDeleteObject prunes.
func handleMsgSysGetObjectOwner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetObjectOwner)

	var ownerCharID uint32
	s.stage.RLock()
	for _, object := range s.stage.objects {
		if object.id == pkt.ObjID {
			ownerCharID = object.ownerCharID
			break
		}
	}
	s.stage.RUnlock()

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(ownerCharID)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysUpdateObjectBinary(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysCleanupObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysAddObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysDelObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysDispObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysHideObject(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented
