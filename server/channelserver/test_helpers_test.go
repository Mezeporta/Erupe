package channelserver

import (
	"net"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"

	"go.uber.org/zap"
)

// mockPacket implements mhfpacket.MHFPacket for testing.
// Imported from v9.2.x-stable.
type mockPacket struct {
	opcode uint16
}

func (m *mockPacket) Opcode() network.PacketID {
	return network.PacketID(m.opcode)
}

func (m *mockPacket) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	if ctx == nil {
		panic("clientContext is nil")
	}
	bf.WriteUint32(0x12345678)
	return nil
}

func (m *mockPacket) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return nil
}

// createMockServer creates a minimal Server for testing.
// Imported from v9.2.x-stable and adapted for main.
func createMockServer() *Server {
	logger, _ := zap.NewDevelopment()
	s := &Server{
		logger:       logger,
		erupeConfig:  &cfg.Config{},
		// stages is a StageMap (zero value is ready to use)
		sessions:     make(map[net.Conn]*Session),
		handlerTable: buildHandlerTable(),
		raviente: &Raviente{
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
	}
	s.i18n = getLangStrings(s)
	s.Registry = NewLocalChannelRegistry([]*Server{s})
	return s
}

// createMockSession creates a minimal Session for testing.
// Imported from v9.2.x-stable and adapted for main.
func createMockSession(charID uint32, server *Server) *Session {
	logger, _ := zap.NewDevelopment()
	return &Session{
		charID:        charID,
		clientContext: &clientctx.ClientContext{},
		sendPackets:   make(chan packet, 20),
		Name:          "TestPlayer",
		server:        server,
		logger:        logger,
		semaphoreID:   make([]uint16, 2),
	}
}
