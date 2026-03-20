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
		logger:      logger,
		erupeConfig: &cfg.Config{},
		stages:      make(map[string]*Stage),
		sessions:    make(map[net.Conn]*Session),
		raviente: &Raviente{
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
	}
	s.i18n = getLangStrings(s)
	s.Registry = NewLocalChannelRegistry([]*Server{s})
	// GuildService is wired lazily by tests that set repos then call ensureGuildService.
	return s
}

// ensureMailService wires the MailService from the server's current repos.
// Call this after setting mailRepo and guildRepo on the mock server.
func ensureMailService(s *Server) {
	s.mailService = NewMailService(s.mailRepo, s.guildRepo, s.logger)
}

// ensureGuildService wires the GuildService from the server's current repos.
// Call this after setting guildRepo, mailRepo, and charRepo on the mock server.
func ensureGuildService(s *Server) {
	ensureMailService(s)
	s.guildService = NewGuildService(s.guildRepo, s.mailService, s.charRepo, s.logger)
}

// ensureAchievementService wires the AchievementService from the server's current repos.
func ensureAchievementService(s *Server) {
	s.achievementService = NewAchievementService(s.achievementRepo, s.logger)
}

// ensureGachaService wires the GachaService from the server's current repos.
func ensureGachaService(s *Server) {
	s.gachaService = NewGachaService(s.gachaRepo, s.userRepo, s.charRepo, s.logger, 100000)
}

// ensureTowerService wires the TowerService from the server's current repos.
func ensureTowerService(s *Server) {
	s.towerService = NewTowerService(s.towerRepo, s.logger)
}

// ensureFestaService wires the FestaService from the server's current repos.
func ensureFestaService(s *Server) {
	s.festaService = NewFestaService(s.festaRepo, s.logger)
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
