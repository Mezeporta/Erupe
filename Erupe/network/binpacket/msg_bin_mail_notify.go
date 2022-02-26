package binpacket

import (
	"github.com/Solenataris/Erupe/common/stringsupport"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type MsgBinMailNotify struct {
	SenderName string
}

func (m MsgBinMailNotify) Parse(bf *byteframe.ByteFrame) error {
	panic("implement me")
}

func (m MsgBinMailNotify) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(0x01) // Unk
	byteName, _ := stringsupport.ConvertUTF8ToShiftJIS(m.SenderName)

	bf.WriteBytes(byteName)
	bf.WriteBytes(make([]byte, 21-len(byteName)))

	return nil
}

func (m MsgBinMailNotify) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}
