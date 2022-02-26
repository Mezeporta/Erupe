package signserver

import (
	"fmt"

	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

func paddedString(x string, size uint) []byte {
	out := make([]byte, size)
	copy(out, x)

	// Null terminate it.
	out[len(out)-1] = 0
	return out
}

func uint8PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func uint16PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func makeSignInFailureResp(respID RespID) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(respID))
	return bf.Data()
}

func (s *Session) makeSignInResp(uid int) []byte {
	// Get the characters from the DB.
	chars, err := s.server.getCharactersForUser(uid)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint8(1)                                   // resp_code
	bf.WriteUint8(0)                                   // file/patch server count
	bf.WriteUint8(4)                                   // entrance server count
	bf.WriteUint8(uint8(len(chars)))                   // character count
	bf.WriteUint32(0xFFFFFFFF)                         // login_token_number
	bf.WriteBytes(paddedString("logintokenstrng", 16)) // login_token (16 byte padded string)
	bf.WriteUint32(1576761190)
	uint8PascalString(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.HostIP, s.server.erupeConfig.Entrance.Port))
	uint8PascalString(bf, "")
	uint8PascalString(bf, "")
	uint8PascalString(bf, "mhf-n.capcom.com.tw")

	for _, char := range chars {
		bf.WriteUint32(char.ID) // character ID 469153291

		// Exp, HR[x] is split by 0, 1, 30, 50, 99, 299, 998, 999
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.Exp)
		}

		bf.WriteUint16(char.Weapon)                         // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                      // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                         // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                   // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(char.SmallGRLevel)                    // GR level if grMode == 0
		bf.WriteBool(char.GROverrideMode)                   // GR mode.
		bf.WriteBytes(paddedString(char.Name, 16))          // Character name
		bf.WriteBytes(paddedString(char.UnkDescString, 32)) // unk str
		if char.GROverrideMode {
			bf.SetLE()
			bf.WriteUint16(char.GROverrideLevel) // GR level override.
			bf.SetBE()
			bf.WriteUint8(char.GROverrideUnk0)   // unk
			bf.WriteUint8(char.GROverrideUnk1)   // unk
		}
	}

	bf.WriteUint8(0)           // friends_list_count
	bf.WriteUint8(0)           // guild_members_count
	bf.WriteUint8(0)           // notice_count
	bf.WriteUint32(0xDEADBEEF) // some_last_played_character_id
	bf.WriteUint32(14)         // unk_flags
	uint8PascalString(bf, "")  // unk_data_blob PascalString

	bf.WriteUint16(51728)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "1000672925")

	bf.WriteUint8(0)

	bf.WriteUint16(51729)
	bf.WriteUint16(1)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "203.191.249.36:8080")

	bf.WriteUint32(1578905116)
	bf.WriteUint32(0)

	return bf.Data()
}
