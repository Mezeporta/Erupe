package channelserver

import (
	"encoding/binary"
	"encoding/hex"
	"erupe-ce/common/mhfcourse"
	_config "erupe-ce/config"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringstack"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

type packet struct {
	data        []byte
	nonBlocking bool
}

// Session holds state for the channel server connection.
type Session struct {
	sync.Mutex
	logger        *zap.Logger
	server        *Server
	rawConn       net.Conn
	cryptConn     *network.CryptConn
	sendPackets   chan packet
	clientContext *clientctx.ClientContext
	lastPacket    time.Time

	objectIndex      uint16
	userEnteredStage bool // If the user has entered a stage before
	stage            *Stage
	reservationStage *Stage // Required for the stateful MsgSysUnreserveStage packet.
	stagePass        string // Temporary storage
	prevGuildID      uint32 // Stores the last GuildID used in InfoGuild
	charID           uint32
	logKey           []byte
	sessionStart     int64
	courses          []mhfcourse.Course
	token            string
	kqf              []byte
	kqfOverride      bool

	playtime     uint32
	playtimeTime time.Time

	semaphore     *Semaphore // Required for the stateful MsgSysUnreserveStage packet.
	semaphoreMode bool
	semaphoreID   []uint16

	// A stack containing the stage movement history (push on enter/move, pop on back)
	stageMoveStack *stringstack.StringStack

	// Accumulated index used for identifying mail for a client
	// I'm not certain why this is used, but since the client is sending it
	// I want to rely on it for now as it might be important later.
	mailAccIndex uint8
	// Contains the mail list that maps accumulated indexes to mail IDs
	mailList []int

	// For Debuging
	Name     string
	closed   bool
	ackStart map[uint32]time.Time
}

// NewSession creates a new Session type.
func NewSession(server *Server, conn net.Conn) *Session {
	s := &Session{
		logger:         server.logger.Named(conn.RemoteAddr().String()),
		server:         server,
		rawConn:        conn,
		cryptConn:      network.NewCryptConn(conn),
		sendPackets:    make(chan packet, 20),
		clientContext:  &clientctx.ClientContext{}, // Unused
		lastPacket:     time.Now(),
		sessionStart:   TimeAdjusted().Unix(),
		stageMoveStack: stringstack.New(),
		ackStart:       make(map[uint32]time.Time),
		semaphoreID:    make([]uint16, 2),
	}
	s.SetObjectID()
	return s
}

// Start starts the session packet send and recv loop(s).
func (s *Session) Start() {
	s.logger.Debug("New connection", zap.String("RemoteAddr", s.rawConn.RemoteAddr().String()))
	// Unlike the sign and entrance server,
	// the client DOES NOT initalize the channel connection with 8 NULL bytes.
	go s.sendLoop()
	go s.recvLoop()
}

// QueueSend queues a packet (raw []byte) to be sent.
func (s *Session) QueueSend(data []byte) {
	s.logMessage(binary.BigEndian.Uint16(data[0:2]), data, "Server", s.Name)
	err := s.cryptConn.SendPacket(append(data, []byte{0x00, 0x10}...))
	if err != nil {
		s.logger.Warn("Failed to send packet")
	}
}

// QueueSendNonBlocking queues a packet (raw []byte) to be sent, dropping the packet entirely if the queue is full.
func (s *Session) QueueSendNonBlocking(data []byte) {
	select {
	case s.sendPackets <- packet{data, true}:
		s.logMessage(binary.BigEndian.Uint16(data[0:2]), data, "Server", s.Name)
	default:
		s.logger.Warn("Packet queue too full, dropping!")
	}
}

// QueueSendMHF queues a MHFPacket to be sent.
func (s *Session) QueueSendMHF(pkt mhfpacket.MHFPacket) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf, s.clientContext)

	// Queue it.
	s.QueueSend(bf.Data())
}

// QueueSendMHFNonBlocking queues a MHFPacket to be sent, dropping the packet entirely if the queue is full.
func (s *Session) QueueSendMHFNonBlocking(pkt mhfpacket.MHFPacket) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf, s.clientContext)

	// Queue it.
	s.QueueSendNonBlocking(bf.Data())
}

// QueueAck is a helper function to queue an MSG_SYS_ACK with the given ack handle and data.
func (s *Session) QueueAck(ackHandle uint32, data []byte) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(network.MSG_SYS_ACK))
	bf.WriteUint32(ackHandle)
	bf.WriteBytes(data)
	s.QueueSend(bf.Data())
}

func (s *Session) sendLoop() {
	var pkt packet
	for {
		var buf []byte
		if s.closed {
			return
		}
		for len(s.sendPackets) > 0 {
			pkt = <-s.sendPackets
			buf = append(buf, pkt.data...)
		}
		if len(buf) > 0 {
			err := s.cryptConn.SendPacket(append(buf, []byte{0x00, 0x10}...))
			if err != nil {
				s.logger.Warn("Failed to send packet")
			}
		}
		time.Sleep(time.Duration(_config.ErupeConfig.LoopDelay) * time.Millisecond)
	}
}

