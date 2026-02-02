package entranceserver

import (
	"testing"

	"erupe-ce/config"
)

func TestNewServer(t *testing.T) {
	cfg := &Config{
		Logger:      nil,
		DB:          nil,
		ErupeConfig: &config.Config{},
	}

	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.isShuttingDown {
		t.Error("New server should not be shutting down")
	}
	if s.erupeConfig == nil {
		t.Error("erupeConfig should not be nil")
	}
}

func TestNewServerWithNilConfig(t *testing.T) {
	cfg := &Config{}
	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil for empty config")
	}
}

func TestServerType(t *testing.T) {
	s := &Server{}
	if s.isShuttingDown {
		t.Error("Zero value server should not be shutting down")
	}
	if s.listener != nil {
		t.Error("Zero value server should have nil listener")
	}
}

func TestConfigFields(t *testing.T) {
	cfg := &Config{
		Logger:      nil,
		DB:          nil,
		ErupeConfig: nil,
	}

	if cfg.Logger != nil {
		t.Error("Config Logger should be nil")
	}
	if cfg.DB != nil {
		t.Error("Config DB should be nil")
	}
	if cfg.ErupeConfig != nil {
		t.Error("Config ErupeConfig should be nil")
	}
}
