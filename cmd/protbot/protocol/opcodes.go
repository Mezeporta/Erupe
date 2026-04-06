// Package protocol implements MHF network protocol message building and parsing.
package protocol

// Packet opcodes (subset from Erupe's network/packetid.go iota).
const (
	MSG_SYS_ACK                   uint16 = 0x0012
	MSG_SYS_LOGIN                 uint16 = 0x0014
	MSG_SYS_LOGOUT                uint16 = 0x0015
	MSG_SYS_PING                  uint16 = 0x0017
	MSG_SYS_CAST_BINARY           uint16 = 0x0018
	MSG_SYS_TIME                  uint16 = 0x001A
	MSG_SYS_CASTED_BINARY         uint16 = 0x001B
	MSG_SYS_ISSUE_LOGKEY          uint16 = 0x001D
	MSG_SYS_ENTER_STAGE           uint16 = 0x0022
	MSG_SYS_ENUMERATE_STAGE       uint16 = 0x002F
	MSG_SYS_INSERT_USER           uint16 = 0x0050
	MSG_SYS_DELETE_USER           uint16 = 0x0051
	MSG_SYS_UPDATE_RIGHT          uint16 = 0x0058
	MSG_SYS_RIGHTS_RELOAD         uint16 = 0x005D
	MSG_MHF_LOADDATA              uint16 = 0x0061
	MSG_MHF_ENUMERATE_QUEST       uint16 = 0x009F
	MSG_MHF_GET_ACHIEVEMENT       uint16 = 0x00D4
	MSG_MHF_ADD_ACHIEVEMENT       uint16 = 0x00D6
	MSG_MHF_DISPLAYED_ACHIEVEMENT uint16 = 0x00D8
	MSG_MHF_GET_WEEKLY_SCHED      uint16 = 0x00E1

	// Boost time / login boost (issue #187)
	MSG_MHF_GET_BOOST_TIME              uint16 = 0x0126
	MSG_MHF_GET_BOOST_TIME_LIMIT        uint16 = 0x0128
	MSG_MHF_GET_BOOST_RIGHT             uint16 = 0x013E
	MSG_MHF_START_BOOST_TIME            uint16 = 0x013F
	MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS uint16 = 0x0159
	MSG_MHF_USE_KEEP_LOGIN_BOOST        uint16 = 0x015A

	// Gacha (issues #175, gacha logging)
	MSG_MHF_GET_GACHA_POINT    uint16 = 0x0131
	MSG_MHF_RECEIVE_GACHA_ITEM uint16 = 0x0137
	MSG_MHF_PLAY_NORMAL_GACHA  uint16 = 0x0150
)
