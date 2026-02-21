package network

import (
	"encoding/hex"
	"errors"
	cfg "erupe-ce/config"
	"erupe-ce/network/crypto"
	"io"
	"net"

	"go.uber.org/zap"
)

// Conn defines the interface for a packet-based connection.
// This interface allows for mocking of connections in tests.
type Conn interface {
	// ReadPacket reads and decrypts a packet from the connection
	ReadPacket() ([]byte, error)

	// SendPacket encrypts and sends a packet on the connection
	SendPacket(data []byte) error
}

// CryptConn represents a MHF encrypted two-way connection,
// it automatically handles encryption, decryption, and key rotation via it's methods.
type CryptConn struct {
	logger                      *zap.Logger
	conn                        net.Conn
	realClientMode              cfg.Mode
	readKeyRot                  uint32
	sendKeyRot                  uint32
	sentPackets                 int32
	prevRecvPacketCombinedCheck uint16
	prevSendPacketCombinedCheck uint16
}

// NewCryptConn creates a new CryptConn with proper default values.
func NewCryptConn(conn net.Conn, mode cfg.Mode, logger *zap.Logger) *CryptConn {
	if logger == nil {
		logger = zap.NewNop()
	}
	cc := &CryptConn{
		logger:         logger,
		conn:           conn,
		realClientMode: mode,
		readKeyRot:     995117,
		sendKeyRot:     995117,
	}
	return cc
}

// ReadPacket reads an packet from the connection and returns the decrypted data.
func (cc *CryptConn) ReadPacket() ([]byte, error) {

	// Read the raw 14 byte header.
	headerData := make([]byte, CryptPacketHeaderLength)
	_, err := io.ReadFull(cc.conn, headerData)
	if err != nil {
		return nil, err
	}

	// Parse the data into a usable struct.
	cph, err := NewCryptPacketHeader(headerData)
	if err != nil {
		return nil, err
	}

	// Now read the encrypted packet body after getting its size from the header.
	var encryptedPacketBody []byte

	// Don't know when support for this was added, works in Forward.4, doesn't work in Season 6.0
	if cc.realClientMode < cfg.F1 {
		encryptedPacketBody = make([]byte, cph.DataSize)
	} else {
		encryptedPacketBody = make([]byte, uint32(cph.DataSize)+(uint32(cph.Pf0-0x03)*0x1000))
	}
	_, err = io.ReadFull(cc.conn, encryptedPacketBody)
	if err != nil {
		return nil, err
	}

	// Update the key rotation before decrypting.
	if cph.KeyRotDelta != 0 {
		cc.readKeyRot = uint32(cph.KeyRotDelta) * (cc.readKeyRot + 1)
	}

	out, combinedCheck, check0, check1, check2 := crypto.Crypto(encryptedPacketBody, cc.readKeyRot, false, nil)
	if cph.Check0 != check0 || cph.Check1 != check1 || cph.Check2 != check2 {
		cc.logger.Warn("Crypto checksum mismatch",
			zap.String("got", hex.EncodeToString([]byte{byte(check0 >> 8), byte(check0), byte(check1 >> 8), byte(check1), byte(check2 >> 8), byte(check2)})),
			zap.String("want", hex.EncodeToString([]byte{byte(cph.Check0 >> 8), byte(cph.Check0), byte(cph.Check1 >> 8), byte(cph.Check1), byte(cph.Check2 >> 8), byte(cph.Check2)})),
			zap.String("headerData", hex.Dump(headerData)),
			zap.String("encryptedPacketBody", hex.Dump(encryptedPacketBody)),
		)

		// Attempt to bruteforce it.
		cc.logger.Warn("Crypto out of sync, attempting bruteforce")
		for key := byte(0); key < 255; key++ {
			out, combinedCheck, check0, check1, check2 = crypto.Crypto(encryptedPacketBody, 0, false, &key)
			if cph.Check0 == check0 && cph.Check1 == check1 && cph.Check2 == check2 {
				cc.logger.Info("Bruteforce successful", zap.Uint8("overrideKey", key))

				// Try to fix key for subsequent packets?
				//cc.readKeyRot = (uint32(key) << 1) + 999983

				cc.prevRecvPacketCombinedCheck = combinedCheck
				return out, nil
			}
		}

		return nil, errors.New("decrypted data checksum doesn't match header")
	}

	cc.prevRecvPacketCombinedCheck = combinedCheck
	return out, nil
}

// SendPacket encrypts and sends a packet.
func (cc *CryptConn) SendPacket(data []byte) error {
	keyRotDelta := byte(3)

	if keyRotDelta != 0 {
		cc.sendKeyRot = uint32(keyRotDelta) * (cc.sendKeyRot + 1)
	}

	// Encrypt the data
	encData, combinedCheck, check0, check1, check2 := crypto.Crypto(data, cc.sendKeyRot, true, nil)

	header := &CryptPacketHeader{}
	header.Pf0 = byte(((uint(len(encData)) >> 12) & 0xF3) | 3)
	header.KeyRotDelta = keyRotDelta
	header.PacketNum = uint16(cc.sentPackets)
	header.DataSize = uint16(len(encData))
	header.PrevPacketCombinedCheck = cc.prevSendPacketCombinedCheck
	header.Check0 = check0
	header.Check1 = check1
	header.Check2 = check2

	headerBytes, err := header.Encode()
	if err != nil {
		return err
	}

	_, err = cc.conn.Write(append(headerBytes, encData...))
	if err != nil {
		return err
	}
	cc.sentPackets++
	cc.prevSendPacketCombinedCheck = combinedCheck

	return nil
}
