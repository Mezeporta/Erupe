// Package protocol implements MHF network protocol message building and parsing.
package protocol

// Packet opcodes (subset from Erupe's network/packetid.go iota).
const (
	MSG_SYS_ACK              uint16 = 0x0012
	MSG_SYS_LOGIN            uint16 = 0x0014
	MSG_SYS_LOGOUT           uint16 = 0x0015
	MSG_SYS_PING             uint16 = 0x0017
	MSG_SYS_CAST_BINARY      uint16 = 0x0018
	MSG_SYS_TIME             uint16 = 0x001A
	MSG_SYS_CASTED_BINARY    uint16 = 0x001B
	MSG_SYS_ISSUE_LOGKEY     uint16 = 0x001D
	MSG_SYS_ENTER_STAGE      uint16 = 0x0022
	MSG_SYS_ENUMERATE_STAGE  uint16 = 0x002F
	MSG_SYS_INSERT_USER      uint16 = 0x0050
	MSG_SYS_DELETE_USER      uint16 = 0x0051
	MSG_SYS_UPDATE_RIGHT     uint16 = 0x0058
	MSG_SYS_RIGHTS_RELOAD    uint16 = 0x005D
	MSG_MHF_LOADDATA         uint16 = 0x0061
	MSG_MHF_ENUMERATE_QUEST  uint16 = 0x009F
	MSG_MHF_GET_WEEKLY_SCHED uint16 = 0x00E1
)
