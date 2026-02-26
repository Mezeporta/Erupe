package signserver

import (
	"fmt"
	"strings"
	"testing"
	"time"

	cfg "erupe-ce/config"
	"go.uber.org/zap"
)

func TestMakeSignResponse_PS3Client(t *testing.T) {
	config := &cfg.Config{
		PatchServerFile:     "http://patch.example.com/file",
		PatchServerManifest: "http://patch.example.com/manifest",
		DebugOptions: cfg.DebugOptions{
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
		},
	}

	server := newMakeSignResponseServer(config)
	server.charRepo = &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, Name: "TestHunter", HR: 100, GR: 50, WeaponType: 3, LastLogin: 1700000000},
		},
	}

	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PS3,
	}

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Error("makeSignResponse() returned empty result")
	}
	if result[0] != uint8(SIGN_SUCCESS) {
		t.Errorf("first byte = %d, want %d (SIGN_SUCCESS)", result[0], SIGN_SUCCESS)
	}
}

func TestMakeSignResponse_PS3NoPatchServer(t *testing.T) {
	config := &cfg.Config{
		PatchServerFile:     "",
		PatchServerManifest: "",
		DebugOptions: cfg.DebugOptions{
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
	}

	server := newMakeSignResponseServer(config)
	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PS3,
	}

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Fatal("makeSignResponse() returned empty result")
	}
	if result[0] != uint8(SIGN_EABORT) {
		t.Errorf("first byte = %d, want %d (SIGN_EABORT)", result[0], SIGN_EABORT)
	}
}

func TestMakeSignResponse_HideLoginNotice(t *testing.T) {
	config := &cfg.Config{
		HideLoginNotice: true,
		DebugOptions: cfg.DebugOptions{
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
		},
	}

	server := newMakeSignResponseServer(config)
	server.charRepo = &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, Name: "TestHunter", HR: 50},
		},
	}

	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("array bounds panic: %v", r)
			}
		}
	}()

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Error("makeSignResponse() returned empty result")
	}
}

func TestMakeSignResponse_MaxLauncherHR(t *testing.T) {
	config := &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			MaxLauncherHR: true,
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
		},
	}

	server := newMakeSignResponseServer(config)
	server.charRepo = &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, Name: "TestHunter", HR: 50},
		},
	}

	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("array bounds panic: %v", r)
			}
		}
	}()

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Error("makeSignResponse() returned empty result")
	}
}

func TestMakeSignResponse_FriendsOverflow(t *testing.T) {
	config := &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
		},
	}

	// Create 300 friends (> 255)
	friends := make([]members, 300)
	for i := range friends {
		friends[i] = members{CID: uint32(i + 1), ID: uint32(i + 1000), Name: fmt.Sprintf("Friend%d", i)}
	}

	server := newMakeSignResponseServer(config)
	server.charRepo = &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, Name: "TestHunter", HR: 50},
		},
		friends: friends,
	}

	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("array bounds panic: %v", r)
			}
		}
	}()

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Error("makeSignResponse() returned empty result")
	}
}

func TestMakeSignResponse_GuildmatesOverflow(t *testing.T) {
	config := &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			CapLink: cfg.CapLinkOptions{
				Values: []uint16{0, 0, 0, 0, 0},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
		},
	}

	guildmates := make([]members, 260)
	for i := range guildmates {
		guildmates[i] = members{CID: uint32(i + 1), ID: uint32(i + 1000), Name: fmt.Sprintf("Mate%d", i)}
	}

	server := newMakeSignResponseServer(config)
	server.charRepo = &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, Name: "TestHunter", HR: 50},
		},
		guildmates: guildmates,
	}
	server.userRepo = &mockSignUserRepo{
		returnExpiry: time.Now().Add(time.Hour * 24 * 30),
		lastLogin:    time.Now(),
	}

	conn := newMockConn()
	session := &Session{
		logger:  zap.NewNop(),
		server:  server,
		rawConn: conn,
		client:  PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("array bounds panic: %v", r)
			}
		}
	}()

	result := session.makeSignResponse(1)
	if len(result) == 0 {
		t.Error("makeSignResponse() returned empty result")
	}
}
