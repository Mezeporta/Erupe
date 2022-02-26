package channelserver

import (
	"encoding/hex"

	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) { // RAVIENTE USE THIS
	// RAVI EVENT
	pkt := p.(*mhfpacket.MsgSysOperateRegister)

	doAckSimpleSucceed(s, pkt.AckHandle, pkt.RawDataPayload)
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)

	// ORION TEMPORARY DISABLE (IN WORK)
	// ravi response
	data, _ := hex.DecodeString("000C000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	//data, _ := hex.DecodeString("00000076001d0001b2c4000227d1000221040000a959000000000000000000000000000000000000000000532d1c0010ee8e001fe0010007f463000000000017e53e00072e250053937a0000194a00002d5a000000000000000000004eb300004cd700000000000008a90000be400001bb16000005dd00000014")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}
