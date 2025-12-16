package channelserver

import (
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
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

	s.logger.Info(fmt.Sprintf("Broadcasting new object: %s (%d)", s.Name, newObj.id))
	s.stage.BroadcastMHF(dupObjUpdate, s)
}

func handleMsgSysDeleteObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysPositionObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPositionObject)
	if s.server.erupeConfig.DebugOptions.LogInboundMessages {
		fmt.Printf("[%s] with objectID [%d] move to (%f,%f,%f)\n\n", s.Name, pkt.ObjID, pkt.X, pkt.Y, pkt.Z)
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

func handleMsgSysRotateObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDuplicateObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetObjectBinary(s *Session, p mhfpacket.MHFPacket) {
	_ = p.(*mhfpacket.MsgSysSetObjectBinary)
	/* This causes issues with PS3 as this actually sends with endiness!
	for _, session := range s.server.sessions {
		if session.charID == s.charID {
			s.server.userBinaryPartsLock.Lock()
			s.server.userBinaryParts[userBinaryPartID{charID: s.charID, index: 3}] = pkt.RawDataPayload
			s.server.userBinaryPartsLock.Unlock()
			msg := &mhfpacket.MsgSysNotifyUserBinary{
				CharID:     s.charID,
				BinaryType: 3,
			}
			s.server.BroadcastMHF(msg, s)
		}
	}
	*/
}

func handleMsgSysGetObjectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetObjectOwner(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysUpdateObjectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCleanupObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAddObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDelObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDispObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysHideObject(s *Session, p mhfpacket.MHFPacket) {}
