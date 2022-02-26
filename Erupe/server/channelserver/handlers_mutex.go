package channelserver

import "github.com/Solenataris/Erupe/network/mhfpacket"

func handleMsgSysCreateMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateOpenMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysOpenMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCloseMutex(s *Session, p mhfpacket.MHFPacket) {}
