package signserver

import (
	"database/sql"
	"encoding/hex"
	"net"
	"sync"

	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Session holds state for the sign server connection.
type Session struct {
	sync.Mutex
	logger    *zap.Logger
	sid       int
	server    *Server
	rawConn   *net.Conn
	cryptConn *network.CryptConn
}

func (s *Session) fail() {
	s.server.Lock()
	delete(s.server.sessions, s.sid)
	s.server.Unlock()

}

func (s *Session) work() {
	for {
		pkt, err := s.cryptConn.ReadPacket()
		if err != nil {
			s.fail()
			return
		}

		err = s.handlePacket(pkt)
		if err != nil {
			s.fail()
			return
		}
	}
}

func (s *Session) handlePacket(pkt []byte) error {
	sugar := s.logger.Sugar()

	bf := byteframe.NewByteFrameFromBytes(pkt)
	reqType := string(bf.ReadNullTerminatedBytes())
	switch reqType {
	case "DLTSKEYSIGN:100":
		fallthrough
	case "DSGN:100":
		err := s.handleDSGNRequest(bf)
		if err != nil {
			return nil
		}
	case "DELETE:100":
		loginTokenString := string(bf.ReadNullTerminatedBytes())
		_ = loginTokenString
		characterID := bf.ReadUint32()

		sugar.Infof("Got delete request for character ID: %v\n", characterID)
		sugar.Infof("remaining unknown data:\n%s\n", hex.Dump(bf.DataFromCurrent()))
	default:
		sugar.Infof("Got unknown request type %s, data:\n%s\n", reqType, hex.Dump(bf.DataFromCurrent()))
	}

	return nil
}

func (s *Session) handleDSGNRequest(bf *byteframe.ByteFrame) error {

	reqUsername := string(bf.ReadNullTerminatedBytes())
	reqPassword := string(bf.ReadNullTerminatedBytes())
	reqUnk := string(bf.ReadNullTerminatedBytes())

	s.server.logger.Info(
		"Got sign in request",
		zap.String("reqUsername", reqUsername),
		zap.String("reqPassword", reqPassword),
		zap.String("reqUnk", reqUnk),
	)

	// TODO(Andoryuuta): remove plaintext password storage if this ever becomes more than a toy project.
	var (
		id       int
		password string
	)
	err := s.server.db.QueryRow("SELECT id, password FROM users WHERE username = $1", reqUsername).Scan(&id, &password)
	var serverRespBytes []byte
	switch {
	case err == sql.ErrNoRows:
		s.logger.Info("Account not found", zap.String("reqUsername", reqUsername))
		serverRespBytes = makeSignInFailureResp(SIGN_EAUTH)

		// HACK(Andoryuuta): Create a new account if it doesn't exit.
		s.logger.Info("Creating account", zap.String("reqUsername", reqUsername), zap.String("reqPassword", reqPassword))
		err = s.server.registerDBAccount(reqUsername, reqPassword)
		if err != nil {
			s.logger.Info("Error on creating new account", zap.Error(err))
			serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
			break
		}

		var id int
		err = s.server.db.QueryRow("SELECT id FROM users WHERE username = $1", reqUsername).Scan(&id)
		if err != nil {
			s.logger.Info("Error on querying account id", zap.Error(err))
			serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
			break
		}

		serverRespBytes = s.makeSignInResp(id)
		break
	case err != nil:
		serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
		s.logger.Warn("Got error on SQL query", zap.Error(err))
		break
	default:
		if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqPassword)) == nil {
			s.logger.Info("Passwords match!")
			serverRespBytes = s.makeSignInResp(id)
		} else {
			s.logger.Info("Passwords don't match!")
			serverRespBytes = makeSignInFailureResp(SIGN_EPASS)
		}

	}

	err = s.cryptConn.SendPacket(serverRespBytes)
	if err != nil {
		return err
	}

	return nil
}
