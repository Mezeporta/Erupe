package channelserver

import (
	"encoding/binary"
	"testing"

	_config "erupe-ce/config"
)

func TestBackportQuest_Basic(t *testing.T) {
	// Set up config for the test
	oldConfig := _config.ErupeConfig
	defer func() { _config.ErupeConfig = oldConfig }()

	_config.ErupeConfig = &_config.Config{}
	_config.ErupeConfig.RealClientMode = _config.ZZ

	// Create a quest data buffer large enough for BackportQuest to work with.
	// The function reads a uint32 from data[0:4] as offset, then works at offset+96.
	// We need at least offset + 96 + 108 + 6*8 bytes.
	// Set offset (wp base) = 0, so wp starts at 96, rp at 100.
	data := make([]byte, 512)
	binary.LittleEndian.PutUint32(data[0:4], 0) // offset = 0

	// Fill some data at the rp positions so we can verify copies
	for i := 100; i < 400; i++ {
		data[i] = byte(i & 0xFF)
	}

	result := BackportQuest(data)
	if result == nil {
		t.Fatal("BackportQuest returned nil")
	}
	if len(result) != len(data) {
		t.Errorf("BackportQuest changed data length: got %d, want %d", len(result), len(data))
	}
}

func TestBackportQuest_S6Mode(t *testing.T) {
	oldConfig := _config.ErupeConfig
	defer func() { _config.ErupeConfig = oldConfig }()

	_config.ErupeConfig = &_config.Config{}
	_config.ErupeConfig.RealClientMode = _config.S6

	data := make([]byte, 512)
	binary.LittleEndian.PutUint32(data[0:4], 0)

	for i := 0; i < len(data); i++ {
		data[i+4] = byte(i % 256)
		if i+4 >= len(data)-1 {
			break
		}
	}

	// Set some values at data[8:12] so we can check they get copied to data[16:20]
	binary.LittleEndian.PutUint32(data[8:12], 0xDEADBEEF)

	result := BackportQuest(data)
	if result == nil {
		t.Fatal("BackportQuest returned nil")
	}

	// In S6 mode, data[16:20] should be copied from data[8:12]
	got := binary.LittleEndian.Uint32(result[16:20])
	if got != 0xDEADBEEF {
		t.Errorf("S6 mode: data[16:20] = 0x%X, want 0xDEADBEEF", got)
	}
}

func TestBackportQuest_G91Mode_PatternReplacement(t *testing.T) {
	oldConfig := _config.ErupeConfig
	defer func() { _config.ErupeConfig = oldConfig }()

	_config.ErupeConfig = &_config.Config{}
	_config.ErupeConfig.RealClientMode = _config.G91

	data := make([]byte, 512)
	binary.LittleEndian.PutUint32(data[0:4], 0)

	// Insert an armor sphere pattern at a known location
	// Pattern: 0x0A, 0x00, 0x01, 0x33 -> should replace bytes at +2 with 0xD7, 0x00
	offset := 300
	data[offset] = 0x0A
	data[offset+1] = 0x00
	data[offset+2] = 0x01
	data[offset+3] = 0x33

	result := BackportQuest(data)

	// After BackportQuest, the pattern's last 2 bytes should be replaced
	if result[offset+2] != 0xD7 || result[offset+3] != 0x00 {
		t.Errorf("G91 pattern replacement failed: got [0x%X, 0x%X], want [0xD7, 0x00]",
			result[offset+2], result[offset+3])
	}
}

func TestBackportQuest_F5Mode(t *testing.T) {
	oldConfig := _config.ErupeConfig
	defer func() { _config.ErupeConfig = oldConfig }()

	_config.ErupeConfig = &_config.Config{}
	_config.ErupeConfig.RealClientMode = _config.F5

	data := make([]byte, 512)
	binary.LittleEndian.PutUint32(data[0:4], 0)

	result := BackportQuest(data)
	if result == nil {
		t.Fatal("BackportQuest returned nil")
	}
}

func TestBackportQuest_G101Mode(t *testing.T) {
	oldConfig := _config.ErupeConfig
	defer func() { _config.ErupeConfig = oldConfig }()

	_config.ErupeConfig = &_config.Config{}
	_config.ErupeConfig.RealClientMode = _config.G101

	data := make([]byte, 512)
	binary.LittleEndian.PutUint32(data[0:4], 0)

	result := BackportQuest(data)
	if result == nil {
		t.Fatal("BackportQuest returned nil")
	}
}
