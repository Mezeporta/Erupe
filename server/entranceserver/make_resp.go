package entranceserver

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/server/channelserver"
)

// Server Entries
var season uint8

// Server Channels
var currentplayers uint16

func encodeServerInfo(config *config.Config, s *Server, local bool) []byte {
	serverInfos := config.Entrance.Entries
	bf := byteframe.NewByteFrame()

	for serverIdx, si := range serverInfos {
		sid := (4096 + serverIdx*256) + 16
		err := s.db.QueryRow("SELECT season FROM servers WHERE server_id=$1", sid).Scan(&season)
		if err != nil {
			season = 0
		}
		if si.IP == "" {
			si.IP = config.Host
		}
		if local {
			bf.WriteUint32(0x0100007F) // 127.0.0.1
		} else {
			bf.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(si.IP).To4()))
		}
		bf.WriteUint16(16 + uint16(serverIdx))
		bf.WriteUint16(0x0000)
		bf.WriteUint16(uint16(len(si.Channels)))
		bf.WriteUint8(si.Type)
		bf.WriteUint8(season)
		bf.WriteUint8(si.Recommended)
		bf.WriteUint8(0) // Prevents malformed server name
		combined := append(stringsupport.UTF8ToSJIS(si.Name), []byte{0x00}...)
		combined = append(combined, stringsupport.UTF8ToSJIS(si.Description)...)
		bf.WriteBytes(stringsupport.PaddedString(string(combined), 65, false))
		bf.WriteUint32(si.AllowedClientFlags)

		for channelIdx, ci := range si.Channels {
			sid = (4096 + serverIdx*256) + (16 + channelIdx)
			bf.WriteUint16(ci.Port)
			bf.WriteUint16(16 + uint16(channelIdx))
			bf.WriteUint16(ci.MaxPlayers)
			err := s.db.QueryRow("SELECT current_players FROM servers WHERE server_id=$1", sid).Scan(&currentplayers)
			if err != nil {
				currentplayers = 0
			}
			bf.WriteUint16(currentplayers)
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint16(319) // Unk
			bf.WriteUint16(252) // Unk
			bf.WriteUint16(248) // Unk
			bf.WriteUint16(0x3039)
		}
	}
	bf.WriteUint32(uint32(channelserver.TimeAdjusted().Unix()))
	bf.WriteUint32(uint32(s.erupeConfig.GameplayOptions.ClanMemberLimits[len(s.erupeConfig.GameplayOptions.ClanMemberLimits)-1][1]))
	return bf.Data()
}

func makeHeader(data []byte, respType string, entryCount uint16, key byte) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteBytes([]byte(respType))
	bf.WriteUint16(entryCount)
	bf.WriteUint16(uint16(len(data)))
	if len(data) > 0 {
		bf.WriteUint32(CalcSum32(data))
		bf.WriteBytes(data)
	}

	dataToEncrypt := bf.Data()

	bf = byteframe.NewByteFrame()
	bf.WriteUint8(key)
	bf.WriteBytes(EncryptBin8(dataToEncrypt, key))
	return bf.Data()
}

func makeSv2Resp(config *config.Config, s *Server, local bool) []byte {
	serverInfos := config.Entrance.Entries
	rawServerData := encodeServerInfo(config, s, local)

	if s.erupeConfig.DevMode && s.erupeConfig.DevModeOptions.LogOutboundMessages {
		fmt.Printf("[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(rawServerData), hex.Dump(rawServerData))
	}

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(makeHeader(rawServerData, "SV2", uint16(len(serverInfos)), 0x00))
	return bf.Data()
}

func makeUsrResp(pkt []byte, s *Server) []byte {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	_ = bf.ReadUint32() // ALL+
	_ = bf.ReadUint8()  // 0x00
	userEntries := bf.ReadUint16()
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(userEntries); i++ {
		cid := bf.ReadUint32()
		var sid uint16
		err := s.db.QueryRow("SELECT(SELECT server_id FROM sign_sessions WHERE char_id=$1) AS _", cid).Scan(&sid)
		if err != nil {
			resp.WriteBytes(make([]byte, 4))
		} else {
			resp.WriteUint16(sid)
			resp.WriteUint16(0)
		}
	}

	if s.erupeConfig.DevMode && s.erupeConfig.DevModeOptions.LogOutboundMessages {
		fmt.Printf("[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(resp.Data()), hex.Dump(resp.Data()))
	}

	return makeHeader(resp.Data(), "USR", userEntries, 0x00)
}
