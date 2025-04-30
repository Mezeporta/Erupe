package signserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/server/channelserver"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

func (s *Session) makeSignResponse(uid uint32) []byte {
	// Get the characters from the DB.
	chars, err := s.server.getCharactersForUser(uid)
	if len(chars) == 0 && uid != 0 {
		err = s.server.newUserChara(uid)
		if err == nil {
			chars, err = s.server.getCharactersForUser(uid)
		}
	}
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()
	var tokenID uint32
	var sessToken string
	if uid == 0 && s.psn != "" {
		tokenID, sessToken, err = s.server.registerPsnToken(s.psn)
	} else {
		tokenID, sessToken, err = s.server.registerUidToken(uid)
	}
	if err != nil {
		bf.WriteUint8(uint8(SIGN_EABORT))
		return bf.Data()
	}

	if s.client == PS3 && (s.server.erupeConfig.PatchServerFile == "" || s.server.erupeConfig.PatchServerManifest == "") {
		bf.WriteUint8(uint8(SIGN_EABORT))
		return bf.Data()
	}

	bf.WriteUint8(uint8(SIGN_SUCCESS))
	bf.WriteUint8(2) // patch server count
	bf.WriteUint8(1) // entrance server count
	bf.WriteUint8(uint8(len(chars)))
	bf.WriteUint32(tokenID)
	bf.WriteBytes([]byte(sessToken))
	bf.WriteUint32(uint32(channelserver.TimeAdjusted().Unix()))
	if s.client == PS3 {
		ps.Uint8(bf, fmt.Sprintf("%s/ps3", s.server.erupeConfig.PatchServerManifest), false)
		ps.Uint8(bf, fmt.Sprintf("%s/ps3", s.server.erupeConfig.PatchServerFile), false)
	} else {
		ps.Uint8(bf, s.server.erupeConfig.PatchServerManifest, false)
		ps.Uint8(bf, s.server.erupeConfig.PatchServerFile, false)
	}
	if strings.Split(s.rawConn.RemoteAddr().String(), ":")[0] == "127.0.0.1" {
		ps.Uint8(bf, fmt.Sprintf("127.0.0.1:%d", s.server.erupeConfig.Entrance.Port), false)
	} else {
		ps.Uint8(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.Host, s.server.erupeConfig.Entrance.Port), false)
	}

	lastPlayed := uint32(0)
	for _, char := range chars {
		if lastPlayed == 0 {
			lastPlayed = char.ID
		}
		bf.WriteUint32(char.ID)
		if s.server.erupeConfig.DebugOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.HR)
		}
		bf.WriteUint16(char.WeaponType)                                          // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                                           // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                                              // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                                        // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(0)                                                         // Old GR
		bf.WriteBool(true)                                                       // Use uint16 GR, no reason not to
		bf.WriteBytes(stringsupport.PaddedString(char.Name, 16, true))           // Character name
		bf.WriteBytes(stringsupport.PaddedString(char.UnkDescString, 32, false)) // unk str
		if s.server.erupeConfig.RealClientMode >= _config.G7 {
			bf.WriteUint16(char.GR)
			bf.WriteUint8(0) // Unk
			bf.WriteUint8(0) // Unk
		}
	}

	friends := s.server.getFriendsForCharacters(chars)
	if len(friends) == 0 {
		bf.WriteUint8(0)
	} else {
		if len(friends) > 255 {
			bf.WriteUint8(255)
			bf.WriteUint16(uint16(len(friends)))
		} else {
			bf.WriteUint8(uint8(len(friends)))
		}
		for _, friend := range friends {
			bf.WriteUint32(friend.CID)
			bf.WriteUint32(friend.ID)
			ps.Uint8(bf, friend.Name, true)
		}
	}

	guildmates := s.server.getGuildmatesForCharacters(chars)
	if len(guildmates) == 0 {
		bf.WriteUint8(0)
	} else {
		if len(guildmates) > 255 {
			bf.WriteUint8(255)
			bf.WriteUint16(uint16(len(guildmates)))
		} else {
			bf.WriteUint8(uint8(len(guildmates)))
		}
		for _, guildmate := range guildmates {
			bf.WriteUint32(guildmate.CID)
			bf.WriteUint32(guildmate.ID)
			ps.Uint8(bf, guildmate.Name, true)
		}
	}

	if s.server.erupeConfig.HideLoginNotice {
		bf.WriteBool(false)
	} else {
		bf.WriteBool(true)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		ps.Uint16(bf, strings.Join(s.server.erupeConfig.LoginNotices[:], "<PAGE>"), true)
	}

	bf.WriteUint32(s.server.getLastCID(uid))
	bf.WriteUint32(s.server.getUserRights(uid))

	namNGWords := []string{}
	msgNGWords := []string{}

	filters := byteframe.NewByteFrame()
	filters.SetLE()
	filters.WriteNullTerminatedBytes([]byte("smc"))
	smc := byteframe.NewByteFrame()
	smc.SetLE()
	smcData := []struct {
		charGroup [][]rune
	}{
		{[][]rune{{'='}, {'＝'}}},
		{[][]rune{{')'}, {'）'}}},
		{[][]rune{{'('}, {'（'}}},
		{[][]rune{{'!'}, {'！'}}},
		{[][]rune{{'/'}, {'／'}}},
		{[][]rune{{'+'}, {'＋'}}},
		{[][]rune{{'&'}, {'＆'}}},
		{[][]rune{{'ぼ'}, {'ボ'}, {'ﾎ', 'ﾞ'}, {'ほ', 'ﾞ'}, {'ホ', 'ﾞ'}, {'ほ', '゛'}, {'ホ', '゛'}, {'ﾎ', '゛'}}},
		{[][]rune{{'べ'}, {'ベ'}, {'ﾍ', 'ﾞ'}, {'へ', 'ﾞ'}, {'ヘ', 'ﾞ'}, {'へ', '゛'}, {'ﾍ', '゛'}, {'ヘ', '゛'}}},
		{[][]rune{{'で'}, {'デ'}, {'ﾃ', 'ﾞ'}, {'て', 'ﾞ'}, {'テ', 'ﾞ'}, {'て', '゛'}, {'テ', '゛'}, {'ﾃ', '゛'}, {'〒', '゛'}, {'〒', 'ﾞ'}, {'乙', 'ﾞ'}, {'乙', '゛'}}},
		{[][]rune{{'び'}, {'ビ'}, {'ﾋ', 'ﾞ'}, {'ひ', 'ﾞ'}, {'ヒ', 'ﾞ'}, {'ひ', '゛'}, {'ヒ', '゛'}, {'ﾋ', '゛'}}},
		{[][]rune{{'ど'}, {'ド'}, {'ﾄ', 'ﾞ'}, {'と', 'ﾞ'}, {'ト', 'ﾞ'}, {'と', '゛'}, {'ト', '゛'}, {'ﾄ', '゛'}, {'┣', 'ﾞ'}, {'┣', '゛'}, {'├', 'ﾞ'}, {'├', '゛'}}},
		{[][]rune{{'ば'}, {'バ'}, {'ﾊ', 'ﾞ'}, {'は', 'ﾞ'}, {'ハ', 'ﾞ'}, {'八', 'ﾞ'}, {'は', '゛'}, {'ハ', '゛'}, {'ﾊ', '゛'}, {'八', '゛'}}},
		{[][]rune{{'つ', 'ﾞ'}, {'ヅ'}, {'ツ', 'ﾞ'}, {'つ', '゛'}, {'ツ', '゛'}, {'ﾂ', 'ﾞ'}, {'ﾂ', '゛'}, {'づ'}, {'っ', 'ﾞ'}, {'ッ', 'ﾞ'}, {'ｯ', 'ﾞ'}, {'っ', '゛'}, {'ッ', '゛'}, {'ｯ', '゛'}}},
		{[][]rune{{'ぶ'}, {'ブ'}, {'ﾌ', 'ﾞ'}, {'ヴ'}, {'ｳ', 'ﾞ'}, {'う', '゛'}, {'う', 'ﾞ'}, {'ウ', 'ﾞ'}, {'ｩ', 'ﾞ'}, {'ぅ', 'ﾞ'}, {'ふ', 'ﾞ'}, {'フ', 'ﾞ'}, {'ﾌ', '゛'}}},
		{[][]rune{{'ぢ'}, {'ヂ'}, {'ﾁ', 'ﾞ'}, {'ち', 'ﾞ'}, {'チ', 'ﾞ'}, {'ち', '゛'}, {'チ', '゛'}, {'ﾁ', '゛'}, {'千', '゛'}, {'千', 'ﾞ'}}},
		{[][]rune{{'だ'}, {'ダ'}, {'ﾀ', 'ﾞ'}, {'た', 'ﾞ'}, {'タ', 'ﾞ'}, {'夕', 'ﾞ'}, {'た', '゛'}, {'タ', '゛'}, {'ﾀ', '゛'}, {'夕', '゛'}}},
		{[][]rune{{'ぞ'}, {'ゾ'}, {'ｿ', 'ﾞ'}, {'そ', 'ﾞ'}, {'ソ', 'ﾞ'}, {'そ', '゛'}, {'ソ', '゛'}, {'ｿ', '゛'}, {'ン', 'ﾞ'}, {'ン', '゛'}, {'ﾝ', '゛'}, {'ﾝ', 'ﾞ'}, {'リ', 'ﾞ'}, {'ﾘ', 'ﾞ'}, {'ﾘ', '゛'}, {'リ', '゛'}}},
		{[][]rune{{'ぜ'}, {'ｾ', 'ﾞ'}, {'せ', 'ﾞ'}, {'セ', 'ﾞ'}, {'せ', '゛'}, {'セ', '゛'}, {'ｾ', '゛'}, {'ゼ'}}},
		{[][]rune{{'ず'}, {'ズ'}, {'ｽ', 'ﾞ'}, {'す', 'ﾞ'}, {'ス', 'ﾞ'}, {'す', '゛'}, {'ス', '゛'}, {'ｽ', '゛'}}},
		{[][]rune{{'じ'}, {'ジ'}, {'ｼ', 'ﾞ'}, {'し', 'ﾞ'}, {'シ', 'ﾞ'}, {'し', '゛'}, {'シ', '゛'}, {'ｼ', '゛'}}},
		{[][]rune{{'ざ'}, {'ザ'}, {'ｻ', 'ﾞ'}, {'さ', 'ﾞ'}, {'サ', 'ﾞ'}, {'さ', '゛'}, {'サ', '゛'}, {'ｻ', '゛'}}},
		{[][]rune{{'ご'}, {'ゴ'}, {'ｺ', 'ﾞ'}, {'こ', 'ﾞ'}, {'コ', 'ﾞ'}, {'こ', '゛'}, {'コ', '゛'}, {'ｺ', '゛'}}},
		{[][]rune{{'げ'}, {'ゲ'}, {'ｹ', 'ﾞ'}, {'け', 'ﾞ'}, {'ケ', 'ﾞ'}, {'け', '゛'}, {'ケ', '゛'}, {'ｹ', '゛'}, {'ヶ', 'ﾞ'}, {'ヶ', '゛'}}},
		{[][]rune{{'ぐ'}, {'グ'}, {'ｸ', 'ﾞ'}, {'く', 'ﾞ'}, {'ク', 'ﾞ'}, {'く', '゛'}, {'ク', '゛'}, {'ｸ', '゛'}}},
		{[][]rune{{'ぎ'}, {'ギ'}, {'ｷ', 'ﾞ'}, {'き', 'ﾞ'}, {'キ', 'ﾞ'}, {'き', '゛'}, {'キ', '゛'}, {'ｷ', '゛'}}},
		{[][]rune{{'が'}, {'ガ'}, {'ｶ', 'ﾞ'}, {'ヵ', 'ﾞ'}, {'カ', 'ﾞ'}, {'か', 'ﾞ'}, {'力', 'ﾞ'}, {'ヵ', '゛'}, {'カ', '゛'}, {'か', '゛'}, {'力', '゛'}, {'ｶ', '゛'}}},
		{[][]rune{{'を'}, {'ヲ'}, {'ｦ'}}},
		{[][]rune{{'わ'}, {'ワ'}, {'ﾜ'}, {'ヮ'}}},
		{[][]rune{{'ろ'}, {'ロ'}, {'ﾛ'}, {'□'}, {'口'}}},
		{[][]rune{{'れ'}, {'レ'}, {'ﾚ'}}},
		{[][]rune{{'る'}, {'ル'}, {'ﾙ'}}},
		{[][]rune{{'り'}, {'リ'}, {'ﾘ'}}},
		{[][]rune{{'ら'}, {'ラ'}, {'ﾗ'}}},
		{[][]rune{{'よ'}, {'ヨ'}, {'ﾖ'}, {'ｮ'}, {'ょ'}, {'ョ'}}},
		{[][]rune{{'ゆ'}, {'ユ'}, {'ﾕ'}, {'ｭ'}, {'ゅ'}, {'ュ'}}},
		{[][]rune{{'や'}, {'ヤ'}, {'ﾔ'}, {'ｬ'}, {'ゃ'}, {'ャ'}}},
		{[][]rune{{'も'}, {'モ'}, {'ﾓ'}}},
		{[][]rune{{'め'}, {'メ'}, {'ﾒ'}, {'M', 'E'}}},
		{[][]rune{{'む'}, {'ム'}, {'ﾑ'}}},
		{[][]rune{{'み'}, {'ミ'}, {'ﾐ'}}},
		{[][]rune{{'ま'}, {'マ'}, {'ﾏ'}}},
		{[][]rune{{'ほ'}, {'ホ'}, {'ﾎ'}}},
		{[][]rune{{'へ'}, {'ヘ'}, {'ﾍ'}}},
		{[][]rune{{'ふ'}, {'フ'}, {'ﾌ'}}},
		{[][]rune{{'ひ'}, {'ヒ'}, {'ﾋ'}}},
		{[][]rune{{'は'}, {'ハ'}, {'ﾊ'}, {'八'}}},
		{[][]rune{{'の'}, {'ノ'}, {'ﾉ'}}},
		{[][]rune{{'ね'}, {'ネ'}, {'ﾈ'}}},
		{[][]rune{{'ぬ'}, {'ヌ'}, {'ﾇ'}}},
		{[][]rune{{'に'}, {'ニ'}, {'ﾆ'}, {'二'}}},
		{[][]rune{{'な'}, {'ナ'}, {'ﾅ'}}},
		{[][]rune{{'と'}, {'ト'}, {'ﾄ'}, {'┣'}, {'├'}}},
		{[][]rune{{'て'}, {'テ'}, {'ﾃ'}, {'〒'}, {'乙'}}},
		{[][]rune{{'つ'}, {'ツ'}, {'ﾂ'}, {'っ'}, {'ッ'}, {'ｯ'}}},
		{[][]rune{{'ち'}, {'チ'}, {'ﾁ'}, {'千'}}},
		{[][]rune{{'た'}, {'タ'}, {'ﾀ'}, {'夕'}}},
		{[][]rune{{'そ'}, {'ソ'}, {'ｿ'}}},
		{[][]rune{{'せ'}, {'セ'}, {'ｾ'}}},
		{[][]rune{{'す'}, {'ス'}, {'ｽ'}}},
		{[][]rune{{'し'}, {'シ'}, {'ｼ'}}},
		{[][]rune{{'さ'}, {'サ'}, {'ｻ'}}},
		{[][]rune{{'こ'}, {'コ'}, {'ｺ'}}},
		{[][]rune{{'け'}, {'ケ'}, {'ｹ'}, {'ヶ'}}},
		{[][]rune{{'く'}, {'ク'}, {'ｸ'}}},
		{[][]rune{{'き'}, {'キ'}, {'ｷ'}}},
		{[][]rune{{'か'}, {'カ'}, {'ｶ'}, {'ヵ'}, {'力'}}},
		{[][]rune{{'お'}, {'オ'}, {'ｵ'}, {'ｫ'}, {'ぉ'}, {'ォ'}}},
		{[][]rune{{'え'}, {'エ'}, {'ｴ'}, {'ｪ'}, {'ぇ'}, {'ェ'}, {'工'}}},
		{[][]rune{{'う'}, {'ウ'}, {'ｳ'}, {'ｩ'}, {'ぅ'}, {'ゥ'}}},
		{[][]rune{{'い'}, {'イ'}, {'ｲ'}, {'ｨ'}, {'ぃ'}, {'ィ'}}},
		{[][]rune{{'あ'}, {'ア'}, {'ｧ'}, {'ｱ'}, {'ぁ'}, {'ァ'}}},
		{[][]rune{{'ー'}, {'―'}, {'‐'}, {'-'}, {'－'}, {'ｰ'}, {'一'}}},
		{[][]rune{{'9'}, {'９'}}},
		{[][]rune{{'8'}, {'８'}}},
		{[][]rune{{'7'}, {'７'}}},
		{[][]rune{{'6'}, {'６'}}},
		{[][]rune{{'5'}, {'５'}}},
		{[][]rune{{'4'}, {'４'}}},
		{[][]rune{{'3'}, {'３'}}},
		{[][]rune{{'2'}, {'２'}}},
		{[][]rune{{'1'}, {'１'}}},
		{[][]rune{{'ぽ'}, {'ポ'}, {'ﾎ', 'ﾟ'}, {'ほ', 'ﾟ'}, {'ホ', 'ﾟ'}, {'ホ', '°'}, {'ほ', '°'}, {'ﾎ', '°'}}},
		{[][]rune{{'ぺ'}, {'ペ'}, {'ﾍ', 'ﾟ'}, {'へ', 'ﾟ'}, {'ヘ', 'ﾟ'}, {'ヘ', '°'}, {'へ', '°'}, {'ﾍ', '°'}}},
		{[][]rune{{'ぷ'}, {'プ'}, {'ﾌ', 'ﾟ'}, {'ふ', 'ﾟ'}, {'フ', 'ﾟ'}, {'フ', '°'}, {'ふ', '°'}, {'ﾌ', '°'}}},
		{[][]rune{{'ぴ'}, {'ピ'}, {'ﾋ', 'ﾟ'}, {'ひ', 'ﾟ'}, {'ヒ', 'ﾟ'}, {'ヒ', '°'}, {'ひ', '°'}, {'ﾋ', '°'}}},
		{[][]rune{{'ぱ'}, {'パ'}, {'ﾊ', 'ﾟ'}, {'は', 'ﾟ'}, {'ハ', 'ﾟ'}, {'ハ', '°'}, {'は', '°'}, {'ﾊ', '°'}, {'八', 'ﾟ'}, {'八', '゜'}}},
		{[][]rune{{'z'}, {'ｚ'}, {'Z'}, {'Ｚ'}, {'Ζ'}}},
		{[][]rune{{'y'}, {'ｙ'}, {'Y'}, {'Ｙ'}, {'Υ'}, {'У'}, {'у'}}},
		{[][]rune{{'x'}, {'ｘ'}, {'X'}, {'Ｘ'}, {'Χ'}, {'χ'}, {'Х'}, {'×'}, {'х'}}},
		{[][]rune{{'w'}, {'ｗ'}, {'W'}, {'Ｗ'}, {'ω'}, {'Ш'}, {'ш'}, {'щ'}}},
		{[][]rune{{'v'}, {'ｖ'}, {'V'}, {'Ｖ'}, {'ν'}, {'υ'}}},
		{[][]rune{{'u'}, {'ｕ'}, {'U'}, {'Ｕ'}, {'μ'}, {'∪'}}},
		{[][]rune{{'t'}, {'ｔ'}, {'T'}, {'Ｔ'}, {'Τ'}, {'τ'}, {'Т'}, {'т'}}},
		{[][]rune{{'s'}, {'ｓ'}, {'S'}, {'Ｓ'}, {'∫'}, {'＄'}, {'$'}}},
		{[][]rune{{'r'}, {'ｒ'}, {'R'}, {'Ｒ'}, {'Я'}, {'я'}}},
		{[][]rune{{'q'}, {'ｑ'}, {'Q'}, {'Ｑ'}}},
		{[][]rune{{'p'}, {'ｐ'}, {'P'}, {'Ｐ'}, {'Ρ'}, {'ρ'}, {'Р'}, {'р'}}},
		{[][]rune{{'o'}, {'ｏ'}, {'O'}, {'Ｏ'}, {'○'}, {'Ο'}, {'ο'}, {'О'}, {'о'}, {'◯'}, {'〇'}, {'0'}, {'０'}}},
		{[][]rune{{'n'}, {'ｎ'}, {'N'}, {'Ｎ'}, {'Ν'}, {'η'}, {'ﾝ'}, {'ん'}, {'ン'}}},
		{[][]rune{{'m'}, {'ｍ'}, {'M'}, {'Ｍ'}, {'Μ'}, {'М'}, {'м'}}},
		{[][]rune{{'l'}, {'ｌ'}, {'L'}, {'Ｌ'}, {'|'}}},
		{[][]rune{{'k'}, {'ｋ'}, {'K'}, {'Ｋ'}, {'Κ'}, {'κ'}, {'К'}, {'к'}}},
		{[][]rune{{'j'}, {'ｊ'}, {'J'}, {'Ｊ'}}},
		{[][]rune{{'i'}, {'ｉ'}, {'I'}, {'Ｉ'}, {'Ι'}}},
		{[][]rune{{'h'}, {'ｈ'}, {'H'}, {'Ｈ'}, {'Η'}, {'Н'}, {'н'}}},
		{[][]rune{{'f'}, {'ｆ'}, {'F'}, {'Ｆ'}}},
		{[][]rune{{'g'}, {'ｇ'}, {'G'}, {'Ｇ'}}},
		{[][]rune{{'e'}, {'ｅ'}, {'E'}, {'Ｅ'}, {'Ε'}, {'ε'}, {'Е'}, {'Ё'}, {'е'}, {'ё'}, {'∈'}}},
		{[][]rune{{'d'}, {'ｄ'}, {'D'}, {'Ｄ'}}},
		{[][]rune{{'c'}, {'ｃ'}, {'C'}, {'С'}, {'с'}, {'Ｃ'}, {'℃'}}},
		{[][]rune{{'b'}, {'Ｂ'}, {'ｂ'}, {'B'}, {'β'}, {'Β'}, {'В'}, {'в'}, {'ъ'}, {'ь'}, {'♭'}}},
		{[][]rune{{'\''}, {'’'}}},
		{[][]rune{{'a'}, {'Ａ'}, {'ａ'}, {'A'}, {'α'}, {'@'}, {'＠'}, {'а'}, {'Å'}, {'А'}, {'Α'}}},
		{[][]rune{{'"'}, {'”'}}},
		{[][]rune{{'%'}, {'％'}}},
	}
	for _, smcGroup := range smcData {
		for _, smcPair := range smcGroup.charGroup {
			smc.WriteUint16(stringsupport.ToNGWord(string(smcPair[0]))[0])
			if len(smcPair) > 1 {
				smc.WriteUint16(stringsupport.ToNGWord(string(smcPair[1]))[0])
			} else {
				smc.WriteUint16(0)
			}
		}
		smc.WriteUint32(0)
	}

	filters.WriteUint32(uint32(len(smc.Data())))
	filters.WriteBytes(smc.Data())

	filters.WriteNullTerminatedBytes([]byte("nam"))
	nam := byteframe.NewByteFrame()
	nam.SetLE()
	for _, word := range namNGWords {
		parts := stringsupport.ToNGWord(word)
		nam.WriteUint32(uint32(len(parts)))
		for _, part := range parts {
			nam.WriteUint16(part)
			var i int16
			j := int16(-1)
			for _, smcGroup := range smcData {
				if rune(part) == rune(stringsupport.ToNGWord(string(smcGroup.charGroup[0][0]))[0]) {
					j = i
					break
				}
				i += int16(len(smcGroup.charGroup) + 1)
			}
			nam.WriteInt16(j)
		}
		nam.WriteUint16(0)
		nam.WriteInt16(-1)
	}
	filters.WriteUint32(uint32(len(nam.Data())))
	filters.WriteBytes(nam.Data())

	filters.WriteNullTerminatedBytes([]byte("msg"))
	msg := byteframe.NewByteFrame()
	msg.SetLE()
	for _, word := range msgNGWords {
		parts := stringsupport.ToNGWord(word)
		msg.WriteUint32(uint32(len(parts)))
		for _, part := range parts {
			msg.WriteUint16(part)
			var i int16
			j := int16(-1)
			for _, smcGroup := range smcData {
				if rune(part) == rune(stringsupport.ToNGWord(string(smcGroup.charGroup[0][0]))[0]) {
					j = i
					break
				}
				i += int16(len(smcGroup.charGroup) + 1)
			}
			msg.WriteInt16(j)
		}
		msg.WriteUint16(0)
		msg.WriteInt16(-1)
	}
	filters.WriteUint32(uint32(len(msg.Data())))
	filters.WriteBytes(msg.Data())

	bf.WriteUint16(uint16(len(filters.Data())))
	bf.WriteBytes(filters.Data())

	if s.client == VITA || s.client == PS3 || s.client == PS4 {
		var psnUser string
		s.server.db.QueryRow("SELECT psn_id FROM users WHERE id = $1", uid).Scan(&psnUser)
		bf.WriteBytes(stringsupport.PaddedString(psnUser, 20, true))
	}

	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[0])
	if s.server.erupeConfig.DebugOptions.CapLink.Values[0] == 51728 {
		bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[1])
		if s.server.erupeConfig.DebugOptions.CapLink.Values[1] == 20000 || s.server.erupeConfig.DebugOptions.CapLink.Values[1] == 20002 {
			ps.Uint16(bf, s.server.erupeConfig.DebugOptions.CapLink.Key, false)
		}
	}
	caStruct := []struct {
		Unk0 uint8
		Unk1 uint32
		Unk2 string
	}{}
	bf.WriteUint8(uint8(len(caStruct)))
	for i := range caStruct {
		bf.WriteUint8(caStruct[i].Unk0)
		bf.WriteUint32(caStruct[i].Unk1)
		ps.Uint8(bf, caStruct[i].Unk2, false)
	}
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[2])
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[3])
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[4])
	if s.server.erupeConfig.DebugOptions.CapLink.Values[2] == 51729 && s.server.erupeConfig.DebugOptions.CapLink.Values[3] == 1 && s.server.erupeConfig.DebugOptions.CapLink.Values[4] == 20000 {
		ps.Uint16(bf, fmt.Sprintf(`%s:%d`, s.server.erupeConfig.DebugOptions.CapLink.Host, s.server.erupeConfig.DebugOptions.CapLink.Port), false)
	}

	bf.WriteUint32(uint32(s.server.getReturnExpiry(uid).Unix()))
	bf.WriteUint32(0)

	tickets := []uint32{
		s.server.erupeConfig.GameplayOptions.MezFesSoloTickets,
		s.server.erupeConfig.GameplayOptions.MezFesGroupTickets,
	}
	stalls := []uint8{
		10, 3, 6, 9, 4, 8, 5, 7,
	}
	if s.server.erupeConfig.GameplayOptions.MezFesSwitchMinigame {
		stalls[4] = 2
	}

	// We can just use the start timestamp as the event ID
	bf.WriteUint32(uint32(channelserver.TimeWeekStart().Unix()))
	// Start time
	bf.WriteUint32(uint32(channelserver.TimeWeekNext().Add(-time.Duration(s.server.erupeConfig.GameplayOptions.MezFesDuration) * time.Second).Unix()))
	// End time
	bf.WriteUint32(uint32(channelserver.TimeWeekNext().Unix()))
	bf.WriteUint8(uint8(len(tickets)))
	for i := range tickets {
		bf.WriteUint32(tickets[i])
	}
	bf.WriteUint8(uint8(len(stalls)))
	for i := range stalls {
		bf.WriteUint8(stalls[i])
	}
	return bf.Data()
}
