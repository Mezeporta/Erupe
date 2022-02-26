package network

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/Solenataris/Erupe/network/crypto"
)

// CryptConn represents a MHF encrypted two-way connection,
// it automatically handles encryption, decryption, and key rotation via it's methods.
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
	cc := &CryptConn{
		conn:       conn,
		readKeyRot: 995117,
		sendKeyRot: 995117,
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
	encryptedPacketBody := make([]byte, cph.DataSize)
	_, err = io.ReadFull(cc.conn, encryptedPacketBody)
	if err != nil {
		return nil, err
	}

	// Update the key rotation before decrypting.
	if cph.KeyRotDelta != 0 {
		cc.readKeyRot = (uint32(cph.KeyRotDelta) * (cc.readKeyRot + 1))
	}

	out, combinedCheck, check0, check1, check2 := crypto.Decrypt(encryptedPacketBody, cc.readKeyRot, nil)
	if cph.Check0 != check0 || cph.Check1 != check1 || cph.Check2 != check2 {
		fmt.Printf("got c0 %X, c1 %X, c2 %X\n", check0, check1, check2)
		fmt.Printf("want c0 %X, c1 %X, c2 %X\n", cph.Check0, cph.Check1, cph.Check2)
		fmt.Printf("headerData:\n%s\n", hex.Dump(headerData))
		fmt.Printf("encryptedPacketBody:\n%s\n", hex.Dump(encryptedPacketBody))

		// Attempt to bruteforce it.
		fmt.Println("Crypto out of sync? Attempting bruteforce")
		for key := byte(0); key < 255; key++ {
			out, combinedCheck, check0, check1, check2 = crypto.Decrypt(encryptedPacketBody, 0, &key)
			//fmt.Printf("Key: 0x%X\n%s\n", key, hex.Dump(out))
			if cph.Check0 == check0 && cph.Check1 == check1 && cph.Check2 == check2 {
				fmt.Printf("Bruceforce successful, override key: 0x%X\n", key)

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
		cc.sendKeyRot = (uint32(keyRotDelta) * (cc.sendKeyRot + 1))
	}

	// Encrypt the data
	encData, combinedCheck, check0, check1, check2 := crypto.Encrypt(data, cc.sendKeyRot, nil)

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

	cc.conn.Write(headerBytes)
	cc.conn.Write(encData)

	cc.sentPackets++
	cc.prevSendPacketCombinedCheck = combinedCheck

	return nil
}
