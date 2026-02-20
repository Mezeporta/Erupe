package signserver

import (
	"database/sql"
	"encoding/hex"
	"erupe-ce/common/stringsupport"
	"net"
	"strings"
	"sync"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"

	"go.uber.org/zap"
)

type client int

const (
	PC100 client = iota
	VITA
	PS3
	PS4
	WIIU
)

// Session holds state for the sign server connection.
type Session struct {
	sync.Mutex
	logger    *zap.Logger
	server    *Server
	rawConn   net.Conn
	cryptConn *network.CryptConn
	client    client
	psn       string
}

func (s *Session) work() {
	pkt, err := s.cryptConn.ReadPacket()

	if s.server.erupeConfig.DebugOptions.LogInboundMessages {
		s.logger.Debug("Inbound packet", zap.Int("bytes", len(pkt)), zap.String("data", hex.Dump(pkt)))
	}

	if err != nil {
		return
	}
	err = s.handlePacket(pkt)
	if err != nil {
		return
	}
}

func (s *Session) handlePacket(pkt []byte) error {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	reqType := string(bf.ReadNullTerminatedBytes())
	switch reqType[:len(reqType)-3] {
	case "DLTSKEYSIGN:", "DSGN:", "SIGN:":
		s.handleDSGN(bf)
	case "PS4SGN:":
		s.client = PS4
		s.handlePSSGN(bf)
	case "PS3SGN:":
		s.client = PS3
		s.handlePSSGN(bf)
	case "VITASGN:":
		s.client = VITA
		s.handlePSSGN(bf)
	case "WIIUSGN:":
		s.client = WIIU
		s.handleWIIUSGN(bf)
	case "VITACOGLNK:", "COGLNK:":
		s.handlePSNLink(bf)
	case "DELETE:":
		token := string(bf.ReadNullTerminatedBytes())
		characterID := int(bf.ReadUint32())
		tokenID := bf.ReadUint32()
		err := s.server.deleteCharacter(characterID, token, tokenID)
		if err == nil {
			s.logger.Info("Deleted character", zap.Int("CharacterID", characterID))
			_ = s.cryptConn.SendPacket([]byte{0x01}) // DEL_SUCCESS
		}
	default:
		s.logger.Warn("Unknown request", zap.String("reqType", reqType))
		if s.server.erupeConfig.DebugOptions.LogInboundMessages {
			s.logger.Debug("Unknown inbound packet", zap.Int("bytes", len(pkt)), zap.String("data", hex.Dump(pkt)))
		}
	}
	return nil
}

func (s *Session) authenticate(username string, password string) {
	newCharaReq := false
	if username[len(username)-1] == 43 { // '+'
		username = username[:len(username)-1]
		newCharaReq = true
	}
	bf := byteframe.NewByteFrame()
	uid, resp := s.server.validateLogin(username, password)
	switch resp {
	case SIGN_SUCCESS:
		if newCharaReq {
			_ = s.server.newUserChara(uid)
		}
		bf.WriteBytes(s.makeSignResponse(uid))
	default:
		bf.WriteUint8(uint8(resp))
	}
	if s.server.erupeConfig.DebugOptions.LogOutboundMessages {
		s.logger.Debug("Outbound packet", zap.Int("bytes", len(bf.Data())), zap.String("data", hex.Dump(bf.Data())))
	}
	_ = s.cryptConn.SendPacket(bf.Data())
}

func (s *Session) handleWIIUSGN(bf *byteframe.ByteFrame) {
	_ = bf.ReadBytes(1)
	wiiuKey := string(bf.ReadBytes(64))
	var uid uint32
	err := s.server.db.QueryRow(`SELECT id FROM users WHERE wiiu_key = $1`, wiiuKey).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Unlinked Wii U attempted to authenticate", zap.String("Key", wiiuKey))
			s.sendCode(SIGN_ECOGLINK)
			return
		}
		s.sendCode(SIGN_EABORT)
		return
	}
	_ = s.cryptConn.SendPacket(s.makeSignResponse(uid))
}

func (s *Session) handlePSSGN(bf *byteframe.ByteFrame) {
	// Prevent reading malformed request
	if s.client != PS4 {
		if len(bf.DataFromCurrent()) < 128 {
			s.sendCode(SIGN_EABORT)
			return
		}

		_ = bf.ReadNullTerminatedBytes() // VITA = 0000000256, PS3 = 0000000255
		_ = bf.ReadBytes(2)              // VITA = 1, PS3 = !
		_ = bf.ReadBytes(82)
	}
	s.psn = string(bf.ReadNullTerminatedBytes())
	var uid uint32
	err := s.server.db.QueryRow(`SELECT id FROM users WHERE psn_id = $1`, s.psn).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = s.cryptConn.SendPacket(s.makeSignResponse(0))
			return
		}
		s.sendCode(SIGN_EABORT)
		return
	}
	_ = s.cryptConn.SendPacket(s.makeSignResponse(uid))
}

func (s *Session) handlePSNLink(bf *byteframe.ByteFrame) {
	_ = bf.ReadNullTerminatedBytes() // Client ID
	credStr, _ := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	credentials := strings.Split(credStr, "\n")
	token := string(bf.ReadNullTerminatedBytes())
	uid, resp := s.server.validateLogin(credentials[0], credentials[1])
	if resp == SIGN_SUCCESS && uid > 0 {
		var psn string
		err := s.server.db.QueryRow(`SELECT psn_id FROM sign_sessions WHERE token = $1`, token).Scan(&psn)
		if err != nil {
			s.sendCode(SIGN_ECOGLINK)
			return
		}

		// Since we check for the psn_id, this will never run
		var exists int
		err = s.server.db.QueryRow(`SELECT count(*) FROM users WHERE psn_id = $1`, psn).Scan(&exists)
		if err != nil {
			s.sendCode(SIGN_ECOGLINK)
			return
		} else if exists > 0 {
			s.sendCode(SIGN_EPSI)
			return
		}

		var currentPSN string
		err = s.server.db.QueryRow(`SELECT COALESCE(psn_id, '') FROM users WHERE username = $1`, credentials[0]).Scan(&currentPSN)
		if err != nil {
			s.sendCode(SIGN_ECOGLINK)
			return
		} else if currentPSN != "" {
			s.sendCode(SIGN_EMBID)
			return
		}

		_, err = s.server.db.Exec(`UPDATE users SET psn_id = $1 WHERE username = $2`, psn, credentials[0])
		if err == nil {
			s.sendCode(SIGN_SUCCESS)
			return
		}
	}
	s.sendCode(SIGN_ECOGLINK)
}

func (s *Session) handleDSGN(bf *byteframe.ByteFrame) {
	user, _ := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	pass, _ := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	_ = string(bf.ReadNullTerminatedBytes()) // Unk
	s.authenticate(user, pass)
}

func (s *Session) sendCode(id RespID) {
	_ = s.cryptConn.SendPacket([]byte{byte(id)})
}
