package protocol

import (
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"erupe-ce/cmd/protbot/conn"
)

// PacketHandler is a callback invoked when a server-pushed packet is received.
type PacketHandler func(opcode uint16, data []byte)

// ChannelConn manages a connection to a channel server.
type ChannelConn struct {
	conn       *conn.MHFConn
	ackCounter uint32
	waiters    sync.Map // map[uint32]chan *AckResponse
	handlers   sync.Map // map[uint16]PacketHandler
	closed     atomic.Bool
}

// OnPacket registers a handler for a specific server-pushed opcode.
// Only one handler per opcode; later registrations replace earlier ones.
func (ch *ChannelConn) OnPacket(opcode uint16, handler PacketHandler) {
	ch.handlers.Store(opcode, handler)
}

// AckResponse holds the parsed ACK data from the server.
type AckResponse struct {
	AckHandle        uint32
	IsBufferResponse bool
	ErrorCode        uint8
	Data             []byte
}

// ConnectChannel establishes a connection to a channel server.
// Channel servers do NOT use the 8 NULL byte initialization.
func ConnectChannel(addr string) (*ChannelConn, error) {
	c, err := conn.DialDirect(addr)
	if err != nil {
		return nil, fmt.Errorf("channel connect: %w", err)
	}

	ch := &ChannelConn{
		conn: c,
	}

	go ch.recvLoop()
	return ch, nil
}

// NextAckHandle returns the next unique ACK handle for packet requests.
func (ch *ChannelConn) NextAckHandle() uint32 {
	return atomic.AddUint32(&ch.ackCounter, 1)
}

// SendPacket encrypts and sends raw packet data (including the 0x00 0x10 terminator
// which is already appended by the Build* functions in packets.go).
func (ch *ChannelConn) SendPacket(data []byte) error {
	return ch.conn.SendPacket(data)
}

// WaitForAck waits for an ACK response matching the given handle.
func (ch *ChannelConn) WaitForAck(handle uint32, timeout time.Duration) (*AckResponse, error) {
	waitCh := make(chan *AckResponse, 1)
	ch.waiters.Store(handle, waitCh)
	defer ch.waiters.Delete(handle)

	select {
	case resp := <-waitCh:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("ACK timeout for handle %d", handle)
	}
}

// Close closes the channel connection.
func (ch *ChannelConn) Close() error {
	ch.closed.Store(true)
	return ch.conn.Close()
}

// recvLoop continuously reads packets from the channel server and dispatches ACKs.
func (ch *ChannelConn) recvLoop() {
	for {
		if ch.closed.Load() {
			return
		}

		pkt, err := ch.conn.ReadPacket()
		if err != nil {
			if ch.closed.Load() {
				return
			}
			fmt.Printf("[channel] read error: %v\n", err)
			return
		}

		if len(pkt) < 2 {
			continue
		}

		// Strip trailing 0x00 0x10 terminator if present for opcode parsing.
		// Packets from server: [opcode uint16][fields...][0x00 0x10]
		opcode := binary.BigEndian.Uint16(pkt[0:2])

		switch opcode {
		case MSG_SYS_ACK:
			ch.handleAck(pkt[2:])
		case MSG_SYS_PING:
			ch.handlePing(pkt[2:])
		default:
			if val, ok := ch.handlers.Load(opcode); ok {
				val.(PacketHandler)(opcode, pkt[2:])
			} else {
				fmt.Printf("[channel] recv opcode 0x%04X (%d bytes)\n", opcode, len(pkt))
			}
		}
	}
}

// handleAck parses an ACK packet and dispatches it to a waiting caller.
// Reference: Erupe network/mhfpacket/msg_sys_ack.go
func (ch *ChannelConn) handleAck(data []byte) {
	if len(data) < 8 {
		return
	}

	ackHandle := binary.BigEndian.Uint32(data[0:4])
	isBuffer := data[4] > 0
	errorCode := data[5]

	var ackData []byte
	if isBuffer {
		payloadSize := binary.BigEndian.Uint16(data[6:8])
		offset := uint32(8)
		if payloadSize == 0xFFFF {
			if len(data) < 12 {
				return
			}
			payloadSize32 := binary.BigEndian.Uint32(data[8:12])
			offset = 12
			if uint32(len(data)) >= offset+payloadSize32 {
				ackData = data[offset : offset+payloadSize32]
			}
		} else {
			if uint32(len(data)) >= offset+uint32(payloadSize) {
				ackData = data[offset : offset+uint32(payloadSize)]
			}
		}
	} else {
		// Simple ACK: 4 bytes of data after the uint16 field.
		if len(data) >= 12 {
			ackData = data[8:12]
		}
	}

	resp := &AckResponse{
		AckHandle:        ackHandle,
		IsBufferResponse: isBuffer,
		ErrorCode:        errorCode,
		Data:             ackData,
	}

	if val, ok := ch.waiters.Load(ackHandle); ok {
		waitCh := val.(chan *AckResponse)
		select {
		case waitCh <- resp:
		default:
		}
	} else {
		fmt.Printf("[channel] unexpected ACK handle %d (error=%d, buffer=%v, %d bytes)\n",
			ackHandle, errorCode, isBuffer, len(ackData))
	}
}

// handlePing responds to a server ping to keep the connection alive.
func (ch *ChannelConn) handlePing(data []byte) {
	var ackHandle uint32
	if len(data) >= 4 {
		ackHandle = binary.BigEndian.Uint32(data[0:4])
	}
	pkt := BuildPingPacket(ackHandle)
	if err := ch.conn.SendPacket(pkt); err != nil {
		fmt.Printf("[channel] ping response failed: %v\n", err)
	}
}
