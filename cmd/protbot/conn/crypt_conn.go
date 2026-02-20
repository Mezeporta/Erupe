package conn

import (
	"encoding/hex"
	"errors"
	"erupe-ce/network/crypto"
	"fmt"
	"io"
	"net"
)

// CryptConn is an MHF encrypted two-way connection.
// Adapted from Erupe's network/crypt_conn.go with config dependency removed.
// Hardcoded to ZZ mode (supports Pf0-based extended data size).
type CryptConn struct {
	conn                        net.Conn
	readKeyRot                  uint32
	sendKeyRot                  uint32
	sentPackets                 int32
	prevRecvPacketCombinedCheck uint16
	prevSendPacketCombinedCheck uint16
}

// NewCryptConn creates a new CryptConn with proper default values.
func NewCryptConn(conn net.Conn) *CryptConn {
	return &CryptConn{
		conn:       conn,
		readKeyRot: 995117,
		sendKeyRot: 995117,
	}
}

// ReadPacket reads a packet from the connection and returns the decrypted data.
func (cc *CryptConn) ReadPacket() ([]byte, error) {
	headerData := make([]byte, CryptPacketHeaderLength)
	_, err := io.ReadFull(cc.conn, headerData)
	if err != nil {
		return nil, err
	}

	cph, err := NewCryptPacketHeader(headerData)
	if err != nil {
		return nil, err
	}

	// ZZ mode: extended data size using Pf0 field.
	encryptedPacketBody := make([]byte, uint32(cph.DataSize)+(uint32(cph.Pf0-0x03)*0x1000))
	_, err = io.ReadFull(cc.conn, encryptedPacketBody)
	if err != nil {
		return nil, err
	}

	if cph.KeyRotDelta != 0 {
		cc.readKeyRot = uint32(cph.KeyRotDelta) * (cc.readKeyRot + 1)
	}

	out, combinedCheck, check0, check1, check2 := crypto.Crypto(encryptedPacketBody, cc.readKeyRot, false, nil)
	if cph.Check0 != check0 || cph.Check1 != check1 || cph.Check2 != check2 {
		fmt.Printf("got c0 %X, c1 %X, c2 %X\n", check0, check1, check2)
		fmt.Printf("want c0 %X, c1 %X, c2 %X\n", cph.Check0, cph.Check1, cph.Check2)
		fmt.Printf("headerData:\n%s\n", hex.Dump(headerData))
		fmt.Printf("encryptedPacketBody:\n%s\n", hex.Dump(encryptedPacketBody))

		// Attempt bruteforce recovery.
		fmt.Println("Crypto out of sync? Attempting bruteforce")
		for key := byte(0); key < 255; key++ {
			out, combinedCheck, check0, check1, check2 = crypto.Crypto(encryptedPacketBody, 0, false, &key)
			if cph.Check0 == check0 && cph.Check1 == check1 && cph.Check2 == check2 {
				fmt.Printf("Bruteforce successful, override key: 0x%X\n", key)
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