func (s *Session) recvLoop() {
	for {
		if s.closed {
			logoutPlayer(s)
			return
		}
		pkt, err := s.cryptConn.ReadPacket()
		if err == io.EOF {
			s.logger.Info(fmt.Sprintf("[%s] Disconnected", s.Name))
			logoutPlayer(s)
			return
		} else if err != nil {
			s.logger.Warn("Error on ReadPacket, exiting recv loop", zap.Error(err))
			logoutPlayer(s)
			return
		}
		s.handlePacketGroup(pkt)
		time.Sleep(time.Duration(_config.ErupeConfig.LoopDelay) * time.Millisecond)
	}
}

func (s *Session) handlePacketGroup(pktGroup []byte) {
	s.lastPacket = time.Now()
	bf := byteframe.NewByteFrameFromBytes(pktGroup)
	opcodeUint16 := bf.ReadUint16()
	if len(bf.Data()) >= 6 {
		s.ackStart[bf.ReadUint32()] = time.Now()
		bf.Seek(2, io.SeekStart)
	}
	opcode := network.PacketID(opcodeUint16)

	// This shouldn't be needed, but it's better to recover and let the connection die than to panic the server.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s]", s.Name)
			fmt.Println("Recovered from panic", r)
		}
	}()

	s.logMessage(opcodeUint16, pktGroup, s.Name, "Server")

	if opcode == network.MSG_SYS_LOGOUT {
		s.closed = true
		return
	}
	// Get the packet parser and handler for this opcode.
	mhfPkt := mhfpacket.FromOpcode(opcode)
	if mhfPkt == nil {
		fmt.Println("Got opcode which we don't know how to parse, can't parse anymore for this group")
		return
	}
	// Parse the packet.
	err := mhfPkt.Parse(bf, s.clientContext)
	if err != nil {
		fmt.Printf("\n!!! [%s] %s NOT IMPLEMENTED !!! \n\n\n", s.Name, opcode)
		return
	}
	// Handle the packet.
	handlerTable[opcode](s, mhfPkt)
	// If there is more data on the stream that the .Parse method didn't read, then read another packet off it.
	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 {
		s.handlePacketGroup(remainingData)
	}
}

func ignored(opcode network.PacketID) bool {
	ignoreList := []network.PacketID{
		network.MSG_SYS_END,
		network.MSG_SYS_PING,
		network.MSG_SYS_NOP,
		network.MSG_SYS_TIME,
		network.MSG_SYS_EXTEND_THRESHOLD,
		network.MSG_SYS_POSITION_OBJECT,
		network.MSG_MHF_SAVEDATA,
	}
	set := make(map[network.PacketID]struct{}, len(ignoreList))
	for _, s := range ignoreList {
		set[s] = struct{}{}
	}
	_, r := set[opcode]
	return r
}

func (s *Session) logMessage(opcode uint16, data []byte, sender string, recipient string) {
	if sender == "Server" && !s.server.erupeConfig.DebugOptions.LogOutboundMessages {
		return
	} else if sender != "Server" && !s.server.erupeConfig.DebugOptions.LogInboundMessages {
		return
	}

	opcodePID := network.PacketID(opcode)
	if ignored(opcodePID) {
		return
	}
	var ackHandle uint32
	if len(data) >= 6 {
		ackHandle = binary.BigEndian.Uint32(data[2:6])
	}
	if t, ok := s.ackStart[ackHandle]; ok {
		fmt.Printf("[%s] -> [%s] (%fs)\n", sender, recipient, float64(time.Now().UnixNano()-t.UnixNano())/1000000000)
	} else {
		fmt.Printf("[%s] -> [%s]\n", sender, recipient)
	}
	fmt.Printf("Opcode: (Dec: %d Hex: 0x%04X Name: %s) \n", opcode, opcode, opcodePID)
	if s.server.erupeConfig.DebugOptions.LogMessageData {
		if len(data) <= s.server.erupeConfig.DebugOptions.MaxHexdumpLength {
			fmt.Printf("Data [%d bytes]:\n%s\n", len(data), hex.Dump(data))
		} else {
			fmt.Printf("Data [%d bytes]: (Too long!)\n\n", len(data))
		}
	} else {
		fmt.Printf("\n")
	}
}

func (s *Session) SetObjectID() {
	for i := uint16(1); i < 127; i++ {
		exists := false
		for _, j := range s.server.objectIDs {
			if i == j {
				exists = true
				break
			}
		}
		if !exists {
			s.server.objectIDs[s] = i
			return
		}
	}
	s.server.objectIDs[s] = 0
}

func (s *Session) NextObjectID() uint32 {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(s.server.objectIDs[s])
	s.objectIndex++
	bf.WriteUint16(s.objectIndex)
	bf.Seek(0, 0)
	return bf.ReadUint32()
}

func (s *Session) GetSemaphoreID() uint32 {
	if s.semaphoreMode {
		return 0x000E0000 + uint32(s.semaphoreID[1])
	} else {
		return 0x000F0000 + uint32(s.semaphoreID[0])
	}
}

func (s *Session) isOp() bool {
	var op bool
	err := s.server.db.QueryRow(`SELECT op FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.charID).Scan(&op)
	if err == nil && op {
		return true
	}
	return false
}
