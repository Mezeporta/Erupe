package scenario

import (
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"

	"erupe-ce/cmd/protbot/protocol"
)

// ChatMessage holds a parsed incoming chat message.
type ChatMessage struct {
	ChatType   uint8
	SenderName string
	Message    string
}

// SendChat sends a chat message via MSG_SYS_CAST_BINARY with a MsgBinChat payload.
// broadcastType controls delivery scope: 0x03 = stage, 0x06 = world.
func SendChat(ch *protocol.ChannelConn, broadcastType, chatType uint8, message, senderName string) error {
	payload := protocol.BuildChatPayload(chatType, message, senderName)
	pkt := protocol.BuildCastBinaryPacket(broadcastType, 1, payload)
	fmt.Printf("[chat] Sending chat (type=%d, broadcast=%d): %s\n", chatType, broadcastType, message)
	return ch.SendPacket(pkt)
}

// ChatCallback is invoked when a chat message is received.
type ChatCallback func(msg ChatMessage)

// ListenChat registers a handler on MSG_SYS_CASTED_BINARY that parses chat
// messages (messageType=1) and invokes the callback.
func ListenChat(ch *protocol.ChannelConn, cb ChatCallback) {
	ch.OnPacket(protocol.MSG_SYS_CASTED_BINARY, func(opcode uint16, data []byte) {
		// MSG_SYS_CASTED_BINARY layout from server:
		//   uint32 unk
		//   uint8  broadcastType
		//   uint8  messageType
		//   uint16 dataSize
		//   []byte payload
		if len(data) < 8 {
			return
		}
		messageType := data[5]
		if messageType != 1 { // Only handle chat messages.
			return
		}
		bf := byteframe.NewByteFrameFromBytes(data)
		_ = bf.ReadUint32() // unk
		_ = bf.ReadUint8()  // broadcastType
		_ = bf.ReadUint8()  // messageType
		dataSize := bf.ReadUint16()
		if dataSize == 0 {
			return
		}
		payload := bf.ReadBytes(uint(dataSize))

		// Parse MsgBinChat inner payload.
		pbf := byteframe.NewByteFrameFromBytes(payload)
		_ = pbf.ReadUint8()  // unk0
		chatType := pbf.ReadUint8()
		_ = pbf.ReadUint16() // flags
		_ = pbf.ReadUint16() // senderNameLen
		_ = pbf.ReadUint16() // messageLen
		msg := stringsupport.SJISToUTF8Lossy(pbf.ReadNullTerminatedBytes())
		sender := stringsupport.SJISToUTF8Lossy(pbf.ReadNullTerminatedBytes())

		cb(ChatMessage{
			ChatType:   chatType,
			SenderName: sender,
			Message:    msg,
		})
	})
}
